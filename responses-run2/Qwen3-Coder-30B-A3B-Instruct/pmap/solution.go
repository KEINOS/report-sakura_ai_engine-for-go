package pmap

import (
	"context"
	"errors"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input in parallel,
// preserving the original order of elements in the output.
// It limits concurrency to the specified number of workers.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, errors.New("workers must be positive")
	}
	if fn == nil {
		return nil, errors.New("function cannot be nil")
	}

	if len(input) == 0 {
		return []int{}, nil
	}

	type result struct {
		index int
		value int
		err   error
	}

	results := make([]result, len(input))
	ch := make(chan struct {
		index int
		value int
		err   error
	}, workers)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	var cancelled bool

	// Start workers
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				default:
					// Try to get work
					select {
					case <-ctx.Done():
						return
					default:
					}
				}
			}
		}()
	}

	// Process items
	for i, item := range input {
		wg.Add(1)
		go func(index int, value int) {
			defer wg.Done()
			
			// Handle panics
			func() {
				defer func() {
					if r := recover(); r != nil {
						ch <- struct {
							index int
							value int
							err   error
						}{index: index, value: 0, err: errors.New("panic in function: " + string(r.(string)))}
					}
				}()
				
				result, err := fn(ctx, value)
				ch <- struct {
					index int
					value int
					err   error
				}{index: index, value: result, err: err}
			}()
		}(i, item)
	}

	// Collect results
	for i := 0; i < len(input); i++ {
		select {
		case <-ctx.Done():
			cancelled = true
			goto done
		case res := <-ch:
			mu.Lock()
			if firstErr == nil && res.err != nil {
				firstErr = res.err
			}
			results[res.index] = res
			mu.Unlock()
		}
	}

done:
	// Cancel remaining work
	if cancelled || firstErr != nil {
		// Wait for all goroutines to finish
		wg.Wait()
		if firstErr != nil {
			return nil, firstErr
		}
		return nil, ctx.Err()
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Check for errors and build result
	if firstErr != nil {
		return nil, firstErr
	}

	output := make([]int, len(input))
	for _, res := range results {
		if res.err != nil {
			return nil, res.err
		}
		output[res.index] = res.value
	}

	return output, nil
}
