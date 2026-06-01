package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

const (
	internalServiceHeader = "X-Internal-Service"
	internalTokenHeader   = "X-Internal-Token"
	contentTypeHeader     = "Content-Type"
	contentTypeJSON       = "application/json"
	internalAPIPrefix     = "/internal/v1"
)

// Client coordinates synchronous outbound HTTP orchestrations targeting Management.
type Client struct {
	baseURL       string
	token         string
	serviceName   string
	retryAttempts int
	httpClient    *http.Client
	log           *slog.Logger
}

func NewClient(cfg *config.Config, log *slog.Logger) *Client {
	if log == nil {
		panic("management client logger is required")
	}

	return &Client{
		baseURL:       cfg.Management.BaseURL,
		token:         cfg.Management.InternalToken,
		serviceName:   cfg.Internal.ServiceName,
		retryAttempts: cfg.Management.RetryAttempts,
		httpClient: &http.Client{
			Timeout: cfg.Management.Timeout(),
		},
		log: log,
	}
}

// GetSessionBootstrap fetches initial quiz questions parameters over secure network tubes.
func (c *Client) GetSessionBootstrap(ctx context.Context, sessionID string) (domain.SessionBootstrap, error) {
	resp, err := c.execute(ctx, "get_session_bootstrap", http.MethodGet, c.bootstrapPath(sessionID), nil, http.StatusOK)
	if err != nil {
		return domain.SessionBootstrap{}, err
	}

	var dto BootstrapResponse
	decodeErr := json.NewDecoder(resp.Body).Decode(&dto)
	closeErr := resp.Body.Close()

	if decodeErr != nil {
		c.log.ErrorContext(ctx, "management response decode failed", "operation", "get_session_bootstrap", "status", resp.StatusCode, "error", decodeErr)
		return domain.SessionBootstrap{}, fmt.Errorf("%w: %w", ErrInvalidResponse, decodeErr)
	}

	if closeErr != nil {
		c.log.ErrorContext(ctx, "management response body close failed", "operation", "get_session_bootstrap", "status", resp.StatusCode, "error", closeErr)
		return domain.SessionBootstrap{}, fmt.Errorf("close response body: %w", closeErr)
	}

	return c.mapBootstrapToDomain(dto), nil
}

// ReportSessionStatus patches upstream progress logs upon room lifecycle transition changes.
func (c *Client) ReportSessionStatus(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error {
	payload := ReportSessionStatusRequest{
		Status:    string(update.Status),
		StartedAt: update.StartedAt,
		EventID:   update.EventID,
	}

	resp, err := c.execute(ctx, "report_session_status", http.MethodPatch, c.statusPath(sessionID), payload, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return err
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.log.ErrorContext(ctx, "management response body close failed", "operation", "report_session_status", "status", resp.StatusCode, "error", closeErr)
		return fmt.Errorf("close response body: %w", closeErr)
	}

	return nil
}

// ReportSessionResults flushes sorted leaderboard tallies onto the core long-term storage ledger.
func (c *Client) ReportSessionResults(ctx context.Context, sessionID string, results domain.SessionResults) error {
	participants := make([]ReportSessionResultParticipant, len(results.Participants))
	for i, p := range results.Participants {
		participants[i] = ReportSessionResultParticipant{
			ParticipantID: p.ParticipantID,
			Nickname:      p.Nickname,
			Score:         p.Score,
			Rank:          p.Rank,
		}
	}

	payload := ReportSessionResultsRequest{
		EventID:      results.EventID,
		FinishReason: string(results.FinishReason),
		FinishedAt:   results.FinishedAt,
		Participants: participants,
	}

	resp, err := c.execute(ctx, "report_session_results", http.MethodPut, c.resultsPath(sessionID), payload, http.StatusOK, http.StatusNoContent)
	if err != nil {
		return err
	}
	if closeErr := resp.Body.Close(); closeErr != nil {
		c.log.ErrorContext(ctx, "management response body close failed", "operation", "report_session_results", "status", resp.StatusCode, "error", closeErr)
		return fmt.Errorf("close response body: %w", closeErr)
	}

	return nil
}

// execute manages the HTTP lifecycle wrap execution with centralized resilience mechanics.
func (c *Client) execute(ctx context.Context, operation, method, path string, payload any, expectedStatuses ...int) (*http.Response, error) {
	attempts := c.effectiveRetryAttempts()

	for attempt := 1; attempt <= attempts; attempt++ {
		req, err := c.newRequest(ctx, method, path, payload)
		if err != nil {
			c.log.ErrorContext(ctx, "management request build failed", "operation", operation, "method", method, "path", path, "attempt", attempt, "error", err)
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		c.log.DebugContext(ctx, "management request started", "operation", operation, "method", method, "path", path, "attempt", attempt)

		resp, err := c.do(req)
		if err != nil {
			if c.shouldRetry(ctx, attempt, attempts, 0, err) {
				c.log.WarnContext(ctx, "management request retrying", "operation", operation, "method", method, "path", path, "attempt", attempt, "error", err)
				continue
			}

			c.log.ErrorContext(ctx, "management request failed", "operation", operation, "method", method, "path", path, "attempt", attempt, "error", err)
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		if statusAllowed(resp.StatusCode, expectedStatuses) {
			c.log.InfoContext(ctx, "management request succeeded", "operation", operation, "method", method, "path", path, "status", resp.StatusCode, "attempt", attempt)
			return resp, nil
		}

		status := resp.StatusCode
		requestErr := c.handleError(resp)

		// Линеаризация закрытия Body избавляет от дублирования кода
		closeErr := resp.Body.Close()
		if closeErr != nil {
			c.log.ErrorContext(ctx, "management response body close failed", "operation", operation, "method", method, "path", path, "status", status, "attempt", attempt, "error", closeErr)
			return nil, fmt.Errorf("%s: %w; close response body: %w", operation, requestErr, closeErr)
		}

		if c.shouldRetry(ctx, attempt, attempts, status, requestErr) {
			c.log.WarnContext(ctx, "management request retrying", "operation", operation, "method", method, "path", path, "status", status, "attempt", attempt, "error", requestErr)
			continue
		}

		c.log.ErrorContext(ctx, "management request failed", "operation", operation, "method", method, "path", path, "status", status, "attempt", attempt, "error", requestErr)
		return nil, fmt.Errorf("%s: %w", operation, requestErr)
	}

	return nil, fmt.Errorf("%s: %w", operation, ErrUpstreamUnavailable)
}

func statusAllowed(status int, expectedStatuses []int) bool {
	for _, expected := range expectedStatuses {
		if status == expected {
			return true
		}
	}
	return false
}

func (c *Client) bootstrapPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/bootstrap", internalAPIPrefix, url.PathEscape(sessionID))
}

func (c *Client) statusPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/status", internalAPIPrefix, url.PathEscape(sessionID))
}

func (c *Client) resultsPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/results", internalAPIPrefix, url.PathEscape(sessionID))
}

