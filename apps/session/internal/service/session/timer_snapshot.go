package session

import (
	"context"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

// GetSessionSnapshotForTimer builds the current session view used by timer-side broadcasts.
func (s *Service) GetSessionSnapshotForTimer(ctx context.Context, sessionID string) (Snapshot, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return Snapshot{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return Snapshot{}, err
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return Snapshot{}, err
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return Snapshot{}, err
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop), nil
}
