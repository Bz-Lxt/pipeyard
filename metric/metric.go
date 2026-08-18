// Package metric 收集作业场计数，供 /v1/stats 使用。
package metric

import "sync"

// Snapshot 是只读副本。
type Snapshot struct {
	Submitted  int64
	Ticked     int64
	Completed  int64
	Failed     int64
	Canceled   int64
	Checkpoints int64
	Leases     int64
}

// Counters 原子累加。
type Counters struct {
	mu sync.Mutex
	s  Snapshot
}

func (c *Counters) AddSubmitted(n int64)  { c.add(func(s *Snapshot) { s.Submitted += n }) }
func (c *Counters) AddTicked(n int64)     { c.add(func(s *Snapshot) { s.Ticked += n }) }
func (c *Counters) AddCompleted(n int64)  { c.add(func(s *Snapshot) { s.Completed += n }) }
func (c *Counters) AddFailed(n int64)     { c.add(func(s *Snapshot) { s.Failed += n }) }
func (c *Counters) AddCanceled(n int64)   { c.add(func(s *Snapshot) { s.Canceled += n }) }
func (c *Counters) AddCheckpoint(n int64) { c.add(func(s *Snapshot) { s.Checkpoints += n }) }
func (c *Counters) AddLease(n int64)      { c.add(func(s *Snapshot) { s.Leases += n }) }

func (c *Counters) add(fn func(*Snapshot)) {
	c.mu.Lock()
	fn(&c.s)
	c.mu.Unlock()
}

func (c *Counters) Snapshot() Snapshot {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.s
}

func (c *Counters) Reset() {
	c.mu.Lock()
	c.s = Snapshot{}
	c.mu.Unlock()
}
