package management

import (
	"context"
	"net/http"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/models"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		baseURL: cfg.Management.BaseURL,
		token:   cfg.Internal.Token,
		httpClient: &http.Client{
			Timeout: cfg.Management.Timeout(),
		},
	}
}

func (c *Client) GetBootstrap(ctx context.Context, sessionID string) (models.SessionBootstrap, error) {
	return models.SessionBootstrap{}, ErrNotImplemented
}
