package pmap

import (
	"context"
	"fmt"
	"sync"
)

// ParallelMapOrdered applies fn to each element of input concurrently,
// preserving the original order in the returned slice.
func ParallelMapOrdered(ctx context.Context, input []int, workers int, fn func(context.Context, int) (int, error)) ([]int, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be positive: %d", workers)
	}
	if fn == nil {
		return nil, fmt.Errorf("fn must not be nil")
	}
	n := len(input)
	if n == 0 {
		return []int{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := make(chan struct{}, workers)
	results := make([]int, n)
	var once sync.Once
	var firstErr error

	var wg sync.WaitGroup
	wg.Add(n)

	for i, v := range input {
		go func(idx int, val int) {
			defer wg.Done()
			var recovered interface{}
			defer func() {
				if r := recover(); r != nil {
					recovered = r
				}
			}()

			select {
			case sem <- struct{}{}:
			case <-ctx.Done():
				return
			}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				return
			}

			res, err := fn(ctx, val)
			if recovered != nil {
				once.Do(func() {
					firstErr = fmt.Errorf("panic in worker: %v", recovered)
					cancel()
				})
				return
			}
			if err != nil {
				once.Do(func() {
					firstErr = err
					cancel()
				})
				return
			}

			results[idx] = res
		}(i, v)
	}

	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return results, nil
}
