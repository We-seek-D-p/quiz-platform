package handler

import (
	"errors"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/response"
)

func (h *InternalSessionHandler) handleInitSessionError(w http.ResponseWriter, err error) {
	handleAppError(w, err)
}

func (h *InternalSessionHandler) handleGetSessionRuntimeError(w http.ResponseWriter, err error) {
	handleAppError(w, err)
}

func (h *InternalSessionHandler) handleDeleteSessionRuntimeError(w http.ResponseWriter, err error) {
	handleAppError(w, err)
}

func handleAppError(w http.ResponseWriter, err error) {
	var appErr *domain.AppError
	if !errors.As(err, &appErr) {
		response.Error(w, http.StatusInternalServerError, "internal_error", "internal error")
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
