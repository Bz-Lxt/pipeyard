package pipeyard_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/graph"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestReadyLimitedKeepsOrder(t *testing.T) {
	g, err := graph.Build(types.Spec{
		Name: "ready-clip",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "a"},
			{ID: "b", Kind: types.KindValidate, Param: "b"},
			{ID: "c", Kind: types.KindValidate, Param: "c"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	before := append([]string(nil), g.Order()...)
	st := map[string]types.NodeStatus{
		"a": types.NodePending,
		"b": types.NodePending,
		"c": types.NodePending,
	}
	_ = g.ReadyLimited(st, 2)
	after := g.Order()
	if len(after) != len(before) {
		t.Fatalf("order len %d -> %d", len(before), len(after))
	}
	for _, id := range after {
		if id == "ghost" {
			t.Fatal("order contains ghost")
		}
	}
	for i := range before {
		if after[i] != before[i] {
			t.Fatalf("order changed: %v -> %v", before, after)
		}
	}

	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 8, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	if _, err := y.Submit(ctx, types.Spec{
		Name: "three",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "hello"},
			{ID: "b", Kind: types.KindValidate, Param: "hello"},
			{ID: "c", Kind: types.KindValidate, Param: "hello"},
		},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := y.Tick(ctx, 2); err != nil {
		t.Fatalf("Tick(2): %v", err)
	}
	if _, err := y.Tick(ctx, 2); err != nil {
		t.Fatalf("Tick(2) again: %v", err)
	}
	jobs, err := y.List(ctx)
	if err != nil || len(jobs) != 1 {
		t.Fatalf("list n=%d err=%v", len(jobs), err)
	}
	for _, id := range []string{"a", "b", "c"} {
		node, ok := jobs[0].Node(id)
		if !ok || node.Status != types.NodeSucceeded {
			t.Fatalf("node %s status=%v ok=%v", id, node.Status, ok)
		}
	}
}
