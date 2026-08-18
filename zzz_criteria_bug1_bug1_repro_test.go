package pipeyard_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestConcurrentTickAllSucceed(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 8, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	if _, err := y.Submit(ctx, types.Spec{
		Name:  "conc",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	}); err != nil {
		t.Fatal(err)
	}

	const workers = 32
	var start, done sync.WaitGroup
	start.Add(workers)
	done.Add(workers)
	errs := make(chan error, workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer done.Done()
			start.Done()
			start.Wait()
			_, err := y.Tick(ctx, 1)
			errs <- err
		}()
	}
	done.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent Tick returned %v", err)
		}
	}
	jobs, err := y.List(ctx)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("list n=%d err=%v", len(jobs), err)
	}
	node, ok := jobs[0].Node("a")
	if !ok || node.Status != types.NodeSucceeded {
		t.Fatalf("node a status=%v", node.Status)
	}
	if y.Metrics().Completed != 1 {
		t.Fatalf("completed=%d want 1", y.Metrics().Completed)
	}
}
