package worker

import (
	"fmt"
	"runtime"
)

// PoolOpts contains options for use by the worker pool
type PoolOpts struct {
	// WorkerCount is the desired number of workers to process jobs that are submitted
	WorkerCount *int

	// BufferSize designates the size of buffered channels used for jobs and results
	// if BufferSize is nil, non-buffered channels are utilized
	// this should be set to the maximum number of jobs you expect to be enqueued at any given time
	BufferSize *int

	// MaxJobGorountines is the max number of simultaneous running gouroutines utilized for non-blocking SubmitJob calls, before blocking or an error is returned from SubmitJob
	// gouroutines are only utilized if BufferSize is nil, or if the buffered channels overflow
	MaxJobGorountines *int

	// BlockSubmissions specifies whether to block or return an error if MaxJobGorountines is exceeded
	BlockSubmissions *bool

	// LogLevel designates the logging verbosity of the worker package
	LogLevel *LogLevel
}

func (po PoolOpts) String() string {
	var bufSize string
	if po.BufferSize == nil {
		bufSize = "<nil>"
	} else {
		bufSize = fmt.Sprintf("%v", *po.BufferSize)
	}
	return fmt.Sprintf("PoolOpts{WorkerCount:%v BufferSize:%v MaxJobGorountines:%v LogLevel:%v}", *po.WorkerCount, bufSize, *po.MaxJobGorountines, *po.LogLevel)
}

func mergePoolOpts(inputOpts []PoolOpts) (opts PoolOpts) {
	// Merge into single PoolOpts
	for _, o := range inputOpts {
		if o.BufferSize != nil {
			opts.BufferSize = o.BufferSize
		}
		if o.WorkerCount != nil {
			opts.WorkerCount = o.WorkerCount
		}
		if o.MaxJobGorountines != nil {
			opts.MaxJobGorountines = o.MaxJobGorountines
		}
		if o.LogLevel != nil {
			opts.LogLevel = o.LogLevel
		}
		if o.BlockSubmissions != nil {
			opts.BlockSubmissions = o.BlockSubmissions
		}
	}

	// defaults & validations
	defaultlogLevel := Info
	if opts.LogLevel == nil {
		opts.LogLevel = &defaultlogLevel
	} else if *opts.LogLevel > Info {
		opts.LogLevel = &defaultlogLevel
	}

	defaultWorkerCount := runtime.NumCPU()
	if opts.WorkerCount == nil {
		opts.WorkerCount = &defaultWorkerCount
	} else if *opts.WorkerCount <= 0 {
		logMsg(*opts.LogLevel, Warning, "option WorkerCount out of range, using default: runtime.NumCPU()")
		opts.WorkerCount = &defaultWorkerCount
	}

	defaultMaxJobGoroutines := 1024
	if opts.MaxJobGorountines == nil {
		opts.MaxJobGorountines = &defaultMaxJobGoroutines
	} else if *opts.MaxJobGorountines < 0 {
		logMsg(*opts.LogLevel, Warning, "option MaxJobGorountines out of range, using default: %v", defaultMaxJobGoroutines)
		opts.MaxJobGorountines = &defaultMaxJobGoroutines
	}

	defaultBufferSize := 512
	if opts.BufferSize != nil && *opts.BufferSize < 0 {
		logMsg(*opts.LogLevel, Warning, "option BufferSize out of range and non-nil, using default: %v", defaultBufferSize)
		opts.BufferSize = &defaultBufferSize
	}

	defaultBlockSubmissions := true
	if opts.BlockSubmissions == nil {
		opts.BlockSubmissions = &defaultBlockSubmissions
	}

	logMsg(*opts.LogLevel, Info, "configured pool options: %v", opts)

	return
}
