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
// The function returns a slice with the results in the same order as input.
// If fn returns an error, the first error is returned and all remaining work
// is cancelled. A panic inside fn is recovered and turned into an error.
//
// Errors:
//   * workers <= 0 : returns an error describing the invalid workers value.
//   * fn == nil    : returns an error describing the nil function.
//   * empty input : returns a non‑nil empty slice and nil error.
//   * parent context cancellation : returns nil and the context error.
//   * fn error   : returns nil and the first error returned by fn.
//   * fn panic   : returns nil and an error of the form "panic in fn: %v".
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("invalid workers %d: must be > 0", workers)
	}
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}
	if len(input) == 0 {
		return make([]int, 0), nil
	}

	// Child context that we can cancel on first error.
	childCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	type task struct {
		idx int
		val int
	}
	tasks := make(chan task)

	results := make([]int, len(input))

	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	worker := func() {
		defer wg.Done()
		for {
			select {
			case <-childCtx.Done():
				return
			case t, ok := <-tasks:
				if !ok {
					return
				}
				// Call fn with panic recovery.
				res, err := func() (int, error) {
					defer func() {
						if r := recover(); r != nil {
							err = fmt.Errorf("panic in fn: %v", r)
						}
					}()
					return fn(childCtx, t.val)
				}()
				if err != nil {
					once.Do(func() {
						firstErr = err
						cancel()
					})
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

	// Feed tasks, aborting early if the context is cancelled.
sendLoop:
	for i, v := range input {
		select {
		case <-childCtx.Done():
			break sendLoop
		case tasks <- task{idx: i, val: v}:
		}
	}
	close(tasks)

	// Wait for all workers to finish.
	wg.Wait()

	// Prefer the first error from fn, otherwise propagate context cancellation.
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
