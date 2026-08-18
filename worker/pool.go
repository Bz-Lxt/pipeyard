package worker

import (
	"context"
	"sync"
)

// Task 是一次可取消的节点执行。
type Task func(ctx context.Context) error

// Pool 限制同时跑的阶段数。
type Pool struct {
	sem chan struct{}
}

func NewPool(n int) *Pool {
	if n <= 0 {
		n = 1
	}
	return &Pool{sem: make(chan struct{}, n)}
}

func (p *Pool) Do(ctx context.Context, fn Task) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.sem <- struct{}{}:
	}
	defer func() { <-p.sem }()
	if err := ctx.Err(); err != nil {
		return err
	}
	return fn(ctx)
}

func (p *Pool) DoAll(ctx context.Context, tasks []Task) error {
	var (
		wg   sync.WaitGroup
		once sync.Once
		ret  error
	)
	for _, t := range tasks {
		t := t
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := p.Do(ctx, t); err != nil {
				once.Do(func() { ret = err })
			}
		}()
	}
	wg.Wait()
	return ret
}

func (p *Pool) Cap() int { return cap(p.sem) }
