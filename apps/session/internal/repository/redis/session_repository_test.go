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

func createTestRuntime() domain.SessionRuntime {
	return domain.SessionRuntime{
		SessionID:     "session-1",
		QuizID:        "quiz-123",
		HostID:        "host-456",
		RoomCode:      "123456",
		Status:        domain.RuntimeStatusLobby,
		InitializedAt: time.Now().UTC(),
		Progress: domain.RuntimeProgress{
			CurrentQuestionIndex: 0,
			TotalQuestions:       5,
			StartedAt:            nil,
			FinishedAt:           nil,
			DeadlineAt:           nil,
			RevealUntil:          nil,
		},
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
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
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
		assert.Equal(t, "0", meta["current_question_index"])
		assert.Equal(t, "5", meta["total_questions"])

		snapshotKey := "session:session-1:quiz_snapshot"
		snapshotJSON, err := client.Get(ctx, snapshotKey).Result()
		require.NoError(t, err)

		var storedQuiz domain.QuizSnapshot
		err = json.Unmarshal([]byte(snapshotJSON), &storedQuiz)
		require.NoError(t, err)
		assert.Equal(t, quiz.Title, storedQuiz.Title)
		assert.Len(t, storedQuiz.Questions, 1)
	})

	t.Run("returns conflict for existing runtime", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		err = repo.Create(ctx, runtime, quiz)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "session_runtime_conflict")
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:0",
			Password:    "",
			DB:          0,
			MaxRetries:  1,
			DialTimeout: 100 * time.Millisecond,
		})
		defer func() { _ = client.Close() }()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)

		require.Error(t, err)
	})
}

func TestSessionRepository_Get(t *testing.T) {
	t.Run("returns runtime from stored meta", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		expectedRuntime := createTestRuntime()
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
		assert.Equal(t, expectedRuntime.Progress.CurrentQuestionIndex, actualRuntime.Progress.CurrentQuestionIndex)
		assert.Equal(t, expectedRuntime.Progress.TotalQuestions, actualRuntime.Progress.TotalQuestions)
		assert.WithinDuration(t, expectedRuntime.InitializedAt, actualRuntime.InitializedAt, time.Second)
	})

	t.Run("parses all time fields correctly", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		now := time.Now().UTC()
		runtime := createTestRuntime()
		runtime.Progress.StartedAt = &now
		runtime.Progress.FinishedAt = &now
		runtime.Progress.DeadlineAt = &now
		runtime.Progress.RevealUntil = &now

		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		actual, err := repo.Get(ctx, "session-1")
		require.NoError(t, err)

		assert.WithinDuration(t, *runtime.Progress.StartedAt, *actual.Progress.StartedAt, time.Second)
		assert.WithinDuration(t, *runtime.Progress.FinishedAt, *actual.Progress.FinishedAt, time.Second)
		assert.WithinDuration(t, *runtime.Progress.DeadlineAt, *actual.Progress.DeadlineAt, time.Second)
		assert.WithinDuration(t, *runtime.Progress.RevealUntil, *actual.Progress.RevealUntil, time.Second)
	})

	t.Run("returns not found for missing runtime", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		_, err := repo.Get(ctx, "non-existent-session")
		require.Error(t, err)
	})

	t.Run("returns error for invalid initialized_at", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		metaKey := "session:invalid-session:meta"
		err := client.HSet(ctx, metaKey, map[string]any{
			"session_id":             "invalid-session",
			"quiz_id":                "quiz-123",
			"host_id":                "host-456",
			"room_code":              "123456",
			"status":                 "lobby",
			"initialized_at":         "not-a-valid-timestamp",
			"current_question_index": "0",
			"total_questions":        "5",
		}).Err()
		require.NoError(t, err)

		_, err = repo.Get(ctx, "invalid-session")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})

	t.Run("handles missing optional fields", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		metaKey := "session:missing-fields:meta"
		err := client.HSet(ctx, metaKey, map[string]any{
			"session_id":     "missing-fields",
			"quiz_id":        "quiz-123",
			"host_id":        "host-456",
			"room_code":      "123456",
			"status":         "lobby",
			"initialized_at": time.Now().UTC().Format(time.RFC3339Nano),
		}).Err()
		require.NoError(t, err)

		snapshotKey := "session:missing-fields:quiz_snapshot"
		quiz := createTestQuizSnapshot()
		quizJSON, err := json.Marshal(quiz)
		require.NoError(t, err)
		err = client.Set(ctx, snapshotKey, quizJSON, 0).Err()
		require.NoError(t, err)

		runtime, err := repo.Get(ctx, "missing-fields")
		require.NoError(t, err)

		assert.Equal(t, -1, runtime.Progress.CurrentQuestionIndex)
		assert.Equal(t, 0, runtime.Progress.TotalQuestions)
		assert.Nil(t, runtime.Progress.StartedAt)
		assert.Nil(t, runtime.Progress.FinishedAt)
		assert.Nil(t, runtime.Progress.DeadlineAt)
		assert.Nil(t, runtime.Progress.RevealUntil)
	})
}

func TestSessionRepository_GetSnapshot(t *testing.T) {
	t.Run("returns quiz snapshot", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		expectedQuiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, expectedQuiz)
		require.NoError(t, err)

		snapshot, err := repo.GetSnapshot(ctx, "session-1")
		require.NoError(t, err)

		assert.Equal(t, expectedQuiz.Title, snapshot.Quiz.Title)
		assert.Len(t, snapshot.Quiz.Questions, 1)
		assert.Equal(t, runtime.SessionID, snapshot.Runtime.SessionID)
		assert.Equal(t, runtime.Status, snapshot.Runtime.Status)
		assert.Equal(t, runtime.Progress.CurrentQuestionIndex, snapshot.Runtime.Progress.CurrentQuestionIndex)
	})

	t.Run("returns not found for missing snapshot", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		_, err := repo.GetSnapshot(ctx, "non-existent")
		require.Error(t, err)
	})
}

