package main

import (
	"context"
	"fmt"
	"sync"
)

// ParallelMapOrdered performs a parallel map operation on the input slice,
// preserving the order of the input elements. It runs no more than `workers`
// concurrent calls at a time.
func ParallelMapOrdered(
	ctx context.Context,
	input []int,
	workers int,
	fn func(context.Context, int) (int, error),
) ([]int, error) {
	if workers <= 0 {
		return nil, fmt.Errorf("workers must be greater than 0")
	}
	if fn == nil {
		return nil, fmt.Errorf("fn cannot be nil")
	}

	var (
		wg       sync.WaitGroup
		results  []int
		errChan  chan error
		itemChan chan int
	)

	// Initialize channels and wait group
	errChan = make(chan error, workers)
	itemChan = make(chan int, len(input))

	// Start worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for item := range itemChan {
				if fnResult, fnErr := fn(ctx, item); fnErr != nil {
					errChan <- fmt.Errorf("fn panicked: %v", fnErr)
					return
				} else {
					results = append(results, fnResult)
				}
			}
		}()
	}

	// Feed input items to worker goroutines
	for _, item := range input {
		itemChan <- item
	}
	close(itemChan)

	// Wait for all workers to finish
	wg.Wait()

	// Close errChan and wait for all errors
	close(errChan)
	var finalErr error
	for err := range errChan {
		if err != nil {
			finalErr = err
			break
		}
	}

	// Cancel remaining work and wait for all started goroutines
	if finalErr != nil {
		ctx, cancel := context.WithCancel(ctx)
		cancel()
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ctx.Done()
		}()
		wg.Wait()
	}

	return results, finalErr
}
