package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

func Run(tasks []Task, n, m int) error {
	if n <= 0 {
		return nil
	}

	var wg sync.WaitGroup
	var index int64
	var errors int32

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				if m > 0 && atomic.LoadInt32(&errors) >= int32(m) {
					return
				}

				i := atomic.AddInt64(&index, 1) - 1
				if int(i) >= len(tasks) {
					return
				}

				if err := tasks[i](); err != nil && m > 0 {
					atomic.AddInt32(&errors, 1)
				}
			}
		}()
	}

	wg.Wait()

	if m > 0 && atomic.LoadInt32(&errors) >= int32(m) {
		return ErrErrorsLimitExceeded
	}
	return nil
}
