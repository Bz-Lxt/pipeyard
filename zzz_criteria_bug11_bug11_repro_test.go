package pipeyard_test

import (
	"context"
	"testing"

	"github.com/Bz-Lxt/pipeyard/stage"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestPersistNilFails(t *testing.T) {
	_, err := stage.Persist(context.Background(), stage.Request{
		Node: types.NodeSpec{ID: "p", Kind: types.KindPersist},
	})
	if err == nil {
		t.Fatal("expected persist without a parent or param to fail")
	}
}
