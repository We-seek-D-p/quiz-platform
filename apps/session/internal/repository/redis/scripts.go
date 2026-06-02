package redis

import goredis "github.com/redis/go-redis/v9"

const (
	scriptStatusOK               = "OK"
	scriptErrNicknameTaken       = "ERR_NICKNAME_TAKEN"
	scriptErrParticipantConflict = "ERR_PARTICIPANT_CONFLICT"
	scriptErrParticipantNotFound = "ERR_PARTICIPANT_NOT_FOUND"
)

var joinParticipantScript = goredis.NewScript(`
local participants_key = KEYS[1]
local token_idx_key = KEYS[2]
local nickname_idx_key = KEYS[3]
local leaderboard_key = KEYS[4]

local participant_id = ARGV[1]
local participant_token = ARGV[2]
local canonical_nickname = ARGV[3]
local payload = ARGV[4]

if redis.call("HEXISTS", nickname_idx_key, canonical_nickname) == 1 then
	return "ERR_NICKNAME_TAKEN"
end

if redis.call("HEXISTS", token_idx_key, participant_token) == 1 then
	return "ERR_PARTICIPANT_CONFLICT"
end

if redis.call("HEXISTS", participants_key, participant_id) == 1 then
	return "ERR_PARTICIPANT_CONFLICT"
end

redis.call("HSET", participants_key, participant_id, payload)
redis.call("HSET", token_idx_key, participant_token, participant_id)
redis.call("HSET", nickname_idx_key, canonical_nickname, participant_id)
redis.call("ZADD", leaderboard_key, 0, participant_id)

return "OK"
`)

var updateParticipantScript = goredis.NewScript(`
local participants_key = KEYS[1]
local leaderboard_key = KEYS[2]

local participant_id = ARGV[1]
local payload = ARGV[2]
local score = ARGV[3]
local update_leaderboard = ARGV[4]

if redis.call("HEXISTS", participants_key, participant_id) == 0 then
	return "ERR_PARTICIPANT_NOT_FOUND"
end

redis.call("HSET", participants_key, participant_id, payload)
if update_leaderboard == "1" then
	redis.call("ZADD", leaderboard_key, score, participant_id)
end

return "OK"
`)
