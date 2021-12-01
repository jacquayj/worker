package worker

import (
	"fmt"
	"runtime"
)

type PoolOpts struct {
	WorkerCount       *int
	BufferSize        *int
	MaxJobGorountines *int
	LogLevel          *LogLevel
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

	defaultMaxJobGoroutines := 512
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

	logMsg(*opts.LogLevel, Info, "configured pool options: %v", opts)

	return
}
