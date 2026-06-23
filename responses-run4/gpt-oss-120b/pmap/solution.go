package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input in parallel while
// preserving the order of the results. At most workers calls to fn are
// executed concurrently.
//
// The function returns a slice of the same length as input containing the
// results in the original order, or nil and an error if fn returns an error,
// panics, or the supplied context is cancelled.
//
// Validation errors:
//   * workers <= 0  -> error
//   * fn == nil     -> error
//
// An empty input returns a non‑nil empty slice.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	// Parameter validation.
	if workers <= 0 {
		return nil, errors.New("workers must be greater than zero")
	}
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}
	if len(input) == 0 {
		return make([]int, 0), nil
	}

	// Context that can be cancelled on first error.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		idx int
		val int
	}
	tasks := make(chan task, workers)

	results := make([]int, len(input))

	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	// Worker function.
	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case t, ok := <-tasks:
				if !ok {
					return
				}
				// If the context was cancelled after the task was fetched,
				// skip processing.
				if ctx.Err() != nil {
					return
				}
				// Call fn with panic protection.
				res, err := func() (r int, e error) {
					defer func() {
						if rcv := recover(); rcv != nil {
							e = fmt.Errorf("panic in fn: %v", rcv)
						}
					}()
					return fn(ctx, t.val)
				}()
				if err != nil {
					once.Do(func() { firstErr = err })
					cancel()
					return
				}
				results[t.idx] = res
			}
		}
	}

	// Launch workers.
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	// Feed tasks.
	for i, v := range input {
		select {
		case <-ctx.Done():
			break
		case tasks <- task{idx: i, val: v}:
		}
	}
	close(tasks)

	// Wait for all workers to finish.
	wg.Wait()

	// Determine the error to return.
	if firstErr != nil {
		return nil, firstErr
	}
	if ctx.Err() != nil && ctx.Err() != context.Canceled {
		// Propagate cancellation or deadline errors.
		return nil, ctx.Err()
	}
	return results, nil
}
