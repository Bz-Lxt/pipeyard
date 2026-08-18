package engine

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/types"
)

func openYard(t *testing.T) *Yard {
	t.Helper()
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 14, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	return y
}

func twoNodeSpec(name, body string) types.Spec {
	return types.Spec{
		Name: name,
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: body},
			{ID: "b", Kind: types.KindMap, Param: "x:"},
		},
		Edges: []types.Edge{{From: "a", To: "b"}},
	}
}

func TestSubmitTickGet(t *testing.T) {
	y := openYard(t)
	ctx := context.Background()
	id, err := y.Submit(ctx, twoNodeSpec("demo", "hello"))
	if err != nil {
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
	b, ok := job.Node("b")
	if !ok || !b.HasArtifact() {
		t.Fatalf("b missing artifact: %+v", b)
	}
	if job.Status != types.JobSucceeded {
		t.Fatalf("status %s", job.Status)
	}
}

func TestReplayKeepsSubmit(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 15, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	cfg := config.Config{Dir: dir, Clock: clk}
	y, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	id, err := y.Submit(ctx, twoNodeSpec("keep", "payload"))
	if err != nil {
		t.Fatal(err)
	}
	_ = y.Close()
	y2, err := Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer y2.Close()
	job, err := y2.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Name != "keep" {
		t.Fatalf("name %s", job.Name)
	}
}

func TestCanceledContextDoesNotInsert(t *testing.T) {
	y := openYard(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := y.Submit(ctx, twoNodeSpec("nope", "x"))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("want canceled, got %v", err)
	}
	jobs, err := y.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 0 {
		t.Fatalf("inserted %d", len(jobs))
	}
}

func TestListCopy(t *testing.T) {
	y := openYard(t)
	ctx := context.Background()
	if _, err := y.Submit(ctx, twoNodeSpec("copy", "z")); err != nil {
		t.Fatal(err)
	}
	a, err := y.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	a[0].Name = "mutated"
	if len(a[0].Nodes) > 0 {
		a[0].Nodes[0].Status = types.NodeFailed
	}
	b, err := y.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if b[0].Name != "copy" {
		t.Fatalf("polluted name %s", b[0].Name)
	}
	if b[0].Nodes[0].Status == types.NodeFailed {
		t.Fatal("polluted node status")
	}
}

func TestRejectCycle(t *testing.T) {
	y := openYard(t)
	_, err := y.Submit(context.Background(), types.Spec{
		Name: "cyc",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "1"},
			{ID: "b", Kind: types.KindMap, Param: "2"},
		},
		Edges: []types.Edge{{From: "a", To: "b"}, {From: "b", To: "a"}},
	})
	if !errors.Is(err, types.ErrCycle) {
		t.Fatalf("got %v", err)
	}
}

func TestSQLitePath(t *testing.T) {
	y := openYard(t)
	if filepath.Base(y.Config().SQLitePath()) != "yard.sqlite" {
		t.Fatal(y.Config().SQLitePath())
	}
}
