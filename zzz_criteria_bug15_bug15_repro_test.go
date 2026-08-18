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

func TestReplayKeepsCancel(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 15, 0, 0, time.FixedZone("CST", 8*3600))}
	cfg := config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30}
	y, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name:  "keep-cancel",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := y.Cancel(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := y.Close(); err != nil {
		t.Fatal(err)
	}
	y2, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y2.Close() })
	job, err := y2.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != types.JobCanceled {
		t.Fatalf("status %s after reopen", job.Status)
	}
}
