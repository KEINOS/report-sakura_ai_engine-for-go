package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input in parallel while
// preserving the order of results. At most workers calls to fn are executed
// concurrently.
//
// The function returns a slice of the same length as input containing the
// results in the original order, or an error if fn returns an error, the
// provided context is cancelled, or a panic occurs inside fn.
//
// The first encountered error (including a panic) cancels the remaining work.
// All started goroutines are waited for before returning.
//
// Errors from fn are returned unchanged so that errors.Is works as expected.
// A panic inside fn is converted to an error of the form
//   "panic in fn: <value>"
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, errors.New("workers must be greater than 0")
	}
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}
	if len(input) == 0 {
		return []int{}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	type task struct {
		idx int
		val int
	}

	tasks := make(chan task, workers)
	results := make([]int, len(input))

	// Context that can be cancelled on first error.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 1) // buffer 1 to avoid blocking

	worker := func() {
		defer wg.Done()
		for t := range tasks {
			// Call fn with panic recovery.
			res, err := func() (int, error) {
				defer func() {
					if r := recover(); r != nil {
						err = fmt.Errorf("panic in fn: %v", r)
					}
				}()
				return fn(ctx, t.val)
			}()
			if err != nil {
				// Record the first error and cancel the context.
				select {
				case errCh <- err:
				default:
				}
				cancel()
				return
			}
			results[t.idx] = res
		}
	}

	// Start workers.
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}

	// Feed tasks.
	go func() {
		defer close(tasks)
		for i, v := range input {
			select {
			case <-ctx.Done():
				return
			case tasks <- task{idx: i, val: v}:
			}
		}
	}()

	// Wait for all workers to finish.
	wg.Wait()

	// Return the first error if any.
	select {
	case err := <-errCh:
		return nil, err
	default:
		return results, nil
	}
}
