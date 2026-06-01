package handler

import (
	"errors"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

func (h *InternalSessionHandler) handleInitSessionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, session.ErrSessionNotFound):
		response.Error(w, http.StatusNotFound, "session_not_found", "session not found")
	case errors.Is(err, session.ErrSessionRuntimeConflict):
		response.Error(w, http.StatusConflict, "session_runtime_conflict", "session runtime conflict")
	case errors.Is(err, session.ErrSessionAlreadyFinished):
		response.Error(w, http.StatusConflict, "already_finished", "session already finished")
	case errors.Is(err, session.ErrBootstrapFetchFailed):
		response.Error(w, http.StatusFailedDependency, "bootstrap_fetch_failed", "failed to fetch bootstrap data")
	case errors.Is(err, session.ErrRoomCodeUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "room_code_unavailable", "room code unavailable")
	case errors.Is(err, session.ErrRuntimeStoreUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "redis_unavailable", "runtime store unavailable")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func (h *InternalSessionHandler) handleGetSessionRuntimeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, session.ErrSessionRuntimeNotFound):
		response.Error(w, http.StatusNotFound, "session_runtime_not_found", "session runtime not found")
	case errors.Is(err, session.ErrRuntimeStoreUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "redis_unavailable", "runtime store unavailable")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}

func (h *InternalSessionHandler) handleDeleteSessionRuntimeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, session.ErrRuntimeStoreUnavailable):
		response.Error(w, http.StatusServiceUnavailable, "redis_unavailable", "runtime store unavailable")
	default:
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
	}
}
