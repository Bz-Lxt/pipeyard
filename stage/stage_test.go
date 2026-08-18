package stage

import (
	"context"
	"testing"

	"github.com/Bz-Lxt/pipeyard/types"
)

func TestPipelineKinds(t *testing.T) {
	reg := NewRegistry()
	ctx := context.Background()
	v, err := reg.Run(ctx, Request{Node: types.NodeSpec{ID: "a", Kind: types.KindValidate, Param: "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	p := v.Artifact
	m, err := reg.Run(ctx, Request{Node: types.NodeSpec{ID: "b", Kind: types.KindMap, Param: "k:"}, Parents: []*types.Artifact{&p}})
	if err != nil {
		t.Fatal(err)
	}
	if !m.Artifact.Valid() {
		t.Fatal("invalid")
	}
}

func TestJoinRejectsNil(t *testing.T) {
	_, err := Join(context.Background(), Request{
		Node:    types.NodeSpec{ID: "j", Kind: types.KindJoin},
		Parents: []*types.Artifact{nil, &types.Artifact{}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
