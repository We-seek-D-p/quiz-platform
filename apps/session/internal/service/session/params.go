package session

import (
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

type InitSessionParams struct {
	SessionID      string
	QuizID         string
	HostID         string
	CreatedAt      time.Time
	IdempotencyKey string
}

type InitSessionResult struct {
	Runtime domain.SessionRuntime
	Created bool
}

type GetSessionRuntimeParams struct {
	SessionID string
}

type DeleteSessionRuntimeParams struct {
	SessionID string
}
