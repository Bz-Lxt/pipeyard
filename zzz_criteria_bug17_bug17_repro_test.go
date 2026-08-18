package pipeyard_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/httpapi"
)

func TestBadJobIDIs400(t *testing.T) {
	y, err := engine.Open(config.Config{
		Dir:   t.TempDir(),
		Clock: clock.Fixed{T: time.Date(2026, 8, 18, 20, 0, 0, 0, time.FixedZone("CST", 8*3600))},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	s := httpapi.New(y, addr, "web")
	done := make(chan struct{})
	go func() {
		_ = s.ListenAndServe()
		close(done)
	}()
	t.Cleanup(func() {
		_ = s.Shutdown(context.Background())
		<-done
	})

	base := "http://" + addr
	deadline := time.Now().Add(3 * time.Second)
	ready := false
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				ready = true
				break
			}
		}
	}
	if !ready {
		t.Fatal("server not ready")
	}

	resp, err := http.Get(base + "/v1/jobs/not-a-hex")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("GET /v1/jobs/not-a-hex status=%d want 400", resp.StatusCode)
	}
}
