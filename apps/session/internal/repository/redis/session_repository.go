package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

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
	metaKey := sessionMetaKey(runtime.SessionID)
	snapshotKey := sessionQuizSnapshotKey(runtime.SessionID)

	exists, err := r.client.Exists(ctx, metaKey).Result()
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
	}
	if exists > 0 {
		return ErrSessionConflict
	}

	quizJSON, err := json.Marshal(quiz)
	if err != nil {
		return fmt.Errorf("marshal quiz error: %w", err)
	}

	pipe := r.client.TxPipeline()

	pipe.HSet(ctx, metaKey, map[string]any{
		"session_id":     runtime.SessionID,
		"quiz_id":        runtime.QuizID,
		"host_id":        runtime.HostID,
		"room_code":      runtime.RoomCode,
		"status":         string(runtime.Status),
		"initialized_at": runtime.InitializedAt.Format(time.RFC3339),
	})

	pipe.Set(ctx, snapshotKey, quizJSON, 0)

	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
	}

	return nil
}

func (r *SessionRepository) Get(ctx context.Context, sessionID string) (domain.SessionBootstrap, error) {
	metaKey := sessionMetaKey(sessionID)
	snapshotKey := sessionQuizSnapshotKey(sessionID)

	pipe := r.client.Pipeline()
	metaCmd := pipe.HGetAll(ctx, metaKey)
	snapshotCmd := pipe.Get(ctx, snapshotKey)

	_, err := pipe.Exec(ctx)
	if err != nil && !errors.Is(err, goredis.Nil) {
		return domain.SessionBootstrap{}, fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
	}

	res, err := metaCmd.Result()
	if err != nil || len(res) == 0 {
		return domain.SessionBootstrap{}, ErrSessionNotFound
	}

	initTime, _ := time.Parse(time.RFC3339, res["initialized_at"])

	runtime := domain.SessionRuntime{
		SessionID:     res["session_id"],
		QuizID:        res["quiz_id"],
		HostID:        res["host_id"],
		RoomCode:      res["room_code"],
		Status:        domain.RuntimeStatus(res["status"]),
		InitializedAt: initTime,
	}

	snapshotData, err := snapshotCmd.Result()
	if err != nil {
		return domain.SessionBootstrap{}, fmt.Errorf("quiz snapshot integrity error: %w", err)
	}

	var quiz domain.QuizSnapshot
	if err := json.Unmarshal([]byte(snapshotData), &quiz); err != nil {
		return domain.SessionBootstrap{}, fmt.Errorf("failed to unmarshal quiz: %w", err)
	}

	return domain.SessionBootstrap{
		SessionID: runtime.SessionID,
		QuizID:    runtime.QuizID,
		HostID:    runtime.HostID,
		Status:    string(runtime.Status),
		Quiz:      quiz,
	}, nil
}

func (r *SessionRepository) Delete(ctx context.Context, sessionID string) error {
	metaKey := sessionMetaKey(sessionID)
	snapshotKey := sessionQuizSnapshotKey(sessionID)

	pipe := r.client.TxPipeline()

	pipe.Del(ctx, metaKey)
	pipe.Del(ctx, snapshotKey)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrRedisUnavailable, err)
	}

	return nil
}
