package pmap

import (
	"context"
	"errors"
	"sync"
)

var (
	ErrInvalidWorkers = errors.New("workers must be greater than 0")
	ErrNilFunction    = errors.New("function cannot be nil")
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

	if len(input) == 0 {
		return []int{}, nil
	}

	results := make([]int, len(input))
	var wg sync.WaitGroup
	errChan := make(chan error, workers)
	resultChan := make(chan int, len(input))

	for i, v := range input {
		wg.Add(1)
		go func(i int, v int) {
			defer wg.Done()
			select {
			case <-ctx.Done():
				return
			case resultChan <- v:
			}
		}(i, v)
	}

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
	}()

	for i := 0; i < len(input); i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case result := <-resultChan:
			results[i] = result
		case err := <-errChan:
			return nil, err
		}
	}

	return results, nil
}
