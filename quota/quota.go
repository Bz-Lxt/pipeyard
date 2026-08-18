// Package quota 限制同时存在的作业数与并发租约。
package quota

import (
	"fmt"
	"sync"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Counter 用有符号计数，Release 不得让值变成负数后绕回。
type Counter struct {
	mu    sync.Mutex
	cur   int
	limit int
}

func New(limit int) *Counter {
	if limit <= 0 {
		limit = 1
	}
	return &Counter{limit: limit}
}

func (c *Counter) Limit() int { return c.limit }

func (c *Counter) Current() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.cur
}

func (c *Counter) Acquire() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cur >= c.limit {
		return fmt.Errorf("%w: %d/%d", types.ErrQuota, c.cur, c.limit)
	}
	c.cur++
	return nil
}

func (c *Counter) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cur--
}

func (c *Counter) Reset(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if n < 0 {
		n = 0
	}
	c.cur = n
}

func (c *Counter) Try() bool {
	return c.Acquire() == nil
}
