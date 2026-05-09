package handler

import (
	"errors"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

type errorMapping struct {
	err     error
	status  int
	code    string
	message string
}

func (h *InternalSessionHandler) handleInitSessionError(w http.ResponseWriter, err error) {
	mappings := []errorMapping{
		{err: session.ErrSessionNotFound, status: http.StatusNotFound, code: "session_not_found", message: "session not found"},
		{err: session.ErrSessionRuntimeConflict, status: http.StatusConflict, code: "session_runtime_conflict", message: "session runtime conflict"},
		{err: session.ErrSessionAlreadyFinished, status: http.StatusConflict, code: "already_finished", message: "session already finished"},
		{err: session.ErrBootstrapFetchFailed, status: http.StatusFailedDependency, code: "bootstrap_fetch_failed", message: "failed to fetch bootstrap data"},
		{err: session.ErrRoomCodeUnavailable, status: http.StatusServiceUnavailable, code: "room_code_unavailable", message: "room code unavailable"},
		{err: session.ErrRuntimeStoreUnavailable, status: http.StatusServiceUnavailable, code: "redis_unavailable", message: "runtime store unavailable"},
	}

	if !writeMappedServiceError(w, err, mappings) {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func (h *InternalSessionHandler) handleGetSessionRuntimeError(w http.ResponseWriter, err error) {
	mappings := []errorMapping{
		{err: session.ErrSessionRuntimeNotFound, status: http.StatusNotFound, code: "session_runtime_not_found", message: "session runtime not found"},
		{err: session.ErrRuntimeStoreUnavailable, status: http.StatusServiceUnavailable, code: "redis_unavailable", message: "runtime store unavailable"},
	}

	if !writeMappedServiceError(w, err, mappings) {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func (h *InternalSessionHandler) handleDeleteSessionRuntimeError(w http.ResponseWriter, err error) {
	mappings := []errorMapping{
		{err: session.ErrRuntimeStoreUnavailable, status: http.StatusServiceUnavailable, code: "redis_unavailable", message: "runtime store unavailable"},
	}

	if !writeMappedServiceError(w, err, mappings) {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func writeMappedServiceError(w http.ResponseWriter, err error, mappings []errorMapping) bool {
	for _, mapping := range mappings {
		if errors.Is(err, mapping.err) {
			response.Error(w, mapping.status, mapping.code, mapping.message)
			return true
		}
	}

	return false
}
