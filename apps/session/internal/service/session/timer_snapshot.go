package session

import (
	"context"
	"errors"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
)

func (s *Service) GetSessionSnapshotForTimer(ctx context.Context, sessionID string) (SnapshotDTO, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return SnapshotDTO{}, ErrInvalidParams
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		if errors.Is(err, redis.ErrSessionNotFound) {
			return SnapshotDTO{}, ErrSessionRuntimeNotFound
		}
		return SnapshotDTO{}, ErrRuntimeStoreUnavailable
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return SnapshotDTO{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return SnapshotDTO{}, err
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop), nil
}
