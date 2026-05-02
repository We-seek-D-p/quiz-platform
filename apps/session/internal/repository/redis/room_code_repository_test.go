package redis

import "testing"

func TestRoomCodeRepository_Reserve(t *testing.T) {
	requireRedisTestEnv(t)

	t.Run("reserves unique code", func(t *testing.T) {
		t.Skip("TODO: assert first SetNX reserve succeeds")
	})

	t.Run("returns false on duplicate code", func(t *testing.T) {
		t.Skip("TODO: assert second reserve for same room code returns false")
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		t.Skip("TODO: assert redis transport errors map to ErrRedisUnavailable")
	})
}

func TestRoomCodeRepository_Release(t *testing.T) {
	requireRedisTestEnv(t)

	t.Run("releases existing room code key", func(t *testing.T) {
		t.Skip("TODO: assert room_code key is removed")
	})

	t.Run("is idempotent for missing key", func(t *testing.T) {
		t.Skip("TODO: assert deleting absent room code key does not fail")
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		t.Skip("TODO: assert redis transport errors map to ErrRedisUnavailable")
	})
}