func (c *Client) newRequest(ctx context.Context, method, path string, payload any) (*http.Request, error) {
	var body io.Reader

	if payload != nil {
		buffer := bytes.NewBuffer(nil)
		if err := json.NewEncoder(buffer).Encode(payload); err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
		body = buffer
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set(internalServiceHeader, c.serviceName)
	req.Header.Set(internalTokenHeader, c.token)

	if payload != nil {
		req.Header.Set(contentTypeHeader, contentTypeJSON)
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	resp, err := c.httpClient.Do(req) // #nosec G107 -- trusted internal service call to configured Management base URL
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrUpstreamUnavailable, err)
	}

	return resp, nil
}

func (c *Client) mapBootstrapToDomain(dto BootstrapResponse) domain.SessionBootstrap {
	questions := make([]domain.QuestionSnapshot, len(dto.QuizSnapshot.Questions))
	for i, q := range dto.QuizSnapshot.Questions {
		options := make([]domain.OptionSnapshot, len(q.Options))
		for j, o := range q.Options {
			options[j] = domain.OptionSnapshot{
				ID:         o.ID,
				Text:       o.Text,
				OrderIndex: o.OrderIndex,
				IsCorrect:  o.IsCorrect,
			}
		}
		questions[i] = domain.QuestionSnapshot{
			ID:               q.ID,
			Text:             q.Text,
			SelectionType:    domain.SelectionType(q.SelectionType),
			TimeLimitSeconds: q.TimeLimitSeconds,
			OrderIndex:       q.OrderIndex,
			Options:          options,
		}
	}

	return domain.SessionBootstrap{
		SessionID: dto.Session.SessionID,
		QuizID:    dto.Session.QuizID,
		HostID:    dto.Session.HostID,
		Status:    domain.PersistedStatus(dto.Session.Status),
		Quiz: domain.QuizSnapshot{
			Title:     dto.QuizSnapshot.Title,
			Questions: questions,
		},
	}
}

func (c *Client) handleError(resp *http.Response) error {
	var errResp ErrorResponse
	upstreamCode := ""
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		upstreamCode = errResp.Code
		switch errResp.Code {
		case "session_not_found":
			return fmt.Errorf("%w: status %d code %s", ErrSessionNotFound, resp.StatusCode, errResp.Code)
		case "already_finished":
			return fmt.Errorf("%w: status %d code %s", ErrAlreadyFinished, resp.StatusCode, errResp.Code)
		}
	}

	if resp.StatusCode >= http.StatusInternalServerError {
		return formatStatusError(ErrUpstreamUnavailable, resp.StatusCode, upstreamCode)
	}

	switch resp.StatusCode {
	case http.StatusConflict:
		return formatStatusError(ErrAlreadyFinished, resp.StatusCode, upstreamCode)
	case http.StatusNotFound:
		return formatStatusError(ErrSessionNotFound, resp.StatusCode, upstreamCode)
	case http.StatusForbidden:
		return formatStatusError(ErrForbidden, resp.StatusCode, upstreamCode)
	case http.StatusUnauthorized:
		return formatStatusError(ErrUnauthorized, resp.StatusCode, upstreamCode)
	default:
		return formatStatusError(ErrUnexpectedStatus, resp.StatusCode, upstreamCode)
	}
}

func formatStatusError(base error, status int, upstreamCode string) error {
	if upstreamCode != "" {
		return fmt.Errorf("%w: status %d code %s", base, status, upstreamCode)
	}
	return fmt.Errorf("%w: status %d", base, status)
}

func (c *Client) effectiveRetryAttempts() int {
	if c.retryAttempts < 1 {
		return 1
	}
	return c.retryAttempts
}

func (c *Client) shouldRetry(ctx context.Context, attempt, maxAttempts, status int, err error) bool {
	if attempt >= maxAttempts || ctx.Err() != nil {
		return false
	}
	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		return true
	}
	return status == 0 && err != nil
}
