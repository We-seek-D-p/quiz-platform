package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
	httptransport "github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/http"
)

type App struct {
	cfg    *config.Config
	log    *slog.Logger
	server *httptransport.Server
}

func New(cfg *config.Config, log *slog.Logger) *App {
	router := httptransport.NewRouter(log)
	server := httptransport.NewServer(cfg.HTTP.Address(), router)

	return &App{
		cfg:    cfg,
		log:    log,
		server: server,
	}
}

func (a *App) Run(ctx context.Context) error {
	serverErrCh := make(chan error, 1)

	a.log.Info("starting http server", "addr", a.cfg.HTTP.Address())
	go func() {
		serverErrCh <- a.server.Run()
	}()

	select {
	case <-ctx.Done():
		a.log.Info("shutting down server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown http server: %w", err)
		}
		a.log.Info("http server stopped")

		return nil

	case err := <-serverErrCh:
		if errors.Is(err, http.ErrServerClosed) {
			a.log.Info("http server stopped")
			return nil
		}

		return fmt.Errorf("run http server: %w", err)
	}
}
