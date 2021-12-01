package worker

import (
	"fmt"
	"sync"
)

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

type JobFunc[R any] func() (R, error)

type ResultFunc[R any] func(R, error) error

type ResultBreakFunc[R any] func(R, error) bool

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

func (p *Pool[R]) Result(callback ResultFunc[R]) error {
	for result := range p.results {
		if err := callback(result.res, result.err); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pool[R]) ResultBreak(callback ResultBreakFunc[R]) {
	for result := range p.results {
		brek := callback(result.res, result.err)
		if brek {
			break
		}
	}
}

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

func (p *Pool[R]) FinishedJobSubmission() {
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
		}()
	}
}
