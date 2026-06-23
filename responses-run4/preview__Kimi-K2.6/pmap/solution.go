package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

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
		return nil, errors.New("fn is nil")
	}
	if len(input) == 0 {
		return []int{}, nil
	}

	if workers > len(input) {
		workers = len(input)
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jobs := make(chan int)
	out := make([]int, len(input))

	var (
		wg       sync.WaitGroup
		firstErr error
		errOnce  sync.Once
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				if ctx.Err() != nil {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							errOnce.Do(func() {
								firstErr = fmt.Errorf("fn panicked: %v", r)
								cancel()
							})
						}
					}()
					res, err := fn(ctx, input[idx])
					if err != nil {
						errOnce.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}
					out[idx] = res
				}()
			}
		}()
	}

loop:
	for i := range input {
		select {
		case <-ctx.Done():
			break loop
		case jobs <- i:
		}
	}
	close(jobs)

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
