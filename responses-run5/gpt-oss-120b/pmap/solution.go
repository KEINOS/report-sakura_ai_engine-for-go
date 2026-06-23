package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input concurrently, preserving
// the order of results. At most workers calls to fn are active at any time.
// The function respects ctx cancellation, returns the first error (or panic)
// encountered, and ensures no goroutine leaks.
//
// Errors:
//   * workers <= 0 : returns an error.
//   * fn == nil    : returns an error.
//   * empty input : returns a non‑nil empty slice and nil error.
//   * first fn error, panic, or ctx cancellation : returns nil slice and the error.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	// Validate arguments.
	if workers <= 0 {
		return nil, errors.New("workers must be > 0")
	}
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}
	if len(input) == 0 {
		return make([]int, 0), nil
	}

	// Result slice, indexed by the original position.
	results := make([]int, len(input))

	// Context used to cancel workers on first error or parent cancellation.
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		idx int
		val int
	}
	taskCh := make(chan task)

	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	// Helper to record the first error and cancel the context.
	recordError := func(err error) {
		once.Do(func() {
			firstErr = err
			cancel()
		})
	}

	// Worker goroutine.
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-childCtx.Done():
				return
			case t, ok := <-taskCh:
				if !ok {
					return
				}
				// Execute fn with panic recovery.
				func() {
					defer func() {
						if r := recover(); r != nil {
							recordError(fmt.Errorf("panic in fn: %v", r))
						}
					}()
					res, err := fn(childCtx, t.val)
					if err != nil {
						recordError(err)
						return
					}
					results[t.idx] = res
				}()
			}
		}
	}

	// Launch workers.
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	// Feed tasks.
	go func() {
		defer close(taskCh)
		for i, v := range input {
			select {
			case <-childCtx.Done():
				return
			case taskCh <- task{idx: i, val: v}:
			}
		}
	}()

	// Wait for all workers to finish.
	wg.Wait()

	// Determine the error to return.
	if firstErr != nil {
		return nil, firstErr
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	return results, nil
}
