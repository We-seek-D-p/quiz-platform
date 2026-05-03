package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type SessionRepository struct {
	client *goredis.Client
}

func NewSessionRepository(client *goredis.Client) *SessionRepository {
	return &SessionRepository{client: client}
}

func (r *SessionRepository) Create(ctx context.Context, runtime domain.SessionRuntime, quiz domain.QuizSnapshot) error {
	metaKey := sessionMetaKey(runtime.SessionID)
	snapshotKey := sessionQuizSnapshotKey(runtime.SessionID)

	exists, err := r.client.Exists(ctx, metaKey).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if exists > 0 {
		return ErrSessionConflict
	}

	quizJSON, err := json.Marshal(quiz)
	if err != nil {
		return fmt.Errorf("marshal quiz: %w", err)
	}

	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, metaKey, map[string]any{
		"session_id":     runtime.SessionID,
		"quiz_id":        runtime.QuizID,
		"host_id":        runtime.HostID,
		"room_code":      runtime.RoomCode,
		"status":         string(runtime.Status),
		"initialized_at": runtime.InitializedAt.UTC().Format(time.RFC3339Nano),
	})
	pipe.Set(ctx, snapshotKey, quizJSON, 0)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (domain.SessionRuntime, error) {
	metaKey := sessionMetaKey(sessionID)

	meta, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if len(meta) == 0 {
		return domain.SessionRuntime{}, ErrSessionNotFound
	}

	initializedAt, err := time.Parse(time.RFC3339, meta["initialized_at"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse initialized_at: %w", err)
	}

	return domain.SessionRuntime{
		SessionID:     meta["session_id"],
		QuizID:        meta["quiz_id"],
		HostID:        meta["host_id"],
		RoomCode:      meta["room_code"],
		Status:        domain.RuntimeStatus(meta["status"]),
		InitializedAt: initializedAt,
	}, nil
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	metaKey := sessionMetaKey(sessionID)
	snapshotKey := sessionQuizSnapshotKey(sessionID)

	meta, err := r.client.HGetAll(ctx, metaKey).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	roomCode := meta["room_code"]

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, metaKey)
	pipe.Del(ctx, snapshotKey)
	if roomCode != "" {
		pipe.Del(ctx, roomCodeKey(roomCode))
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}
