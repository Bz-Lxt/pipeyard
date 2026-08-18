package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Bz-Lxt/pipeyard/digest"
	"github.com/Bz-Lxt/pipeyard/types"
)

type submitBody struct {
	Name  string          `json:"name"`
	Nodes []types.NodeSpec `json:"nodes"`
	Edges []types.Edge    `json:"edges"`
	Quota int             `json:"quota"`
}

func (s *Server) handleJobs(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		jobs, err := s.yard.List(r.Context())
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err)
			return
		}
		writeJSON(w, http.StatusOK, jobs)
	case http.MethodPost:
		var body submitBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, http.StatusBadRequest, err)
			return
		}
		id, err := s.yard.Submit(r.Context(), types.Spec{Name: body.Name, Nodes: body.Nodes, Edges: body.Edges, Quota: body.Quota})
		if err != nil {
			writeErr(w, statusOf(err), err)
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"id": string(id)})
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleJob(w http.ResponseWriter, r *http.Request) {
	idText := trimJobID(r.URL.Path)
	if idText == "" {
		http.NotFound(w, r)
		return
	}
	id, err := digest.Parse(idText)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/cancel") && r.Method == http.MethodPost {
		if err := s.yard.Cancel(r.Context(), id); err != nil {
			writeErr(w, statusOf(err), err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "canceled"})
		return
	}
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	job, err := s.yard.Get(r.Context(), id)
	if err != nil {
		writeErr(w, statusOf(err), err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func statusOf(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case err == types.ErrNotFound:
		return http.StatusNotFound
	case err == types.ErrBadSpec, err == types.ErrEmptyGraph, err == types.ErrUnknownKind, err == types.ErrUnknownNode, err == types.ErrCycle:
		return http.StatusBadRequest
	case err == types.ErrQuota, err == types.ErrConflict, err == types.ErrDone:
		return http.StatusConflict
	case err == types.ErrClosed, err == types.ErrReadOnly:
		return http.StatusServiceUnavailable
	default:
		return http.StatusOK
	}
}
