package redis

import (
	"context"
	"errors"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

type RoomCodeRepository struct {
	client *goredis.Client
}

func NewRoomCodeRepository(client *goredis.Client) *RoomCodeRepository {
	return &RoomCodeRepository{client: client}
}

func (r *RoomCodeRepository) Reserve(ctx context.Context, roomCode, sessionID string) (bool, error) {
	key := roomCodeKey(roomCode)

	ok, err := r.client.SetNX(ctx, key, sessionID, 0).Result()
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return ok, nil
}

func (r *RoomCodeRepository) GetSessionID(ctx context.Context, roomCode string) (string, error) {
	key := roomCodeKey(roomCode)

	sessionID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return "", ErrRoomNotFound
		}

		return "", fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return sessionID, nil
}

func (r *RoomCodeRepository) Release(ctx context.Context, roomCode string) error {
	key := roomCodeKey(roomCode)

	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}
