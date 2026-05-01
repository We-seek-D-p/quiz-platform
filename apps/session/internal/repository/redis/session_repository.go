package redis

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
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

func (r *SessionRepository) Create(ctx context.Context, runtime domain.SessionRuntime, quiz domain.QuizSnapshot) error {
	return ErrNotImplemented
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (domain.SessionRuntime, error) {
	return domain.SessionRuntime{}, ErrNotImplemented
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	return ErrNotImplemented
}
