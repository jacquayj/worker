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

	gotResults := 0
	randIntpool.Result(func(result int, err error) error {
		gotResults++
		log.Print(result)
		return nil
	})

	// assert gotResults == 100
	if gotResults != 100 {
		t.Errorf("Expected 100, got %d", gotResults)
	}
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

	gotResults := 0
	randIntpool.Result(func(result int, err error) error {
		gotResults++
		log.Print(result)
		return nil
	})

	// assert gotResults == 100
	if gotResults != 100 {
		t.Errorf("Expected 100, got %d", gotResults)
	}
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

	gotResults := 0
	randIntpool.Result(func(result RandJobResult, err error) error {
		gotResults++
		log.Print(result)

		return nil
	})

	// assert gotResults == 100
	if gotResults != 100 {
		t.Errorf("Expected 100, got %d", gotResults)
	}
}

func Test_Panic_Single_Job_Race(t *testing.T) {
	logErr := worker.LogLevel(worker.Error)
	randIntpool := worker.NewPool[int](worker.PoolOpts{
		LogLevel: &logErr,
	})

	// you had one job
	randIntpool.SubmitJob(func() (int, error) {
		return rand.Int(), nil
	})
	randIntpool.FinishedJobSubmission()

	total := 0
	randIntpool.Result(func(result int, err error) error {
		total++
		return nil
	})
	log.Print(total)

	// assert total == 1
	if total != 1 {
		t.Errorf("Expected 1, got %d", total)
	}
}
