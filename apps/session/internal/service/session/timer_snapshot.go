package session

import (
	"context"
	"errors"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
)

// GetSessionSnapshotForTimer builds the current session view used by timer-side broadcasts.
func (s *Service) GetSessionSnapshotForTimer(ctx context.Context, sessionID string) (Snapshot, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return Snapshot{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		if errors.Is(err, redis.ErrSessionNotFound) {
			return Snapshot{}, domain.NewNotFound("session_runtime_not_found", "session runtime not found", err)
		}
		return Snapshot{}, domain.NewInternal("internal_error", "runtime storage unavailable", err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return Snapshot{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return Snapshot{}, err
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop), nil
}
