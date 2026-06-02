package redis

import "github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"

var (
	ErrRedisUnavailable = domain.NewInternal("redis_unavailable", "redis service is unavailable", nil)
	ErrSessionNotFound  = domain.NewNotFound("session_not_found", "session not found in redis", nil)
	ErrSessionConflict  = domain.NewConflict("session_runtime_conflict", "session already exists", nil)
)

func errAnswerAlreadySubmitted(err error) error {
	return domain.NewConflict("answer_already_submitted", "answer already submitted", err)
}

func errAnswerNotFound(err error) error {
	return domain.NewNotFound("answer_not_found", "answer not found", err)
}

func errInvalidNickname(err error) error {
	return domain.NewInvalidInput("invalid_payload", "nickname length must be between 2 and 64 chars", err)
}

func errNicknameTaken(err error) error {
	return domain.NewConflict("nickname_taken", "nickname already taken", err)
}

func errParticipantConflict(err error) error {
	return domain.NewConflict("participant_conflict", "participant already exists", err)
}

func errParticipantNotFound(err error) error {
	return domain.NewNotFound("participant_not_found", "participant not found", err)
}

func errRoomNotFound(err error) error {
	return domain.NewNotFound("room_not_found", "room not found", err)
}

func errRuntimeStoreUnavailable(err error) error {
	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

func errSerializationFailure(message string, err error) error {
	return domain.NewInternal("internal_error", message, err)
}
