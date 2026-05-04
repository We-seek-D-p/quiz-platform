package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
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
	totalQuestions := runtime.Progress.TotalQuestions
	if totalQuestions == 0 {
		totalQuestions = len(quiz.Questions)
	}

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
		"session_id":             runtime.SessionID,
		"quiz_id":                runtime.QuizID,
		"host_id":                runtime.HostID,
		"room_code":              runtime.RoomCode,
		"status":                 string(runtime.Status),
		"initialized_at":         runtime.InitializedAt.UTC().Format(time.RFC3339Nano),
		"current_question_index": runtime.Progress.CurrentQuestionIndex,
		"total_questions":        totalQuestions,
		"started_at":             formatOptionalTime(runtime.Progress.StartedAt),
		"finished_at":            formatOptionalTime(runtime.Progress.FinishedAt),
		"deadline_at":            formatOptionalTime(runtime.Progress.DeadlineAt),
		"reveal_until":           formatOptionalTime(runtime.Progress.RevealUntil),
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

	initializedAt, err := time.Parse(time.RFC3339Nano, meta["initialized_at"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse initialized_at: %w", err)
	}

	currentQuestionIndex, err := parseOptionalInt(meta["current_question_index"], -1)
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse current_question_index: %w", err)
	}

	totalQuestions, err := parseOptionalInt(meta["total_questions"], 0)
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse total_questions: %w", err)
	}

	startedAt, err := parseOptionalTime(meta["started_at"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse started_at: %w", err)
	}

	finishedAt, err := parseOptionalTime(meta["finished_at"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse finished_at: %w", err)
	}

	deadlineAt, err := parseOptionalTime(meta["deadline_at"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse deadline_at: %w", err)
	}

	revealUntil, err := parseOptionalTime(meta["reveal_until"])
	if err != nil {
		return domain.SessionRuntime{}, fmt.Errorf("parse reveal_until: %w", err)
	}

	return domain.SessionRuntime{
		SessionID:     meta["session_id"],
		QuizID:        meta["quiz_id"],
		HostID:        meta["host_id"],
		RoomCode:      meta["room_code"],
		Status:        domain.RuntimeStatus(meta["status"]),
		InitializedAt: initializedAt,
		Progress: domain.RuntimeProgress{
			CurrentQuestionIndex: currentQuestionIndex,
			TotalQuestions:       totalQuestions,
			StartedAt:            startedAt,
			FinishedAt:           finishedAt,
			DeadlineAt:           deadlineAt,
			RevealUntil:          revealUntil,
		},
	}, nil
}

func (r *SessionRepository) GetSnapshot(ctx context.Context, sessionID string) (domain.SessionSnapshot, error) {
	runtime, err := r.Get(ctx, sessionID)
	if err != nil {
		return domain.SessionSnapshot{}, err
	}

	snapshotJSON, err := r.client.Get(ctx, sessionQuizSnapshotKey(sessionID)).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.SessionSnapshot{}, ErrSessionNotFound
		}

		return domain.SessionSnapshot{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	var quiz domain.QuizSnapshot
	if err := json.Unmarshal([]byte(snapshotJSON), &quiz); err != nil {
		return domain.SessionSnapshot{}, fmt.Errorf("unmarshal quiz snapshot: %w", err)
	}

	return domain.SessionSnapshot{Runtime: runtime, Quiz: quiz}, nil
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

func (r *SessionRepository) UpdateRuntime(ctx context.Context, runtime domain.SessionRuntime) error {
	metaKey := sessionMetaKey(runtime.SessionID)

	exists, err := r.client.Exists(ctx, metaKey).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if exists == 0 {
		return ErrSessionNotFound
	}

	update := map[string]any{
		"session_id":             runtime.SessionID,
		"quiz_id":                runtime.QuizID,
		"host_id":                runtime.HostID,
		"room_code":              runtime.RoomCode,
		"status":                 string(runtime.Status),
		"initialized_at":         runtime.InitializedAt.UTC().Format(time.RFC3339Nano),
		"current_question_index": runtime.Progress.CurrentQuestionIndex,
		"total_questions":        runtime.Progress.TotalQuestions,
		"started_at":             formatOptionalTime(runtime.Progress.StartedAt),
		"finished_at":            formatOptionalTime(runtime.Progress.FinishedAt),
		"deadline_at":            formatOptionalTime(runtime.Progress.DeadlineAt),
		"reveal_until":           formatOptionalTime(runtime.Progress.RevealUntil),
	}

	if err := r.client.HSet(ctx, metaKey, update).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}

func (r *SessionRepository) SetStatusAndProgress(
	ctx context.Context,
	sessionID string,
	status domain.RuntimeStatus,
	progress domain.RuntimeProgress,
) error {
	metaKey := sessionMetaKey(sessionID)

	exists, err := r.client.Exists(ctx, metaKey).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if exists == 0 {
		return ErrSessionNotFound
	}

	update := map[string]any{
		"status":                 string(status),
		"current_question_index": progress.CurrentQuestionIndex,
		"total_questions":        progress.TotalQuestions,
		"started_at":             formatOptionalTime(progress.StartedAt),
		"finished_at":            formatOptionalTime(progress.FinishedAt),
		"deadline_at":            formatOptionalTime(progress.DeadlineAt),
		"reveal_until":           formatOptionalTime(progress.RevealUntil),
	}

	if err := r.client.HSet(ctx, metaKey, update).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}

	return value.UTC().Format(time.RFC3339Nano)
}

func parseOptionalTime(value string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}

	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func parseOptionalInt(value string, defaultValue int) (int, error) {
	if value == "" {
		return defaultValue, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}
