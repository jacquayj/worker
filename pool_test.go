package worker_test

import (
	"log"
	"math/rand"
	"testing"

	"github.com/jacquayj/worker"
)

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

func Test_Worker_Routine(t *testing.T) {
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

func Test_Worker_Struct(t *testing.T) {

	type RandJobResult struct {
		Inx int
		Num int
	}

	randIntpool := worker.NewPool[RandJobResult]()

	for i := 0; i < 100; i++ {
		inx := i
		randIntpool.SubmitJob(func() (RandJobResult, error) {
			return RandJobResult{inx, rand.Int()}, nil
		})
	}
	randIntpool.FinishedJobSubmission()

	randIntpool.Result(func(result RandJobResult, err error) error {

		log.Print(result)

		return nil
	})
}
