package response

import (
	"encoding/json"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http/dto"
)

func JSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if value == nil {
		return
	}

	_ = json.NewEncoder(w).Encode(value)
}

func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, dto.ErrorResponse{
		Code:    code,
		Message: message,
	})
}
