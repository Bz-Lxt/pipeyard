// Package engine 把图、WAL、SQLite 与阶段注册表串成作业场。
package engine

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/event"
	"github.com/Bz-Lxt/pipeyard/graph"
	"github.com/Bz-Lxt/pipeyard/metric"
	"github.com/Bz-Lxt/pipeyard/policy"
	"github.com/Bz-Lxt/pipeyard/quota"
	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/store"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/wal"
	"github.com/Bz-Lxt/pipeyard/worker"
)

// Yard 是作业场门面。所有会改账本的方法必须持有 mu。
type Yard struct {
	mu       sync.Mutex
	cfg      config.Config
	db       *store.DB
	journal  *wal.Journal
	reg      *stage.Registry
	jobs     *quota.Counter
	leases   *quota.Counter
	book     *worker.Book
	bus      *event.Bus
	metrics  metric.Counters
	rule     policy.Rule
	graphs   map[string]*graph.Graph
	closed   bool
}

func (y *Yard) Config() config.Config { return y.cfg }

func (y *Yard) Clock() clock.Clock { return y.cfg.Clock }

func (y *Yard) Close() error {
	y.mu.Lock()
	defer y.mu.Unlock()
	y.closed = true
	err1 := y.journal.Close()
	err2 := y.db.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func (y *Yard) guard(ctx context.Context) error {
	if y.closed {
		return types.ErrClosed
	}
	if y.cfg.ReadOnly {
		return types.ErrReadOnly
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return nil
}

func (y *Yard) Events(n int) []event.Event { return y.bus.Recent(n) }

func (y *Yard) Metrics() metric.Snapshot { return y.metrics.Snapshot() }

func encodeSpec(spec types.Spec) []byte {
	b, _ := json.Marshal(spec)
	return b
}

func decodeSpec(raw []byte) (types.Spec, error) {
	var s types.Spec
	err := json.Unmarshal(raw, &s)
	return s, err
}
