package pipeyard_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/types"
	"github.com/Bz-Lxt/pipeyard/worker"
)

func TestReleaseClearsLease(t *testing.T) {
	b := worker.NewBook()
	b.Hold("job", "n", "w1")
	b.Release("job", "n")
	if _, ok := b.Owner("job", "n"); ok {
		t.Fatal("lease book still holds after release")
	}

	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 10, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name:  "lease-clear",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if n, err := y.Tick(ctx, 1); err != nil || n != 1 {
		t.Fatalf("tick n=%d err=%v", n, err)
	}
	job, err := y.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	node, ok := job.Node("a")
	if !ok {
		t.Fatal("missing node a")
	}
	if node.LeaseBy != "" {
		t.Fatalf("lease_by=%q after node finished", node.LeaseBy)
	}
}
