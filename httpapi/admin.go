package httpapi

import (
	"net/http"

	"github.com/Bz-Lxt/pipeyard/export"
)

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	st, err := s.yard.Stats(r.Context())
	if err != nil {
		writeErr(w, statusOf(err), err)
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func (s *Server) handleCheckpoint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if err := s.yard.Checkpoint(r.Context()); err != nil {
		writeErr(w, statusOf(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ev := s.yard.Events(50)
	writeJSON(w, http.StatusOK, ev)
}

func (s *Server) handleExport(w http.ResponseWriter, r *http.Request) {
	jobs, err := s.yard.List(r.Context())
	if err != nil {
		writeErr(w, statusOf(err), err)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(export.JobsText(jobs)))
}
