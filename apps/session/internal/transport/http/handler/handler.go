package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/dto"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

const idempotencyKeyHeader = "Idempotency-Key"

type InternalSessionHandler struct {
	service *session.Service
}

func NewInternalSessionHandler(service *session.Service) *InternalSessionHandler {
	return &InternalSessionHandler{service: service}
}

func (h *InternalSessionHandler) InitSession(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(chi.URLParam(r, "session_id"))
	if sessionID == "" {
		response.Error(w, http.StatusBadRequest, "invalid_payload", "invalid payload")
		return
	}

	idempotencyKey := strings.TrimSpace(r.Header.Get(idempotencyKeyHeader))
	if idempotencyKey == "" {
		response.Error(w, http.StatusBadRequest, "idempotency_key_required", "idempotency key is required")
		return
	}

	var req dto.InitSessionRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid_payload", "invalid payload")
		return
	}

	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		response.Error(w, http.StatusBadRequest, "invalid_payload", "invalid payload")
		return
	}

	if strings.TrimSpace(req.QuizID) == "" || strings.TrimSpace(req.HostID) == "" || req.CreatedAt.IsZero() {
		response.Error(w, http.StatusBadRequest, "invalid_payload", "invalid payload")
		return
	}

	result, err := h.service.InitSession(r.Context(), session.InitSessionParams{
		SessionID:      sessionID,
		QuizID:         strings.TrimSpace(req.QuizID),
		HostID:         strings.TrimSpace(req.HostID),
		CreatedAt:      req.CreatedAt,
		IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		h.handleInitSessionError(w, err)
		return
	}

	status := http.StatusCreated
	if !result.Created {
		status = http.StatusOK
	}

	response.JSON(w, status, mapRuntimeToResponse(result.Runtime))
}

func (h *InternalSessionHandler) GetSessionRuntime(w http.ResponseWriter, r *http.Request) {
	sessionID := strings.TrimSpace(chi.URLParam(r, "session_id"))
	if sessionID == "" {
		response.Error(w, http.StatusBadRequest, "invalid_payload", "invalid payload")
		return
	}

	runtime, err := h.service.GetSessionRuntime(r.Context(), session.GetSessionRuntimeParams{SessionID: sessionID})
	if err != nil {
		h.handleGetSessionRuntimeError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, mapRuntimeToResponse(runtime))
}

func (h *InternalSessionHandler) DeleteSessionRuntime(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "delete session runtime is not implemented yet")
}

func mapRuntimeToResponse(runtime domain.SessionRuntime) dto.SessionRuntimeResponse {
	return dto.SessionRuntimeResponse{
		SessionID:     runtime.SessionID,
		RoomCode:      runtime.RoomCode,
		Status:        string(runtime.Status),
		InitializedAt: runtime.InitializedAt,
	}
}
