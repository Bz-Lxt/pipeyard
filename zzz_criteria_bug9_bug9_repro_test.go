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

func TestNowTextFollowsClock(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 8, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name:  "clock-follow",
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
	want := clock.Format(clk.Now())
	if job.UpdatedAt == "1970-01-01 08:00:00" {
		t.Fatalf("updated_at pinned to epoch")
	}
	if job.UpdatedAt != want {
		t.Fatalf("updated_at=%q want %q", job.UpdatedAt, want)
	}
}
