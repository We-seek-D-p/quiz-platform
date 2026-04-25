package logger

import (
	"log/slog"
	"os"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/config"
)

func New(cfg *config.Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.Logger.Level),
	}

	var handler slog.Handler

	switch strings.ToLower(strings.TrimSpace(cfg.Logger.Format)) {
	case "text":
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	}

	return slog.New(handler).With(
		"service", cfg.App.Name,
	)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
