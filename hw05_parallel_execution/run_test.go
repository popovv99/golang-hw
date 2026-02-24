package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, runTasksCount, int32(tasksCount), "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})

	t.Run("m zero ignores errors", func(t *testing.T) {
		var executed int32

		tasks := []Task{
			func() error {
				atomic.AddInt32(&executed, 1)
				return errors.New("err")
			},
			func() error {
				atomic.AddInt32(&executed, 1)
				return errors.New("err")
			},
			func() error {
				atomic.AddInt32(&executed, 1)
				return nil
			},
		}

		err := Run(tasks, 2, 0)

		require.NoError(t, err)
		require.Equal(t, int32(3), executed, "all tasks must be executed when m == 0")
	})

	t.Run("no goroutine leaks", func(t *testing.T) {
		tasks := make([]Task, 100)
		for i := 0; i < len(tasks); i++ {
			tasks[i] = func() error {
				return errors.New("fail")
			}
		}

		_ = Run(tasks, 5, 1)
	})

	t.Run("too many tasks started", func(t *testing.T) {
		const (
			tasksCount = 100
			workers    = 5
			maxErrors  = 3
		)

		var started int32

		tasks := make([]Task, tasksCount)
		for i := 0; i < tasksCount; i++ {
			tasks[i] = func() error {
				atomic.AddInt32(&started, 1)
				return errors.New("err")
			}
		}

		err := Run(tasks, workers, maxErrors)

		require.ErrorIs(t, err, ErrErrorsLimitExceeded)
		require.LessOrEqual(
			t,
			int(started),
			workers+maxErrors,
			"too many tasks were started",
		)
	})
}
