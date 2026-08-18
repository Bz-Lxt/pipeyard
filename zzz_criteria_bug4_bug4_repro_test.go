package pipeyard_test

import (
	"context"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestParseKeepsFullDigest(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name:  "keep-id",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := digest.Parse(string(id))
	if err != nil {
		t.Fatal(err)
	}
	if parsed != id {
		t.Fatalf("Parse folded %s -> %s", id, parsed)
	}
	job, err := y.Get(ctx, parsed)
	if err != nil {
		t.Fatalf("Get after Parse: %v", err)
	}
	if job.Name != "keep-id" {
		t.Fatalf("name %s", job.Name)
	}
}
