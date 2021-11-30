package worker_test

import (
	"log"
	"math/rand"
	"testing"

	"github.com/jacquayj/worker"
)

type RandJobOpts struct {
	Num int
}

func Test_Worker(t *testing.T) {
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
