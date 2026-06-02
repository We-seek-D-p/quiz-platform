package session

import (
	"errors"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/client/management"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

// mapManagementError converts Management client errors into application errors.
func (s *Service) mapManagementError(err error) error {
	if _, ok := errors.AsType[*domain.AppError](err); ok {
		return err
	}
	if errors.Is(err, management.ErrSessionNotFound) {
		return domain.NewNotFound("session_not_found", "session not found", err)
	}
	if errors.Is(err, management.ErrAlreadyFinished) {
		return domain.NewConflict("already_finished", "session already finished", err)
	}
	return domain.NewInternal("bootstrap_fetch_failed", "failed to fetch bootstrap data", err)
}

func isAppErrorCode(err error, code string) bool {
	if appErr, ok := errors.AsType[*domain.AppError](err); ok {
		return appErr.Code == code
	}
	return false
}
