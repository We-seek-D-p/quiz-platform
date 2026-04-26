package handler

import (
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

type InternalSessionHandler struct{}

func NewInternalSessionHandler() *InternalSessionHandler {
	return &InternalSessionHandler{}
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
