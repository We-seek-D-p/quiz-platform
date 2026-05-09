package management

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSessionBootstrap(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/internal/v1/sessions/test-123/bootstrap", r.URL.Path)
			assert.Equal(t, "GET", r.Method)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(BootstrapResponse{
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
			if err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("test-123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var dto BootstrapResponse
		err = json.NewDecoder(resp.Body).Decode(&dto)
		require.NoError(t, err)

		result := repo.mapBootstrapToDomain(dto)

		assert.Equal(t, "test-123", result.SessionID)
		assert.Equal(t, "quiz-1", result.QuizID)
		assert.Equal(t, "host-1", result.HostID)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			err := json.NewEncoder(w).Encode(ErrorResponse{Code: "session_not_found"})
			if err != nil {
				t.Fatalf("failed to encode error response: %v", err)
			}
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("999"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("forbidden", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})
}

func TestReportSessionStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/internal/v1/sessions/123/status", r.URL.Path)
			assert.Equal(t, "PATCH", r.Method)

			var req ReportSessionStatusRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "in_progress", req.Status)
			assert.Equal(t, "event-123", req.EventID)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "in_progress",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
			err := json.NewEncoder(w).Encode(ErrorResponse{Code: "already_finished"})
			if err != nil {
				t.Fatalf("failed to encode error response: %v", err)
			}
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "finished",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "finished",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("999"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "finished",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrUnauthorized)
	})

	t.Run("forbidden", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusForbidden)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "finished",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrForbidden)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "finished",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})
}

func TestReportSessionResults(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/internal/v1/sessions/123/results", r.URL.Path)
			assert.Equal(t, "PUT", r.Method)

			var req ReportSessionResultsRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			assert.NoError(t, err)
			assert.Equal(t, "event-123", req.EventID)
			assert.Equal(t, "completed", req.FinishReason)

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionResultsRequest{
			EventID:      "event-123",
			FinishReason: "completed",
			FinishedAt:   time.Now(),
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPut, repo.resultsPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionResultsRequest{
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPut, repo.resultsPath("999"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("conflict", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusConflict)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionResultsRequest{
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPut, repo.resultsPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrAlreadyFinished)
	})

	t.Run("server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionResultsRequest{
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPut, repo.resultsPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrUpstreamUnavailable)
	})
}

func TestInternalHeaders(t *testing.T) {
	t.Run("sends internal headers", func(t *testing.T) {
		var receivedService, receivedToken string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedService = r.Header.Get("X-Internal-Service")
			receivedToken = r.Header.Get("X-Internal-Token")
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(BootstrapResponse{
				Session: BootstrapSessionDTO{SessionID: "123"},
			})
			if err != nil {
				t.Fatalf("failed to encode response: %v", err)
			}
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "secret-token",
			serviceName: "my-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodGet, repo.bootstrapPath("123"), nil)
		require.NoError(t, err)

		_, err = repo.do(req)
		assert.NoError(t, err)
		assert.Equal(t, "my-service", receivedService)
		assert.Equal(t, "secret-token", receivedToken)
	})

	t.Run("sends content-type for PATCH and PUT", func(t *testing.T) {
		var contentType string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType = r.Header.Get("Content-Type")
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		payload := ReportSessionStatusRequest{
			Status:  "lobby",
			EventID: "event-123",
		}

		req, err := testNewRequest(repo, context.Background(), http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		_, err = repo.do(req)
		assert.NoError(t, err)
		assert.Equal(t, "application/json", contentType)
	})
}
