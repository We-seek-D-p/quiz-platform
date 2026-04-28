package redis

import (
	"context"

	goredis "github.com/redis/go-redis/v9"
)

type RoomCodeRepository struct {
	client *goredis.Client
}

func NewRoomCodeRepository(client *goredis.Client) *RoomCodeRepository {
	return &RoomCodeRepository{
		client: client,
	}
}

func (r *RoomCodeRepository) Reserve(ctx context.Context, roomCode string, sessionID string) (bool, error) {
	return false, ErrNotImplemented
}

func (r *RoomCodeRepository) Release(ctx context.Context, roomCode string) error {
	return ErrNotImplemented
}
