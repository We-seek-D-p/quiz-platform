package redis

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoomCodeRepository_Reserve(t *testing.T) {
	t.Run("reserves unique code", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		ok, err := repo.Reserve(ctx, "12345678", "session-1")

		require.NoError(t, err)
		assert.True(t, ok)

		val, err := client.Get(ctx, "room_code:12345678").Result()
		require.NoError(t, err)
		assert.Equal(t, "session-1", val)
	})

	t.Run("returns false on duplicate code", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		ok1, err := repo.Reserve(ctx, "12345678", "session-1")
		require.NoError(t, err)
		assert.True(t, ok1)

		ok2, err := repo.Reserve(ctx, "12345678", "session-2")
		require.NoError(t, err)
		assert.False(t, ok2)

		val, err := client.Get(ctx, "room_code:12345678").Result()
		require.NoError(t, err)
		assert.Equal(t, "session-1", val)
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			Password:    "",
			DB:          0,
			MaxRetries:  1,
			DialTimeout: 100 * time.Millisecond,
		})
		defer func() { _ = client.Close() }()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		ok, err := repo.Reserve(ctx, "12345678", "session-1")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrRedisUnavailable)
		assert.False(t, ok)
	})
}

func TestRoomCodeRepository_Release(t *testing.T) {
	t.Run("releases existing room code key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		ok, err := repo.Reserve(ctx, "12345678", "session-1")
		require.NoError(t, err)
		assert.True(t, ok)

		exists, err := client.Exists(ctx, "room_code:12345678").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		err = repo.Release(ctx, "12345678")
		require.NoError(t, err)

		exists, err = client.Exists(ctx, "room_code:12345678").Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("is idempotent for missing key", func(t *testing.T) {
		mr, client := setupTestRedis(t)
		defer func() {
			mr.Close()
			_ = client.Close()
		}()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		err := repo.Release(ctx, "99999999")
		require.NoError(t, err)

		err = repo.Release(ctx, "99999999")
		require.NoError(t, err)
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		client := redis.NewClient(&redis.Options{
			Addr:        "localhost:9999",
			Password:    "",
			DB:          0,
			MaxRetries:  1,
			DialTimeout: 100 * time.Millisecond,
		})
		defer func() { _ = client.Close() }()

		repo := NewRoomCodeRepository(client)
		ctx := context.Background()

		err := repo.Release(ctx, "12345678")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrRedisUnavailable)
	})
}
