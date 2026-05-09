package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestRuntime(sessionID string) domain.SessionRuntime {
	return domain.SessionRuntime{
		SessionID:     sessionID,
		QuizID:        "quiz-123",
		HostID:        "host-456",
		RoomCode:      "12345678",
		Status:        domain.RuntimeStatusLobby,
		InitializedAt: time.Now().UTC(),
	}
}

func createTestQuizSnapshot() domain.QuizSnapshot {
	return domain.QuizSnapshot{
		Title: "Test Quiz",
		Questions: []domain.QuestionSnapshot{
			{
				ID:               "q1",
				Text:             "What is 2+2?",
				SelectionType:    domain.SelectionTypeSingle,
				TimeLimitSeconds: 30,
				OrderIndex:       0,
				Options: []domain.OptionSnapshot{
					{ID: "opt1", Text: "3", OrderIndex: 0, IsCorrect: false},
					{ID: "opt2", Text: "4", OrderIndex: 1, IsCorrect: true},
				},
			},
		},
	}
}

func TestSessionRepository_Create(t *testing.T) {
	t.Run("creates meta and quiz snapshot keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		metaKey := "session:session-1:meta"
		meta, err := client.HGetAll(ctx, metaKey).Result()
		require.NoError(t, err)

		assert.Equal(t, runtime.SessionID, meta["session_id"])
		assert.Equal(t, runtime.QuizID, meta["quiz_id"])
		assert.Equal(t, runtime.HostID, meta["host_id"])
		assert.Equal(t, runtime.RoomCode, meta["room_code"])
		assert.Equal(t, string(runtime.Status), meta["status"])

		snapshotKey := "session:session-1:quiz_snapshot"
		snapshotJSON, err := client.Get(ctx, snapshotKey).Result()
		require.NoError(t, err)

		var storedQuiz domain.QuizSnapshot
		err = json.Unmarshal([]byte(snapshotJSON), &storedQuiz)
		require.NoError(t, err)
		assert.Equal(t, quiz.Title, storedQuiz.Title)
		assert.Len(t, storedQuiz.Questions, 1)
		assert.Equal(t, "What is 2+2?", storedQuiz.Questions[0].Text)
	})

	t.Run("returns conflict for existing runtime", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		err = repo.Create(ctx, runtime, quiz)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionConflict)
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			Password:    "",
			DB:          0,
			MaxRetries:  1,
			DialTimeout: 100 * time.Millisecond,
		})
		defer client.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrRedisUnavailable)
	})
}

func TestSessionRepository_Get(t *testing.T) {
	t.Run("returns runtime from stored meta", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		expectedRuntime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, expectedRuntime, quiz)
		require.NoError(t, err)

		actualRuntime, err := repo.Get(ctx, "session-1")
		require.NoError(t, err)

		assert.Equal(t, expectedRuntime.SessionID, actualRuntime.SessionID)
		assert.Equal(t, expectedRuntime.QuizID, actualRuntime.QuizID)
		assert.Equal(t, expectedRuntime.HostID, actualRuntime.HostID)
		assert.Equal(t, expectedRuntime.RoomCode, actualRuntime.RoomCode)
		assert.Equal(t, expectedRuntime.Status, actualRuntime.Status)
		assert.WithinDuration(t, expectedRuntime.InitializedAt, actualRuntime.InitializedAt, time.Second)
	})

	t.Run("returns not found for missing runtime", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		_, err := repo.Get(ctx, "non-existent-session")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("maps parse errors from invalid timestamps", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		metaKey := "session:invalid-session:meta"
		err := client.HSet(ctx, metaKey, map[string]any{
			"session_id":     "invalid-session",
			"quiz_id":        "quiz-123",
			"host_id":        "host-456",
			"room_code":      "12345678",
			"status":         "lobby",
			"initialized_at": "not-a-valid-timestamp",
		}).Err()
		require.NoError(t, err)

		_, err = repo.Get(ctx, "invalid-session")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse initialized_at")
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Run("deletes meta and snapshot keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		metaKey := "session:session-1:meta"
		snapshotKey := "session:session-1:quiz_snapshot"

		exists, err := client.Exists(ctx, metaKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		exists, err = client.Exists(ctx, snapshotKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		err = repo.Delete(ctx, "session-1")
		require.NoError(t, err)

		exists, err = client.Exists(ctx, metaKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)

		exists, err = client.Exists(ctx, snapshotKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("deletes room code index when present", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		roomCodeRepo := NewRoomCodeRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime("session-1")
		quiz := createTestQuizSnapshot()

		reserved, err := roomCodeRepo.Reserve(ctx, runtime.RoomCode, runtime.SessionID)
		require.NoError(t, err)
		assert.True(t, reserved)

		err = repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		roomCodeKey := "room_code:12345678"
		exists, err := client.Exists(ctx, roomCodeKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		err = repo.Delete(ctx, "session-1")
		require.NoError(t, err)

		exists, err = client.Exists(ctx, roomCodeKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("is idempotent for missing runtime keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		err := repo.Delete(ctx, "non-existent-session")
		require.NoError(t, err)

		err = repo.Delete(ctx, "non-existent-session")
		require.NoError(t, err)
	})

	t.Run("handles session without room code", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer mr.Close()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := domain.SessionRuntime{
			SessionID:     "session-no-roomcode",
			QuizID:        "quiz-123",
			HostID:        "host-456",
			RoomCode:      "",
			Status:        domain.RuntimeStatusLobby,
			InitializedAt: time.Now().UTC(),
		}
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		err = repo.Delete(ctx, "session-no-roomcode")
		require.NoError(t, err)

		metaKey := "session:session-no-roomcode:meta"
		snapshotKey := "session:session-no-roomcode:quiz_snapshot"

		exists, err := client.Exists(ctx, metaKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)

		exists, err = client.Exists(ctx, snapshotKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})
}
