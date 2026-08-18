package pipeyard_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestGetSeesTickResult(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 14, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name: "stale-view",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "hello"},
			{ID: "b", Kind: types.KindMap, Param: "x:"},
		},
		Edges: []types.Edge{{From: "a", To: "b"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := y.Get(ctx, id); err != nil {
		t.Fatal(err)
	}
	if n, err := y.Tick(ctx, 1); err != nil || n != 1 {
		t.Fatalf("tick1 n=%d err=%v", n, err)
	}
	if n, err := y.Tick(ctx, 1); err != nil || n != 1 {
		t.Fatalf("tick2 n=%d err=%v", n, err)
	}
	job, err := y.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != types.JobSucceeded {
		t.Fatalf("status %s after ticks", job.Status)
	}
	b, ok := job.Node("b")
	if !ok || !b.HasArtifact() {
		t.Fatalf("node b missing artifact: %+v", b)
	}
}
