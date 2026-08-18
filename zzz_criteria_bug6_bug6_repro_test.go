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

func TestCheckpointThenSubmit(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	if _, err := y.Submit(ctx, types.Spec{
		Name:  "before-cp",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	}); err != nil {
		t.Fatal(err)
	}
	if err := y.Checkpoint(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := y.Submit(ctx, types.Spec{
		Name:  "after-cp",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	}); err != nil {
		t.Fatalf("submit after checkpoint: %v", err)
	}
}
