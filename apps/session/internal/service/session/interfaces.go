package session

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

type ManagementRepository interface {
	GetSessionBootstrap(ctx context.Context, sessionID string) (domain.SessionBootstrap, error)
	ReportSessionStatus(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error
	ReportSessionResults(ctx context.Context, sessionID string, results domain.SessionResults) error
}

type RuntimeRepository interface {
	Create(ctx context.Context, runtime domain.SessionRuntime, quiz domain.QuizSnapshot) error
	Get(ctx context.Context, sessionID string) (domain.SessionRuntime, error)
	GetSnapshot(ctx context.Context, sessionID string) (domain.SessionSnapshot, error)
	Delete(ctx context.Context, sessionID string) error
}

type RoomCodeRepository interface {
	Reserve(ctx context.Context, roomCode string, sessionID string) (bool, error)
	GetSessionID(ctx context.Context, roomCode string) (string, error)
	Release(ctx context.Context, roomCode string) error
}

type RoomCodeGenerator interface {
	Generate() string
}
