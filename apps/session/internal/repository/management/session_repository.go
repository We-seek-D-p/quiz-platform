package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type Repository struct {
	baseURL     string
	token       string
	serviceName string
	httpClient  *http.Client
}

func NewRepository(cfg *config.Config) *Repository {
	return &Repository{
		baseURL:     cfg.Management.BaseURL,
		token:       cfg.Management.InternalToken,
		serviceName: cfg.Internal.ServiceName,
		httpClient: &http.Client{
			Timeout: cfg.Management.Timeout(),
		},
	}
}

func (r *Repository) GetSessionBootstrap(ctx context.Context, sessionID string) (domain.SessionBootstrap, error) {
	req, err := r.newRequest(ctx, http.MethodGet, r.bootstrapPath(sessionID), nil)
	if err != nil {
		return domain.SessionBootstrap{}, err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return domain.SessionBootstrap{}, fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return domain.SessionBootstrap{}, r.handleError(resp)
	}

	var dto BootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		return domain.SessionBootstrap{}, fmt.Errorf("%w: %v", ErrInvalidResponse, err)
	}

	return r.mapBootstrapToDomain(dto), nil
}

func (r *Repository) ReportSessionStatus(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error {
	payload := ReportSessionStatusRequest{
		Status:    string(update.Status),
		StartedAt: update.StartedAt,
		EventID:   update.EventID,
	}

	req, err := r.newRequest(ctx, http.MethodPatch, r.statusPath(sessionID), payload)
	if err != nil {
		return err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return r.handleError(resp)
	}

	return nil
}

func (r *Repository) ReportSessionResults(ctx context.Context, sessionID string, results domain.SessionResults) error {
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
		FinishReason: results.FinishReason,
		FinishedAt:   results.FinishedAt,
		Participants: participants,
	}

	req, err := r.newRequest(ctx, http.MethodPut, r.resultsPath(sessionID), payload)
	if err != nil {
		return err
	}

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUpstreamUnavailable, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			return
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return r.handleError(resp)
	}

	return nil
}

func (r *Repository) bootstrapPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/bootstrap", internalAPIPrefix, url.PathEscape(sessionID))
}

func (r *Repository) statusPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/status", internalAPIPrefix, url.PathEscape(sessionID))
}

func (r *Repository) resultsPath(sessionID string) string {
	return fmt.Sprintf("%s/sessions/%s/results", internalAPIPrefix, url.PathEscape(sessionID))
}

func (r *Repository) newRequest(ctx context.Context, method string, path string, payload any) (*http.Request, error) {
	var body *bytes.Buffer

	if payload != nil {
		body = bytes.NewBuffer(nil)
		if err := json.NewEncoder(body).Encode(payload); err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, r.baseURL+path, body)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	req.Header.Set(internalServiceHeader, r.serviceName)
	req.Header.Set(internalTokenHeader, r.token)

	if payload != nil {
		req.Header.Set(contentTypeHeader, contentTypeJSON)
	}

	return req, nil
}

func (r *Repository) mapBootstrapToDomain(dto BootstrapResponse) domain.SessionBootstrap {
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
		Status:    dto.Session.Status,
		Quiz: domain.QuizSnapshot{
			Title:     dto.QuizSnapshot.Title,
			Questions: questions,
		},
	}
}

func (r *Repository) handleError(resp *http.Response) error {
	var errResp ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil {
		switch errResp.Code {
		case "session_not_found":
			return ErrSessionNotFound
		case "already_finished":
			return ErrAlreadyFinished
		}
	}

	switch resp.StatusCode {
	case http.StatusNotFound:
		return ErrSessionNotFound
	case http.StatusForbidden:
		return ErrForbidden
	case http.StatusUnauthorized:
		return ErrUnauthorized
	default:
		return fmt.Errorf("%w: status %d", ErrUnexpectedStatus, resp.StatusCode)
	}
}
