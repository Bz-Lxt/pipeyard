package graph

import (
	"testing"

	"github.com/Bz-Lxt/pipeyard/types"
)

func TestReadyCopy(t *testing.T) {
	g, err := Build(types.Spec{
		Name: "g",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "1"},
			{ID: "b", Kind: types.KindMap, Param: "2"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	st := map[string]types.NodeStatus{"a": types.NodePending, "b": types.NodePending}
	r := g.Ready(st)
	r[0] = "mutated"
	r2 := g.Ready(st)
	if r2[0] == "mutated" {
		t.Fatal("shared backing array")
	}
}

func TestCycle(t *testing.T) {
	_, err := Build(types.Spec{
		Name: "c",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "1"},
			{ID: "b", Kind: types.KindMap, Param: "2"},
		},
		Edges: []types.Edge{{From: "a", To: "b"}, {From: "b", To: "a"}},
	})
	if err != types.ErrCycle {
		t.Fatalf("got %v", err)
	}
}
