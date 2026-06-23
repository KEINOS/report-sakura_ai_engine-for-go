package pmap

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if fn == nil {
		return nil, errors.New("function cannot be nil")
	}
	if workers <= 0 {
		return nil, errors.New("workers must be positive")
	}

	if len(input) == 0 {
		return []int{}, nil
	}

	// Use at most the number of input elements as workers
	workers = min(workers, len(input))

	// Channel to receive results in order
	resultCh := make(chan result, len(input))
	// Channel to signal completion
	doneCh := make(chan struct{})
	// Wait group for goroutines
	var wg sync.WaitGroup
	// Mutex for result ordering
	var mu sync.Mutex
	// Store results in order
	results := make([]int, len(input))
	// Track which results are ready
	ready := make([]bool, len(input))
	// Track the next expected index
	nextIndex := 0
	// Error channel
	errCh := make(chan error, 1)

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case idx, ok := <-resultCh:
					if !ok {
						return
					}
					// Process the result
					mu.Lock()
					if !ready[idx] {
						// Store result in order
						results[idx] = idx
						ready[idx] = true
						// Check if we have all results
						for nextIndex < len(input) && ready[nextIndex] {
							nextIndex++
						}
					}
					mu.Unlock()
				}
			}
		}()
	}

	// Start processing
	go func() {
		defer close(resultCh)
		defer close(doneCh)
		defer wg.Wait()

		// Process each input
		for i, val := range input {
			select {
			case <-ctx.Done():
				return
			case resultCh <- result{i, val}:
			}
		}
	}()

	// Wait for all results or error
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case err := <-errCh:
			return nil, err
		case <-doneCh:
			// All workers done, check if we have all results
			if nextIndex == len(input) {
				return results, nil
			}
			// Should not happen, but return what we have
			return results, nil
		}
	}
}

// result represents a result from a worker
type result struct {
	index int
	value int
}
