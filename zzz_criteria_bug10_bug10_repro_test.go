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

func TestFingerprintIncludesNameAndTime(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 10, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	nodes := []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}}
	id1, err := y.Submit(ctx, types.Spec{Name: "alpha-same", Nodes: nodes})
	if err != nil {
		t.Fatal(err)
	}
	id2, err := y.Submit(ctx, types.Spec{Name: "bravo-same", Nodes: nodes})
	if err != nil {
		t.Fatalf("second submit failed: %v", err)
	}
	if id1 == id2 {
		t.Fatalf("ids collided: %s", id1)
	}
}
