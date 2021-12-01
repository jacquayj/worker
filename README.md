# worker

Worker pool using go generics

1. Compile and install go `master` from source https://golang.org/doc/install/source
2. `/my/go/install/go run -gcflags=-G=3 test.go`

test.go:

```go
package main

import (
    "log"
    "math/rand"

    "github.com/jacquayj/worker"
)

func main() {
    randIntpool := worker.NewPool[int]()

    for i := 0; i < 100; i++ {
        randIntpool.SubmitJob(func() (int, error) {
            return rand.Int(), nil
        })
    }
    randIntpool.FinishedJobSubmission()

    randIntpool.Result(func(result int, err error) error {
        log.Print(result)
        return nil
    })
}

```

Or begin processing results before all jobs are submitted:

```go

func main() {
    randIntpool := worker.NewPool[int]()

    go func() {
        for i := 0; i < 100; i++ {
            randIntpool.SubmitJob(func() (int, error) {
                return rand.Int(), nil
            })
        }
        randIntpool.FinishedJobSubmission()
    }()

    randIntpool.Result(func(result int, err error) error {
        log.Print(result)
        return nil
    })
}
```

Use a struct as the results type:

```go
func main() {

    type RandJobResult struct {
        Inx int
        Num int
    }

    randIntpool := worker.NewPool[RandJobResult]()

    go func() {
        for i := 0; i < 100; i++ {
            inx := i
            randIntpool.SubmitJob(func() (RandJobResult, error) {
                return RandJobResult{inx, rand.Int()}, nil
            })
        }
        randIntpool.FinishedJobSubmission()
    }()

    randIntpool.Result(func(result RandJobResult, err error) error {
        log.Print(result)
        return nil
    })
}
```
