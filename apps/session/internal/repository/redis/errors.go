package redis

import "errors"

var (
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionConflict  = errors.New("session already exists")
	ErrRedisUnavailable = errors.New("redis service unavailable")
)
