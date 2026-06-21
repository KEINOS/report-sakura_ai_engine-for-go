package pmap

import (
	"errors"
	"fmt"
	"sync"
)

// ParallelMapOrdered implements the ParallelMapOrdered function
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 || workers == 0 {
		return nil, errors.New("number of workers cannot be zero or negative")
	}

	if input == nil {
		return nil, errors.New("input cannot be nil")
	}

	var results []int
	var wg sync.WaitGroup

	for i := 0; i < len(input); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, err := fn(ctx, i)
			if err != nil {
				if errors.Is(err, fmt.Errorf("panic")) {
					results = append(results, result)
				} else {
					results = append(results, result)
					if err != nil {
						results = append(results, err)
					}
				}
			}
		}()
	}

	wg.Wait()

	if len(results) == 0 {
		return results, nil
	}

	return results, nil
}
