package redis

import "github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"

func errSessionRuntimeNotFound(err error) error {
	return domain.NewNotFound("session_runtime_not_found", "session runtime not found", err)
}

func errSessionRuntimeConflict(err error) error {
	return domain.NewConflict("session_runtime_conflict", "session runtime conflict", err)
}

func errRoomNotFound(err error) error {
	return domain.NewNotFound("room_not_found", "room not found", err)
}

func errNicknameTaken(err error) error {
	return domain.NewConflict("nickname_taken", "nickname already taken", err)
}

func errParticipantNotFound(err error) error {
	return domain.NewNotFound("participant_not_found", "participant not found", err)
}

func errParticipantConflict(err error) error {
	return domain.NewConflict("participant_conflict", "participant already exists", err)
}

func errAnswerAlreadySubmitted(err error) error {
	return domain.NewConflict("answer_already_submitted", "answer already submitted", err)
}

func errAnswerNotFound(err error) error {
	return domain.NewNotFound("answer_not_found", "answer not found", err)
}

func errInvalidNickname(err error) error {
	return domain.NewInvalidInput("invalid_payload", "nickname length must be between 2 and 64 chars", err)
}

func errRuntimeStoreUnavailable(err error) error {
	return domain.NewInternal("internal_error", "runtime storage unavailable", err)
}

func errSerializationFailure(message string, err error) error {
	return domain.NewInternal("internal_error", message, err)
}
