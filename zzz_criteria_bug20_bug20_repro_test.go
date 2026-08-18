package pipeyard_test

import (
	"context"
	"testing"

	"github.com/Bz-Lxt/pipeyard/shard"
	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestIndexNonNegative(t *testing.T) {
	if got := shard.Index("bravo", 4); got < 0 {
		t.Fatalf("Index(bravo,4)=%d want >=0", got)
	}
	res, err := stage.Fanout(context.Background(), stage.Request{
		Node: types.NodeSpec{ID: "b", Kind: types.KindFanout, Param: "bravo,hello"},
	})
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, it := range res.Artifact.Items {
		seen[it] = true
	}
	if !seen["bravo"] || !seen["hello"] {
		t.Fatalf("fanout items=%v want bravo and hello", res.Artifact.Items)
	}
}
