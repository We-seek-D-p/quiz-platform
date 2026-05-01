package management

import (
	"context"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

type Repository struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewRepository(cfg *config.Config) *Repository {
	return &Repository{
		baseURL: cfg.Management.BaseURL,
		token:   cfg.Management.InternalToken,
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
