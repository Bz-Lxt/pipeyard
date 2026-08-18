package httpapi

import (
	"errors"
	"net/http"

	"github.com/Bz-Lxt/pipeyard/types"
)

func mapError(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch {
	case errors.Is(err, types.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, types.ErrBadSpec), errors.Is(err, types.ErrEmptyGraph),
		errors.Is(err, types.ErrUnknownKind), errors.Is(err, types.ErrUnknownNode),
		errors.Is(err, types.ErrCycle):
		return http.StatusBadRequest
	case errors.Is(err, types.ErrQuota), errors.Is(err, types.ErrConflict), errors.Is(err, types.ErrDone):
		return http.StatusConflict
	case errors.Is(err, types.ErrClosed), errors.Is(err, types.ErrReadOnly):
		return http.StatusServiceUnavailable
	case errors.Is(err, types.ErrWAL), errors.Is(err, types.ErrCorrupt):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}
