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

	n := len(input)
	if n == 0 {
		return make([]int, 0), nil
	}

	if workers > n {
		workers = n
	}

	type job struct {
		idx int
		val int
	}

	jobs := make(chan job, workers)
	out := make([]int, n)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup
	var errOnce sync.Once
	var firstErr error

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case j, ok := <-jobs:
					if !ok {
						return
					}
					if ctx.Err() != nil {
						return
					}
					func() {
						defer func() {
							if r := recover(); r != nil {
								errOnce.Do(func() {
									firstErr = fmt.Errorf("panic in worker: %v", r)
									cancel()
								})
							}
						}()
						res, err := fn(ctx, j.val)
						if err != nil {
							errOnce.Do(func() {
								firstErr = err
								cancel()
							})
							return
						}
						out[j.idx] = res
					}()
				}
			}
		}()
	}

loop:
	for i := 0; i < n; i++ {
		select {
		case <-ctx.Done():
			break loop
		case jobs <- job{idx: i, val: input[i]}:
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