func TestSessionRepository_SetStatusAndProgress(t *testing.T) {
	t.Run("updates status and progress fields", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		progress := domain.RuntimeProgress{
			CurrentQuestionIndex: 2,
			TotalQuestions:       5,
		}

		err = repo.SetStatusAndProgress(ctx, "session-1", "in_progress", progress)
		require.NoError(t, err)

		updated, err := repo.Get(ctx, "session-1")
		require.NoError(t, err)

		assert.Equal(t, domain.RuntimeStatus("in_progress"), updated.Status)
		assert.Equal(t, 2, updated.Progress.CurrentQuestionIndex)
		assert.Equal(t, 5, updated.Progress.TotalQuestions)
	})

	t.Run("returns not found for missing session", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		progress := domain.RuntimeProgress{
			CurrentQuestionIndex: 1,
			TotalQuestions:       5,
		}

		err := repo.SetStatusAndProgress(ctx, "non-existent", "in_progress", progress)
		require.Error(t, err)
	})
}

func TestSessionRepository_UpdateRuntime(t *testing.T) {
	t.Run("updates all runtime fields", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		now := time.Now().UTC()
		updatedRuntime := runtime
		updatedRuntime.Status = domain.RuntimeStatus("finished")
		updatedRuntime.Progress.CurrentQuestionIndex = 5
		updatedRuntime.Progress.TotalQuestions = 5
		updatedRuntime.Progress.FinishedAt = &now

		err = repo.UpdateRuntime(ctx, updatedRuntime)
		require.NoError(t, err)

		stored, err := repo.Get(ctx, "session-1")
		require.NoError(t, err)

		assert.Equal(t, domain.RuntimeStatus("finished"), stored.Status)
		assert.Equal(t, 5, stored.Progress.CurrentQuestionIndex)
		assert.WithinDuration(t, now, *stored.Progress.FinishedAt, time.Second)
	})

	t.Run("returns not found for missing session", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		runtime.SessionID = "non-existent"

		err := repo.UpdateRuntime(ctx, runtime)
		require.Error(t, err)
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	t.Run("deletes meta and snapshot keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		metaKey := "session:session-1:meta"
		snapshotKey := "session:session-1:quiz_snapshot"
		exists, _ := client.Exists(ctx, metaKey).Result()
		assert.Equal(t, int64(1), exists)
		exists, _ = client.Exists(ctx, snapshotKey).Result()
		assert.Equal(t, int64(1), exists)

		err = repo.Delete(ctx, "session-1")
		require.NoError(t, err)

		exists, _ = client.Exists(ctx, metaKey).Result()
		assert.Equal(t, int64(0), exists)
		exists, _ = client.Exists(ctx, snapshotKey).Result()
		assert.Equal(t, int64(0), exists)
	})

	t.Run("deletes room code index when present", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		roomCodeRepo := NewRoomCodeRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		reserved, err := roomCodeRepo.Reserve(ctx, runtime.RoomCode, runtime.SessionID)
		require.NoError(t, err)
		assert.True(t, reserved)

		err = repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		roomCodeKey := "room_code:123456"
		exists, err := client.Exists(ctx, roomCodeKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		err = repo.Delete(ctx, "session-1")
		require.NoError(t, err)

		exists, err = client.Exists(ctx, roomCodeKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("deletes additional session keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		runtime := createTestRuntime()
		quiz := createTestQuizSnapshot()

		err := repo.Create(ctx, runtime, quiz)
		require.NoError(t, err)

		participantsKey := "session:session-1:participants"
		leaderboardKey := "session:session-1:leaderboard"
		answersKey := "session:session-1:answers:q1"

		err = client.SAdd(ctx, participantsKey, "user1", "user2").Err()
		require.NoError(t, err)
		err = client.ZAdd(ctx, leaderboardKey, redis.Z{Score: 100, Member: "user1"}).Err()
		require.NoError(t, err)
		err = client.Set(ctx, answersKey, "some answer", 0).Err()
		require.NoError(t, err)

		exists, _ := client.Exists(ctx, participantsKey).Result()
		assert.Equal(t, int64(1), exists)
		exists, _ = client.Exists(ctx, leaderboardKey).Result()
		assert.Equal(t, int64(1), exists)
		exists, _ = client.Exists(ctx, answersKey).Result()
		assert.Equal(t, int64(1), exists)

		err = repo.Delete(ctx, "session-1")
		require.NoError(t, err)

		exists, _ = client.Exists(ctx, participantsKey).Result()
		assert.Equal(t, int64(0), exists)
		exists, _ = client.Exists(ctx, leaderboardKey).Result()
		assert.Equal(t, int64(0), exists)
		exists, _ = client.Exists(ctx, answersKey).Result()
		assert.Equal(t, int64(0), exists)
	})

	t.Run("is idempotent for missing runtime keys", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewSessionRepository(client)
		ctx := context.Background()

		err := repo.Delete(ctx, "non-existent-session")
		require.NoError(t, err)

		err = repo.Delete(ctx, "non-existent-session")
		require.NoError(t, err)
	})
}
