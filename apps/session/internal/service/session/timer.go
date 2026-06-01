package session

import (
	"context"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

// HandleTick advances a session state machine when a timer deadline has expired.
func (s *Service) HandleTick(ctx context.Context, sessionID string, now time.Time) ([]domain.SessionDomainEvent, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return nil, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return nil, s.mapRedisError(err)
	}

	switch snapshot.Runtime.Status {
	case domain.RuntimeStatusLobby, domain.RuntimeStatusFinished:
		return nil, nil
	case domain.RuntimeStatusQuestionOpen:
		if snapshot.Runtime.Progress.DeadlineAt != nil && !snapshot.Runtime.Progress.DeadlineAt.After(now) {
			payload, err := s.CloseCurrentQuestionAndBuildReveal(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			return []domain.SessionDomainEvent{{SessionID: sessionID, Type: domain.EventAnswerRevealed, Payload: payload}}, nil
		}
	case domain.RuntimeStatusAnswerReveal:
		if snapshot.Runtime.Progress.RevealUntil != nil && !snapshot.Runtime.Progress.RevealUntil.After(now) {
			payload, err := s.AdvanceToLeaderboardReveal(ctx, sessionID)
			if err != nil {
				return nil, err
			}
			return []domain.SessionDomainEvent{{SessionID: sessionID, Type: domain.EventLeaderboardRevealed, Payload: payload}}, nil
		}
	case domain.RuntimeStatusLeaderboardReveal:
		if snapshot.Runtime.Progress.RevealUntil != nil && !snapshot.Runtime.Progress.RevealUntil.After(now) {
			payload, err := s.AdvanceAfterLeaderboardReveal(ctx, sessionID)
			if err != nil {
				return nil, err
			}

			eventType := domain.EventQuestionOpened
			if payload.Runtime.Status == domain.RuntimeStatusFinished {
				eventType = domain.EventGameFinished
			}

			return []domain.SessionDomainEvent{{SessionID: sessionID, Type: eventType, Payload: payload}}, nil
		}
	}

	return nil, nil
}
