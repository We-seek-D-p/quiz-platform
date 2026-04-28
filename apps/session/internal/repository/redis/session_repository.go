package redis

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/models"
	goredis "github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *goredis.Client
}

func NewSessionRepository(client *goredis.Client) *SessionRepository {
	return &SessionRepository{
		client: client,
	}
}

func (r *SessionRepository) Create(ctx context.Context, runtime models.SessionRuntime, quiz models.QuizSnapshot) error {
	return ErrNotImplemented
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (models.SessionRuntime, error) {
	return models.SessionRuntime{}, ErrNotImplemented
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	return ErrNotImplemented
}
