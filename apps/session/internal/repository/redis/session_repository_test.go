package redis

import "testing"

func TestSessionRepository_Create(t *testing.T) {
	requireRedisTestEnv(t)

	t.Run("creates meta and quiz snapshot keys", func(t *testing.T) {
		t.Skip("TODO: assert session meta hash and quiz snapshot string are written")
	})

	t.Run("returns conflict for existing runtime", func(t *testing.T) {
		t.Skip("TODO: assert duplicate create returns ErrSessionConflict")
	})

	t.Run("maps redis unavailable error", func(t *testing.T) {
		t.Skip("TODO: assert redis transport errors map to ErrRedisUnavailable")
	})
}

func TestSessionRepository_Get(t *testing.T) {
	requireRedisTestEnv(t)

	t.Run("returns runtime from stored meta", func(t *testing.T) {
		t.Skip("TODO: assert runtime fields are decoded from redis hash")
	})

	t.Run("returns not found for missing runtime", func(t *testing.T) {
		t.Skip("TODO: assert missing session returns ErrSessionNotFound")
	})

	t.Run("maps parse errors from invalid timestamps", func(t *testing.T) {
		t.Skip("TODO: assert invalid initialized_at produces parse error")
	})
}

func TestSessionRepository_Delete(t *testing.T) {
	requireRedisTestEnv(t)

	t.Run("deletes meta and snapshot keys", func(t *testing.T) {
		t.Skip("TODO: assert session meta and quiz snapshot keys are deleted")
	})

	t.Run("deletes room code index when present", func(t *testing.T) {
		t.Skip("TODO: assert room_code key is deleted when room_code exists")
	})

	t.Run("is idempotent for missing runtime keys", func(t *testing.T) {
		t.Skip("TODO: assert deleting absent runtime does not fail")
	})
}
