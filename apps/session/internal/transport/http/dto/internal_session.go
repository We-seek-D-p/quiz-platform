package dto

import "time"

type InitSessionRequest struct {
	QuizID    string    `json:"quiz_id"`
	HostID    string    `json:"host_id"`
	CreatedAt time.Time `json:"created_at"`
}

type SessionRuntimeResponse struct {
	SessionID     string    `json:"session_id"`
	RoomCode      string    `json:"room_code"`
	Status        string    `json:"status"`
	InitializedAt time.Time `json:"initialized_at"`
}
