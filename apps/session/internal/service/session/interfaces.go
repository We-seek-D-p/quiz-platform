package session

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/models"
)

type ManagementClient interface {
	GetBootstrap(ctx context.Context, sessionID string) (models.SessionBootstrap, error)
}

type RuntimeRepository interface {
	Create(ctx context.Context, runtime models.SessionRuntime, quiz models.QuizSnapshot) error
	Get(ctx context.Context, sessionID string) (models.SessionRuntime, error)
	Delete(ctx context.Context, sessionID string) error
}

type RoomCodeRepository interface {
	Reserve(ctx context.Context, roomCode string, sessionID string) (bool, error)
	Release(ctx context.Context, roomCode string) error
}

type RoomCodeGenerator interface {
	Generate() string
}
