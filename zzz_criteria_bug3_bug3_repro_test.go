package pipeyard_test

import (
	"context"
	"testing"

	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestValidateRejectsCanceledAndEmpty(t *testing.T) {
	reg := stage.NewRegistry()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := reg.Run(ctx, stage.Request{
		Node: types.NodeSpec{ID: "a", Kind: types.KindValidate, Param: "hello"},
	}); err == nil {
		t.Fatal("canceled Run succeeded")
	}
	if _, err := stage.Validate(ctx, stage.Request{
		Node: types.NodeSpec{ID: "a", Kind: types.KindValidate, Param: "hello"},
	}); err == nil {
		t.Fatal("canceled Validate succeeded")
	}
	if _, err := stage.Validate(context.Background(), stage.Request{
		Node: types.NodeSpec{ID: "a", Kind: types.KindValidate, Param: ""},
	}); err == nil {
		t.Fatal("empty body Validate succeeded")
	}
}
