package session

import (
	"errors"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/client/management"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
)

// mapManagementError converts Management client errors into application errors.
func (s *Service) mapManagementError(err error) error {
	if errors.Is(err, management.ErrSessionNotFound) {
		return domain.NewNotFound("session_not_found", "session not found", err)
	}
	if errors.Is(err, management.ErrAlreadyFinished) {
		return domain.NewConflict("already_finished", "session already finished", err)
	}
	return domain.NewInternal("bootstrap_fetch_failed", "failed to fetch bootstrap data", err)
}

// mapRedisError converts session runtime repository errors into application errors.
func (s *Service) mapRedisError(err error) error {
	if errors.Is(err, redis.ErrSessionNotFound) {
		return domain.NewNotFound("session_runtime_not_found", "session runtime not found", err)
	}
	if errors.Is(err, redis.ErrSessionConflict) {
		return domain.NewConflict("session_runtime_conflict", "session runtime conflict", err)
	}
	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

// mapRoomCodeError converts room code repository errors into application errors.
func (s *Service) mapRoomCodeError(err error) error {
	if errors.Is(err, redis.ErrRoomNotFound) {
		return domain.NewNotFound("room_not_found", "room not found", err)
	}

	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

// mapParticipantRepositoryError converts participant repository errors into application errors.
func (s *Service) mapParticipantRepositoryError(err error) error {
	if errors.Is(err, redis.ErrNicknameTaken) {
		return domain.NewConflict("nickname_taken", "nickname already taken", err)
	}
	if errors.Is(err, redis.ErrParticipantNotFound) {
		return domain.NewNotFound("participant_not_found", "participant not found", err)
	}
	if errors.Is(err, redis.ErrInvalidNickname) {
		return domain.NewInvalidInput("invalid_payload", "invalid payload", err)
	}

	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

// mapAnswerRepositoryError converts answer repository errors into application errors.
func (s *Service) mapAnswerRepositoryError(err error) error {
	if errors.Is(err, redis.ErrAnswerAlreadySubmitted) {
		return domain.NewConflict("answer_already_submitted", "answer already submitted", err)
	}

	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

// mapLeaderboardRepositoryError converts leaderboard repository errors into application errors.
func (s *Service) mapLeaderboardRepositoryError(err error) error {
	if errors.Is(err, redis.ErrLeaderboardEntryNotFound) {
		return domain.NewNotFound("participant_not_found", "participant not found", err)
	}

	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}
