package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	} else if ctx.Err() != nil {
		return nil, ctx.Err()
	} else if len(in) == 0 {
		return []Result[R]{}, nil
	} else {
		if workers > len(in) {
    		workers = len(in)
		}

		ans := make([]Result[R], len(in))
		jobs := make(chan int)

		var wg sync.WaitGroup
		for w := 0; w < workers; w++ {
    		wg.Add(1)
    		go func() {
        		defer wg.Done()
        		for i := range jobs {
            		v, err := fn(ctx, in[i])
            		if err != nil {
                		ans[i] = Result[R]{Err: err}
            		} else {
                		ans[i] = Result[R]{Value: v}
            		}		
        		}
    		}()
		}

		go func() {
    		defer close(jobs)
    		for i := range in {
        		select {
					case jobs <- i:
        			case <-ctx.Done():
            			return
        		}
    		}
		}()

		wg.Wait()

		if err := ctx.Err(); err != nil {
    		return nil, err
		}
		return ans, nil
	}
}
