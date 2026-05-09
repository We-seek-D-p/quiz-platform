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
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
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

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(context.Background(), repo, http.MethodGet, repo.bootstrapPath("test-123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		var dto BootstrapResponse
		err = json.NewDecoder(resp.Body).Decode(&dto)
		require.NoError(t, err)

		result := repo.mapBootstrapToDomain(dto)

		assert.Equal(t, "test-123", result.SessionID)
		assert.Equal(t, "quiz-1", result.QuizID)
		assert.Equal(t, "host-1", result.HostID)
	})

	t.Run("not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Code: "session_not_found"})
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "test-token",
			serviceName: "test-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(context.Background(), repo, http.MethodGet, repo.bootstrapPath("999"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		err = repo.handleError(resp)
		assert.ErrorIs(t, err, ErrSessionNotFound)
	})

	t.Run("unauthorized", func(t *testing.T) {
		testHTTPError(t, http.StatusUnauthorized, ErrUnauthorized, repoBootstrapCall)
	})

	t.Run("forbidden", func(t *testing.T) {
		testHTTPError(t, http.StatusForbidden, ErrForbidden, repoBootstrapCall)
	})

	t.Run("server error", func(t *testing.T) {
		testHTTPError(t, http.StatusInternalServerError, ErrUpstreamUnavailable, repoBootstrapCall)
	})
}

func TestReportSessionStatus(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := newTestServer(t, "status", http.StatusOK)
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

		req, err := testNewRequest(context.Background(), repo, http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("conflict", func(t *testing.T) {
		testStatusError(t, http.StatusConflict, ErrAlreadyFinished, repoStatusCall, "123")
	})

	t.Run("not found", func(t *testing.T) {
		testStatusError(t, http.StatusNotFound, ErrSessionNotFound, repoStatusCall, "999")
	})

	t.Run("unauthorized", func(t *testing.T) {
		testStatusError(t, http.StatusUnauthorized, ErrUnauthorized, repoStatusCall, "123")
	})

	t.Run("forbidden", func(t *testing.T) {
		testStatusError(t, http.StatusForbidden, ErrForbidden, repoStatusCall, "123")
	})

	t.Run("server error", func(t *testing.T) {
		testStatusError(t, http.StatusInternalServerError, ErrUpstreamUnavailable, repoStatusCall, "123")
	})
}

func TestReportSessionResults(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := newTestServer(t, "results", http.StatusOK)
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

		req, err := testNewRequest(context.Background(), repo, http.MethodPut, repo.resultsPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("not found", func(t *testing.T) {
		testResultsError(t, http.StatusNotFound, ErrSessionNotFound, repoResultsCall, "999")
	})

	t.Run("conflict", func(t *testing.T) {
		testResultsError(t, http.StatusConflict, ErrAlreadyFinished, repoResultsCall, "123")
	})

	t.Run("server error", func(t *testing.T) {
		testResultsError(t, http.StatusInternalServerError, ErrUpstreamUnavailable, repoResultsCall, "123")
	})
}

func TestInternalHeaders(t *testing.T) {
	t.Run("sends internal headers", func(t *testing.T) {
		var receivedService, receivedToken string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			receivedService = r.Header.Get("X-Internal-Service")
			receivedToken = r.Header.Get("X-Internal-Token")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(BootstrapResponse{
				Session: BootstrapSessionDTO{SessionID: "123"},
			})
		}))
		defer server.Close()

		repo := &Repository{
			baseURL:     server.URL,
			token:       "secret-token",
			serviceName: "my-service",
			httpClient:  &http.Client{Timeout: 5 * time.Second},
		}

		req, err := testNewRequest(context.Background(), repo, http.MethodGet, repo.bootstrapPath("123"), nil)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
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

		req, err := testNewRequest(context.Background(), repo, http.MethodPatch, repo.statusPath("123"), payload)
		require.NoError(t, err)

		resp, err := repo.do(req)
		require.NoError(t, err)
		defer func() { _ = resp.Body.Close() }()
		assert.Equal(t, "application/json", contentType)
	})
}

type repoCallFunc func(*Repository, string) (*http.Response, error)

func repoBootstrapCall(repo *Repository, sessionID string) (*http.Response, error) {
	req, err := testNewRequest(context.Background(), repo, http.MethodGet, repo.bootstrapPath(sessionID), nil)
	if err != nil {
		return nil, err
	}
	return repo.do(req)
}

func repoStatusCall(repo *Repository, sessionID string) (*http.Response, error) {
	payload := ReportSessionStatusRequest{
		Status:  "finished",
		EventID: "event-123",
	}
	req, err := testNewRequest(context.Background(), repo, http.MethodPatch, repo.statusPath(sessionID), payload)
	if err != nil {
		return nil, err
	}
	return repo.do(req)
}

func repoResultsCall(repo *Repository, sessionID string) (*http.Response, error) {
	payload := ReportSessionResultsRequest{
		EventID: "event-123",
	}
	req, err := testNewRequest(context.Background(), repo, http.MethodPut, repo.resultsPath(sessionID), payload)
	if err != nil {
		return nil, err
	}
	return repo.do(req)
}

// Generic test server for status and results endpoints
type endpointConfig struct {
	path   string
	method string
}

func newTestServer(t *testing.T, endpointType string, statusCode int) *httptest.Server {
	t.Helper()

	var config endpointConfig
	if endpointType == "status" {
		config = endpointConfig{
			path:   "/internal/v1/sessions/123/status",
			method: http.MethodPatch,
		}
	} else {
		config = endpointConfig{
			path:   "/internal/v1/sessions/123/results",
			method: http.MethodPut,
		}
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != config.path {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.Method != config.method {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if statusCode == http.StatusOK {
			if endpointType == "status" {
				var req ReportSessionStatusRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if req.Status != "in_progress" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if req.EventID != "event-123" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			} else {
				var req ReportSessionResultsRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if req.EventID != "event-123" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if req.FinishReason != "completed" {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
			}
		}

		w.WriteHeader(statusCode)
	}))
}

func testHTTPError(t *testing.T, statusCode int, expectedErr error, call repoCallFunc) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(statusCode)
	}))
	defer server.Close()

	repo := &Repository{
		baseURL:     server.URL,
		token:       "test-token",
		serviceName: "test-service",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := call(repo, "123")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	err = repo.handleError(resp)
	assert.ErrorIs(t, err, expectedErr)
}

func testStatusError(t *testing.T, statusCode int, expectedErr error, call repoCallFunc, sessionID string) {
	server := newTestServer(t, "status", statusCode)
	defer server.Close()

	repo := &Repository{
		baseURL:     server.URL,
		token:       "test-token",
		serviceName: "test-service",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := call(repo, sessionID)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	err = repo.handleError(resp)
	assert.ErrorIs(t, err, expectedErr)
}

func testResultsError(t *testing.T, statusCode int, expectedErr error, call repoCallFunc, sessionID string) {
	server := newTestServer(t, "results", statusCode)
	defer server.Close()

	repo := &Repository{
		baseURL:     server.URL,
		token:       "test-token",
		serviceName: "test-service",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
	}

	resp, err := call(repo, sessionID)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	err = repo.handleError(resp)
	assert.ErrorIs(t, err, expectedErr)
}
