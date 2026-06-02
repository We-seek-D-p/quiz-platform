package dto

import (
	"errors"
	"strings"
	"time"
)

type InitSessionRequest struct {
	QuizID    string    `json:"quiz_id"`
	HostID    string    `json:"host_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (r *InitSessionRequest) Validate() error {
	if strings.TrimSpace(r.QuizID) == "" {
		return errors.New("quiz_id is required")
	}
	if strings.TrimSpace(r.HostID) == "" {
		return errors.New("host_id is required")
	}
	if r.CreatedAt.IsZero() {
		return errors.New("created_at must be a valid non-zero timestamp")
	}
	return nil
}

type SessionRuntimeResponse struct {
	SessionID     string    `json:"session_id"`
	RoomCode      string    `json:"room_code"`
	Status        string    `json:"status"`
	InitializedAt time.Time `json:"initialized_at"`
}
