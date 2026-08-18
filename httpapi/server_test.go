package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Bz-Lxt/pipeyard/clock"
	"github.com/Bz-Lxt/pipeyard/config"
	"github.com/Bz-Lxt/pipeyard/engine"
)

func TestHealth(t *testing.T) {
	y, err := engine.Open(config.Config{
		Dir: t.TempDir(), Clock: clock.Fixed{T: time.Date(2026, 8, 18, 16, 0, 0, 0, time.FixedZone("CST", 8*3600))},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer y.Close()
	s := New(y, ":0", "web")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	s.http.Handler.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["status"] != "ok" {
		t.Fatalf("%v", body)
	}
}
