package pmap

import (
	"context"
	"errors"
	"sync"
)

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
		return nil, errors.New("function cannot be nil")
	}

	if len(input) == 0 {
		return []int{}, nil
	}

	var wg sync.WaitGroup
	results := make([]int, len(input))
	errorsCh := make(chan error, len(input))
	resultCh := make(chan int, len(input))

	for i, v := range input {
		wg.Add(1)
		go func(i int, v int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case resultCh <- fn(ctx, v).(int):
			case err := <-errorsCh:
				if err != nil {
					errorsCh <- err
					return
				}
			}
		}(i, v)
	}

	go func() {
		wg.Wait()
		close(resultCh)
		close(errorsCh)
	}()

	for i := 0; i < len(input); i++ {
		select {
		case result := <-resultCh:
			results[i] = result
		case err := <-errorsCh:
			return nil, err
		}
	}

	return results, nil
}
