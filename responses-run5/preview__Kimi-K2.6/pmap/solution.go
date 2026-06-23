package pmap

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, errors.New("pmap: workers must be > 0")
	}
	if fn == nil {
		return nil, errors.New("pmap: fn must not be nil")
	}
	if len(input) == 0 {
		return []int{}, nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	result := make([]int, len(input))

	var next atomic.Int64
	var wg sync.WaitGroup
	var once sync.Once
	var firstErr error

	w := workers
	if w > len(input) {
		w = len(input)
	}

	for i := 0; i < w; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				idx := int(next.Add(1) - 1)
				if idx >= len(input) {
					return
				}
				if err := ctx.Err(); err != nil {
					return
				}
				func() {
					defer func() {
						if r := recover(); r != nil {
							once.Do(func() {
								firstErr = fmt.Errorf("pmap: worker panic: %v", r)
								cancel()
							})
						}
					}()
					val, err := fn(ctx, input[idx])
					if err != nil {
						once.Do(func() {
							firstErr = err
							cancel()
						})
						return
					}
					result[idx] = val
				}()
			}
		}()
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}
