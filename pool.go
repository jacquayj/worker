package worker

import (
	"fmt"
	"sync"
)

// Pool is a generic abstraction for a worker pool
type Pool[R any] struct {
	jobs      chan JobFunc[R]
	results   chan jobResult[R]
	opts      PoolOpts
	jobWg     sync.WaitGroup
	resultsWg sync.WaitGroup
	closed    bool

	launchedRoutines int
}

type jobResult[R any] struct {
	res R
	err error
}

// JobFunc is a generic type representing job functions
type JobFunc[R any] func() (R, error)

// ResultFunc is a generic type representing job result callback functions
type ResultFunc[R any] func(R, error) error

// ResultBreakFunc is a generic type representing job result callback functions supporting non-error breaks
type ResultBreakFunc[R any] func(R, error) bool

// NewPool creates a new generic thread pool
func NewPool[R any](opts ...PoolOpts) *Pool[R] {
	mergedOpts := mergePoolOpts(opts)

	var jobs chan JobFunc[R]
	var results chan jobResult[R]

	if mergedOpts.BufferSize != nil {
		jobs = make(chan JobFunc[R], *mergedOpts.BufferSize)
		results = make(chan jobResult[R], *mergedOpts.BufferSize)
	} else {
		jobs = make(chan JobFunc[R])
		results = make(chan jobResult[R])
	}

	pool := &Pool[R]{
		jobs:    jobs,
		results: results,
		opts:    mergedOpts,
	}

	pool.initWorkers()

	return pool
}

// Result invokes the provided callback when results are received, and breaks + returns the callback error if non-nil
func (p *Pool[R]) Result(callback ResultFunc[R]) error {
	for result := range p.results {
		if err := callback(result.res, result.err); err != nil {
			return err
		}
	}
	return nil
}

// ResultBreak invokes the provided callback when results are received, and breaks subsequent callbacks if the callback return boolean is true
func (p *Pool[R]) ResultBreak(callback ResultBreakFunc[R]) {
	for result := range p.results {
		brek := callback(result.res, result.err)
		if brek {
			break
		}
	}
}

// SubmitJob queues the job function for execution
func (p *Pool[R]) SubmitJob(job JobFunc[R]) error {
	if p.closed {
		err := fmt.Errorf("unable to submit job, FinishedJobSubmissions() already called")
		logMsg(*p.opts.LogLevel, Error, err.Error())
		return err
	}

	p.jobWg.Add(1)

	select {
	case p.jobs <- job:
		p.jobWg.Done()
	default:
		// channel is full or blocked, launch new goutine to prevent blocking

		// ensure caller doesn't cause goroutine leak
		if p.launchedRoutines >= *p.opts.MaxJobGorountines {
			err := fmt.Errorf("unable to submit job, number of queued goroutines would excede MaxJobGorountines (%d)", *p.opts.MaxJobGorountines)
			logMsg(*p.opts.LogLevel, Error, err.Error())
			return err
		}

		p.launchedRoutines += 1
		go func() {
			p.jobs <- job
			p.jobWg.Done()
			p.launchedRoutines -= 1
		}()
	}

	return nil
}

// FinishedJobSubmission informs the pool that no more jobs will be submitted. Attempts to submit additional jobs to the pool, after this function is called, will return an error. Will trigger result callback functions to return after job execution completes.
func (p *Pool[R]) FinishedJobSubmission() {
	logMsg(*p.opts.LogLevel, Info, "FinishedJobSubmission called")

	p.closed = true

	go func() {
		// Don't close jobs until all goroutine-queued submissions are sent
		p.jobWg.Wait()
		close(p.jobs)
	}()

	go func() {
		// Don't close results until all goroutine-queued results are sent
		p.resultsWg.Wait()
		close(p.results)
	}()
}

func (p *Pool[R]) initWorkers() {
	for i := 0; i < *p.opts.WorkerCount; i++ {
		workerInx := i

		// Spin up worker goroutine
		go func() {
			for job := range p.jobs {

				p.resultsWg.Add(1)

				// Execute job
				result, err := job()

				select {
				case p.results <- jobResult[R]{result, err}:
					p.resultsWg.Done()
				default:
					// channel is full or blocked, launch new goutine to prevent blocking

					// ensure caller doesn't cause goroutine leak
					if p.launchedRoutines >= *p.opts.MaxJobGorountines {
						stallErr := fmt.Sprintf("worker stalled: unable to send job results, number of queued goroutines would excede MaxJobGorountines (%d)", *p.opts.MaxJobGorountines)
						logMsg(*p.opts.LogLevel, Error, stallErr)

						p.results <- jobResult[R]{
							result, err,
						}
						p.resultsWg.Done()
					} else {
						p.launchedRoutines += 1
						go func() {
							p.results <- jobResult[R]{
								result, err,
							}
							p.resultsWg.Done()
							p.launchedRoutines -= 1
						}()
					}

				}
			}

			logMsg(*p.opts.LogLevel, Info, "Worker %v finished", workerInx)
		}()
	}
}
