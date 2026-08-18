package pipeyard_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
	"github.com/Bz-Lxt/pipeyard/httpapi"
	"github.com/Bz-Lxt/pipeyard/types"
)

func TestHTTPTickNRunsNodes(t *testing.T) {
	dir := t.TempDir()
	clk := clock.Fixed{T: time.Date(2026, 8, 18, 20, 16, 0, 0, time.FixedZone("CST", 8*3600))}
	y, err := engine.Open(config.Config{Dir: dir, Clock: clk, Workers: 4, MaxJobs: 32, LeaseSec: 30})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = y.Close() })
	ctx := context.Background()
	id, err := y.Submit(ctx, types.Spec{
		Name: "multi-tick",
		Nodes: []types.NodeSpec{
			{ID: "a", Kind: types.KindValidate, Param: "one"},
			{ID: "b", Kind: types.KindValidate, Param: "two"},
			{ID: "c", Kind: types.KindValidate, Param: "three"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	if err := ln.Close(); err != nil {
		t.Fatal(err)
	}
	s := httpapi.New(y, addr, t.TempDir())
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
	for time.Now().Before(deadline) {
		resp, err := http.Get(base + "/health")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == 200 {
				break
			}
		}
	}

	req := httptest.NewRequest(http.MethodPost, base+"/v1/tick?n=2", nil)
	req.RequestURI = ""
	rec, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer rec.Body.Close()
	if rec.StatusCode != 200 {
		t.Fatalf("tick http %d", rec.StatusCode)
	}
	var body map[string]int
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	job, err := y.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	advanced := 0
	for _, n := range job.Nodes {
		if n.Status != types.NodePending && n.Status != types.NodeReady {
			advanced++
		}
	}
	if body["ran"] == 0 && advanced == 0 {
		t.Fatalf("POST /v1/tick?n=2 ran=%d advanced=%d", body["ran"], advanced)
	}
	if body["ran"] != 2 && advanced < 2 {
		t.Fatalf("want ran=2 or two nodes advanced, ran=%d advanced=%d", body["ran"], advanced)
	}
}
