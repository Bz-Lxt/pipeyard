// Package httpapi 提供作业场 HTTP 面。
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Bz-Lxt/pipeyard/engine"
)

type Server struct {
	yard *engine.Yard
	addr string
	web  string
	http *http.Server
}

func New(y *engine.Yard, addr, web string) *Server {
	if addr == "" {
		addr = ":8080"
	}
	if web == "" {
		web = "web"
	}
	s := &Server{yard: y, addr: addr, web: web}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/v1/jobs", s.handleJobs)
	mux.HandleFunc("/v1/jobs/", s.handleJob)
	mux.HandleFunc("/v1/tick", s.handleTick)
	mux.HandleFunc("/v1/stats", s.handleStats)
	mux.HandleFunc("/v1/checkpoint", s.handleCheckpoint)
	mux.HandleFunc("/v1/events", s.handleEvents)
	mux.HandleFunc("/", s.handleIndex)
	s.http = &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	return s
}

func (s *Server) ListenAndServe() error { return s.http.ListenAndServe() }

func (s *Server) Shutdown(ctx context.Context) error { return s.http.Shutdown(ctx) }

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, err error) {
	msg := "error"
	if err != nil {
		msg = err.Error()
	}
	writeJSON(w, code, map[string]string{"error": msg})
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	p := filepath.Join(s.web, "index.html")
	if _, err := os.Stat(p); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(fallbackHTML))
		return
	}
	http.ServeFile(w, r, p)
}

func trimJobID(path string) string {
	path = strings.TrimPrefix(path, "/v1/jobs/")
	path = strings.Trim(path, "/")
	if i := strings.Index(path, "/"); i >= 0 {
		return path[:i]
	}
	return path
}

const fallbackHTML = `<!doctype html><html lang="zh-CN"><meta charset="utf-8"><title>Pipeyard</title>
<body style="font-family:ui-sans-serif;padding:2rem"><h1>Pipeyard</h1><p>作业场在线。</p></body>`
