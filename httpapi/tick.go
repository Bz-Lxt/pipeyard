package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func (s *Server) handleTick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	n := 1
	if r.URL.Query().Get("n") != "" {
		if v, err := strconv.Atoi(r.URL.Query().Get("n")); err == nil {
			n = v
		}
	}
	if r.Header.Get("Content-Type") != "" && r.ContentLength > 0 {
		var body struct {
			N int `json:"n"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err == nil && body.N > 0 {
			n = body.N
		}
	}
	ran, err := s.yard.Tick(r.Context(), n)
	if err != nil {
		writeErr(w, statusOf(err), err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"ran": ran})
}
