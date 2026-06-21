package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input concurrently,
// preserving the order of results.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if fn == nil {
		return nil, errors.New("fn must not be nil")
	}
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be positive, got %d", workers)
	}

	if len(input) == 0 {
		return []int{}, nil
	}

	// Create a cancellable context for our workers
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Channel to send work items
	type workItem struct {
		index int
		value int
	}
	workCh := make(chan workItem, len(input))

	// Channel to receive results
	type result struct {
		index int
		value int
		err   error
	}
	resultCh := make(chan result, len(input))

	// Start workers
	var wg sync.WaitGroup
	numWorkers := workers
	if workers > len(input) {
		numWorkers = len(input)
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case work, ok := <-workCh:
					if !ok {
						return
					}

					func() {
						defer func() {
							if r := recover(); r != nil {
								resultCh <- result{
									index: work.index,
									err:   fmt.Errorf("panic in fn: %v", r),
								}
							}
						}()

						val, err := fn(ctx, work.value)
						resultCh <- result{
							index: work.index,
							value: val,
							err:   err,
						}
					}()
				}
			}
		}()
	}

	// Send work
	go func() {
		defer close(workCh)
		for i, v := range input {
			select {
			case <-ctx.Done():
				return
			case workCh <- workItem{index: i, value: v}:
			}
		}
	}()

	// Collect results in order
	output := make([]int, len(input))
	var firstErr error
	completed := 0

	for completed < len(input) {
		select {
		case <-ctx.Done():
			firstErr = ctx.Err()
			cancel()
			wg.Wait()
			return nil, firstErr
		case res := <-resultCh:
			if res.err != nil && firstErr == nil {
				firstErr = res.err
				cancel()
			}
			if res.err == nil {
				output[res.index] = res.value
			}
			completed++
		}
	}

	// Wait for all workers to finish before returning
	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return output, nil
}
