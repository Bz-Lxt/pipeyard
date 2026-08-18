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

func TestFormatWritesWallClock(t *testing.T) {
	ts := time.Date(2026, 8, 18, 20, 15, 0, 0, time.FixedZone("CST", 8*3600))
	if clock.Format(ts) == "" {
		t.Fatal("Format of a non-zero time returned empty text")
	}

	y, err := engine.Open(config.Config{Dir: t.TempDir(), Clock: clock.Fixed{T: ts}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	id, err := y.Submit(context.Background(), types.Spec{
		Name:  "wall",
		Nodes: []types.NodeSpec{{ID: "a", Kind: types.KindValidate, Param: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	job, err := y.Get(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	if job.CreatedAt == "" {
		t.Fatal("CreatedAt is empty after submit")
	}
}
