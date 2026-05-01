package management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	return domain.SessionBootstrap{}, ErrNotImplemented
}

func (r *Repository) ReportSessionStatus(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error {
	return ErrNotImplemented
}

func (r *Repository) ReportSessionResults(ctx context.Context, sessionID string, results domain.SessionResults) error {
	return ErrNotImplemented
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
