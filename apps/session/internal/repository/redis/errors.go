package redis

import "errors"

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrRoomNotFound        = errors.New("room not found")
	ErrNicknameTaken       = errors.New("nickname already taken")
	ErrParticipantNotFound = errors.New("participant not found")
	ErrParticipantConflict = errors.New("participant already exists")
	ErrSessionConflict     = errors.New("session already exists")
	ErrRedisUnavailable    = errors.New("redis service unavailable")
	ErrInvalidNickname     = errors.New("invalid nickname")
)
