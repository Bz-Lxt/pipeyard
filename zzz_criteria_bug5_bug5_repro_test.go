package pipeyard_test

import (
	"errors"
	"testing"

	"github.com/Bz-Lxt/pipeyard/quota"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestQuotaAcquireStopsAtLimit(t *testing.T) {
	c := quota.New(1)
	if err := c.Acquire(); err != nil {
		t.Fatal(err)
	}
	err := c.Acquire()
	if err == nil {
		t.Fatal("second Acquire succeeded")
	}
	if !errors.Is(err, types.ErrQuota) {
		t.Fatalf("want quota exceeded, got %v", err)
	}
}
