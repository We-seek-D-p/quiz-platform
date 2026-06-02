package ws

import (
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

// Ping reports Session WS API availability for browser-side health checks.
func (h *Handler) Ping(w http.ResponseWriter, _ *http.Request) {
	response.JSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CheckRoomLiveness validates that a room can accept player WebSocket clients.
func (h *Handler) CheckRoomLiveness(w http.ResponseWriter, r *http.Request) {
	roomCode := strings.TrimSpace(chi.URLParam(r, "room_code"))
	if err := h.service.CheckRoomLiveness(r.Context(), roomCode); err != nil {
		h.writeHTTPError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "active"})
}

// CheckSessionLiveness validates that a session can accept host WebSocket clients.
func (h *Handler) CheckSessionLiveness(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(chi.URLParam(r, "session_id"))
	if err := h.service.CheckSessionLiveness(r.Context(), sessionID); err != nil {
		h.writeHTTPError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{"status": "active"})
}

func (h *Handler) writeHTTPError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}

	status := http.StatusInternalServerError
	switch appErr.Type {
	case domain.ErrTypeInvalidInput:
		status = http.StatusBadRequest
	case domain.ErrTypeForbidden:
		status = http.StatusForbidden
	case domain.ErrTypeNotFound:
		status = http.StatusNotFound
	case domain.ErrTypeConflict:
		status = http.StatusConflict
	case domain.ErrTypeInternal:
		status = http.StatusInternalServerError
	}

	response.Error(w, status, appErr.Code, appErr.Message)
}
