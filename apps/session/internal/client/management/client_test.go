package management

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestClient(baseURL string) *Client {
	return newTestClientWithAttempts(baseURL, 3)
}

func newTestClientWithAttempts(baseURL string, retryAttempts int) *Client {
	return &Client{
		baseURL:       baseURL,
		token:         "test-token",
		serviceName:   "test-service",
		retryAttempts: retryAttempts,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		log:           slog.New(slog.DiscardHandler),
	}
}

func TestGetSessionBootstrap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "test-service", r.Header.Get("X-Internal-Service"))
			assert.Equal(t, "test-token", r.Header.Get("X-Internal-Token"))
			assert.Equal(t, "/internal/v1/sessions/test-123/bootstrap", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(BootstrapResponse{
				Session: BootstrapSessionDTO{
					SessionID: "test-123",
					QuizID:    "quiz-1",
					HostID:    "host-1",
					Status:    "lobby",
				},
				QuizSnapshot: QuizSnapshotDTO{
					Title: "Test Quiz",
				},
			})
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		result, err := repo.GetSessionBootstrap(context.Background(), "test-123")

		require.NoError(t, err)
		assert.Equal(t, "test-123", result.SessionID)
		assert.Equal(t, "quiz-1", result.QuizID)
		assert.Equal(t, "host-1", result.HostID)
		assert.Equal(t, "Test Quiz", result.Quiz.Title)
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"session": {`))
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		_, err := repo.GetSessionBootstrap(context.Background(), "test-123")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidResponse)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Code: "session_not_found"})
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		_, err := repo.GetSessionBootstrap(context.Background(), "999")
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		_, err := repo.GetSessionBootstrap(context.Background(), "123")
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("forbidden", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		_, err := repo.GetSessionBootstrap(context.Background(), "123")
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		_, err := repo.GetSessionBootstrap(context.Background(), "123")
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})

	t.Run("context cancelled", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := repo.GetSessionBootstrap(ctx, "test-123")
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_, err := repo.GetSessionBootstrap(ctx, "test-123")
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func testContextDeadline(t *testing.T, name string, fn func(ctx context.Context, repo *Client) error) {
	t.Run(name, func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := fn(ctx, repo)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestReportSessionStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var receivedBody ReportSessionStatusRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/internal/v1/sessions/123/status", r.URL.Path)
			assert.Equal(t, http.MethodPatch, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		update := domain.SessionStatusUpdate{
			Status:  "in_progress",
			EventID: "event-123",
		}

		err := repo.ReportSessionStatus(context.Background(), "123", update)

		require.NoError(t, err)
		assert.Equal(t, "in_progress", receivedBody.Status)
		assert.Equal(t, "event-123", receivedBody.EventID)
	})

	t.Run("success with 204 No Content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		update := domain.SessionStatusUpdate{
			Status:  "finished",
			EventID: "event-123",
		}

		err := repo.ReportSessionStatus(context.Background(), "123", update)
		require.NoError(t, err)
	})

	t.Run("conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Code: "already_finished"})
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		update := domain.SessionStatusUpdate{
			Status:  "finished",
			EventID: "event-123",
		}

		err := repo.ReportSessionStatus(context.Background(), "123", update)
		assert.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		update := domain.SessionStatusUpdate{
			Status:  "in_progress",
			EventID: "event-123",
		}

		err := repo.ReportSessionStatus(context.Background(), "999", update)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		update := domain.SessionStatusUpdate{
			Status:  "in_progress",
			EventID: "event-123",
		}

		err := repo.ReportSessionStatus(context.Background(), "123", update)
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})

	testContextDeadline(t, "context deadline exceeded", func(ctx context.Context, repo *Client) error {
		update := domain.SessionStatusUpdate{
			Status:  "in_progress",
			EventID: "event-123",
		}
		return repo.ReportSessionStatus(ctx, "123", update)
	})
}

func TestReportSessionResults(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var receivedBody ReportSessionResultsRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/internal/v1/sessions/123/results", r.URL.Path)
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			_ = json.NewDecoder(r.Body).Decode(&receivedBody)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
			FinishedAt:   time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		}

		err := repo.ReportSessionResults(context.Background(), "123", results)

		require.NoError(t, err)
		assert.Equal(t, "event-123", receivedBody.EventID)
		assert.Equal(t, "completed", receivedBody.FinishReason)
	})

	t.Run("success with 204 No Content", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
			FinishedAt:   time.Now(),
		}

		err := repo.ReportSessionResults(context.Background(), "123", results)
		require.NoError(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
		}

		err := repo.ReportSessionResults(context.Background(), "999", results)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
		}

		err := repo.ReportSessionResults(context.Background(), "123", results)
		assert.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := newTestClient(server.URL)

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
		}

		err := repo.ReportSessionResults(context.Background(), "123", results)
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})

	testContextDeadline(t, "context deadline exceeded", func(ctx context.Context, repo *Client) error {
		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
		}
		return repo.ReportSessionResults(ctx, "123", results)
	})
}

func TestNewRequestErrors(t *testing.T) {
	t.Run("returns error for invalid URL", func(t *testing.T) {
		repo := &Client{
			baseURL:     "http://invalid\x7furl",
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
			log:         slog.New(slog.DiscardHandler),
		}

		ctx := context.Background()

		_, err := repo.GetSessionBootstrap(ctx, "test-123")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse")

		update := domain.SessionStatusUpdate{
			Status:  "in_progress",
			EventID: "event-123",
		}
		err = repo.ReportSessionStatus(ctx, "test-123", update)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse")

		results := domain.SessionResults{
			EventID:      "event-123",
			FinishReason: "completed",
		}
		err = repo.ReportSessionResults(ctx, "test-123", results)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parse")
	})
}
