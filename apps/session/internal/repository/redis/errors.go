package redis

import "errors"

var (
	ErrNotImplemented   = errors.New("not implemented")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionConflict  = errors.New("session already exists")
	ErrRedisUnavailable = errors.New("redis service unavailable") // Ошибка доступности
)
