package handler

import (
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

const idempotencyKeyHeader = "Idempotency-Key"

type InternalSessionHandler struct {
	service *session.Service
}

func NewInternalSessionHandler(service *session.Service) *InternalSessionHandler {
	return &InternalSessionHandler{service: service}
}

func (h *InternalSessionHandler) InitSession(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "init session is not implemented yet")
}

func (h *InternalSessionHandler) GetSessionRuntime(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "get session runtime is not implemented yet")
}

func (h *InternalSessionHandler) DeleteSessionRuntime(w http.ResponseWriter, _ *http.Request) {
	response.Error(w, http.StatusNotImplemented, "not_implemented", "delete session runtime is not implemented yet")
}
