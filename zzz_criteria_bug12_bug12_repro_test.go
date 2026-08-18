package pipeyard_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestAppendGrowsFile(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 12, 0, 0, time.FixedZone("CST", 8*3600))}
	cfg := config.Config{Dir: dir, Clock: clk, Workers: 2, MaxJobs: 32, LeaseSec: 30}
	y, err := engine.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	if _, err := y.Submit(ctx, types.Spec{
		Name:  "wal-grow",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	}); err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(cfg.JournalPath())
	if err != nil {
		t.Fatal(err)
	}
	if st.Size() <= 0 {
		t.Fatalf("journal size %d", st.Size())
	}
}
