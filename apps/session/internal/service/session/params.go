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

type HostConnectParams struct {
	SessionID  string
	HostUserID string
}

type PlayerJoinParams struct {
	RoomCode string
	Nickname string
}

type PlayerReconnectParams struct {
	RoomCode         string
	ParticipantToken string
}

type StartGameParams struct {
	SessionID  string
	HostUserID string
}

type SubmitAnswerParams struct {
	SessionID         string
	ParticipantID     string
	QuestionID        string
	SelectedOptionIDs []string
}

type FinishGameParams struct {
	SessionID  string
	HostUserID string
}
