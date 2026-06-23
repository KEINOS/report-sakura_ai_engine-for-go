package pmap

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInvalidWorkers = errors.New("workers must be greater than 0")
	ErrNilFunction    = errors.New("function cannot be nil")
	ErrInvalidInput   = errors.New("input cannot be nil")
)

func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}
	if fn == nil {
		return nil, ErrNilFunction
	}
	if input == nil {
		return []int{}, nil
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
				errorsCh <- ctx.Err()
				return
			default:
				result, err := fn(ctx, v)
				if err != nil {
					errorsCh <- err
					return
				}
				resultCh <- result
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
