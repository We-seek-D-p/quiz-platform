package session

import "time"

type InitSessionParams struct {
	SessionID      string
	QuizID         string
	HostID         string
	CreatedAt      time.Time
	IdempotencyKey string
}

type GetSessionRuntimeParams struct {
	SessionID string
}

type DeleteSessionRuntimeParams struct {
	SessionID string
}
