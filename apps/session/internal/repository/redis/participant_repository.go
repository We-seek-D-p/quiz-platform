package redis

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type ParticipantRepository struct {
	client *goredis.Client
}

func NewParticipantRepository(client *goredis.Client) *ParticipantRepository {
	return &ParticipantRepository{client: client}
}

func (r *ParticipantRepository) Create(ctx context.Context, sessionID string, participant domain.RuntimeParticipant) error {
	participantsKey := sessionParticipantsKey(sessionID)
	tokenIndexKey := sessionParticipantTokenIndexKey(sessionID)
	nicknameIndexKey := sessionParticipantNicknameIndexKey(sessionID)
	leaderboardKey := sessionLeaderboardKey(sessionID)

	if err := validateNickname(participant.Nickname); err != nil {
		return err
	}

	nicknameKey := canonicalNicknameKey(participant.Nickname)
	payload, err := json.Marshal(participant)
	if err != nil {
		return errSerializationFailure("failed to marshal participant", err)
	}

	keys := []string{participantsKey, tokenIndexKey, nicknameIndexKey, leaderboardKey}
	args := []any{participant.ParticipantID, participant.ParticipantToken, nicknameKey, string(payload)}
	res, err := joinParticipantScript.Run(ctx, r.client, keys, args...).Result()
	if err != nil {
		return errRuntimeStoreUnavailable(err)
	}

	return mapParticipantScriptStatus(res)
}

func (r *ParticipantRepository) GetByToken(ctx context.Context, sessionID, participantToken string) (domain.RuntimeParticipant, error) {
	tokenIndexKey := sessionParticipantTokenIndexKey(sessionID)

	participantID, err := r.client.HGet(ctx, tokenIndexKey, participantToken).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeParticipant{}, errParticipantNotFound(err)
		}

		return domain.RuntimeParticipant{}, errRuntimeStoreUnavailable(err)
	}

	return r.GetByID(ctx, sessionID, participantID)
}

func (r *ParticipantRepository) GetByNickname(ctx context.Context, sessionID, nickname string) (domain.RuntimeParticipant, error) {
	nicknameIndexKey := sessionParticipantNicknameIndexKey(sessionID)

	if err := validateNickname(nickname); err != nil {
		return domain.RuntimeParticipant{}, err
	}

	nicknameKey := canonicalNicknameKey(nickname)

	participantID, err := r.client.HGet(ctx, nicknameIndexKey, nicknameKey).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeParticipant{}, errParticipantNotFound(err)
		}

		return domain.RuntimeParticipant{}, errRuntimeStoreUnavailable(err)
	}

	return r.GetByID(ctx, sessionID, participantID)
}

func (r *ParticipantRepository) GetByID(ctx context.Context, sessionID, participantID string) (domain.RuntimeParticipant, error) {
	participantsKey := sessionParticipantsKey(sessionID)

	payload, err := r.client.HGet(ctx, participantsKey, participantID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeParticipant{}, errParticipantNotFound(err)
		}

		return domain.RuntimeParticipant{}, errRuntimeStoreUnavailable(err)
	}

	participant, err := unmarshalParticipant(payload)
	if err != nil {
		return domain.RuntimeParticipant{}, err
	}

	return participant, nil
}

func (r *ParticipantRepository) List(ctx context.Context, sessionID string) ([]domain.RuntimeParticipant, error) {
	participantsKey := sessionParticipantsKey(sessionID)

	payloads, err := r.client.HVals(ctx, participantsKey).Result()
	if err != nil {
		return nil, errRuntimeStoreUnavailable(err)
	}

	participants := make([]domain.RuntimeParticipant, 0, len(payloads))
	for _, payload := range payloads {
		participant, err := unmarshalParticipant(payload)
		if err != nil {
			return nil, err
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (r *ParticipantRepository) SetConnected(ctx context.Context, sessionID, participantID string, connected bool) error {
	participant, err := r.GetByID(ctx, sessionID, participantID)
	if err != nil {
		return err
	}

	participant.Connected = connected
	now := time.Now().UTC()
	participant.LastSeenAt = &now

	return r.update(ctx, sessionID, participant, false)
}

func (r *ParticipantRepository) UpdateScoreAndRank(ctx context.Context, sessionID, participantID string, score, rank int) error {
	participant, err := r.GetByID(ctx, sessionID, participantID)
	if err != nil {
		return err
	}

	participant.Score = score
	participant.Rank = rank

	return r.update(ctx, sessionID, participant, true)
}

func (r *ParticipantRepository) update(ctx context.Context, sessionID string, participant domain.RuntimeParticipant, updateLeaderboard bool) error {
	participantsKey := sessionParticipantsKey(sessionID)
	leaderboardKey := sessionLeaderboardKey(sessionID)

	payload, err := json.Marshal(participant)
	if err != nil {
		return errSerializationFailure("failed to marshal participant", err)
	}

	updateLeaderboardFlag := "0"
	if updateLeaderboard {
		updateLeaderboardFlag = "1"
	}

	keys := []string{participantsKey, leaderboardKey}
	args := []any{participant.ParticipantID, string(payload), strconv.Itoa(participant.Score), updateLeaderboardFlag}
	res, err := updateParticipantScript.Run(ctx, r.client, keys, args...).Result()
	if err != nil {
		return errRuntimeStoreUnavailable(err)
	}

	return mapParticipantScriptStatus(res)
}

func unmarshalParticipant(payload string) (domain.RuntimeParticipant, error) {
	var participant domain.RuntimeParticipant
	if err := json.Unmarshal([]byte(payload), &participant); err != nil {
		return domain.RuntimeParticipant{}, errSerializationFailure("failed to unmarshal participant", err)
	}

	return participant, nil
}

func validateNickname(nickname string) error {
	n := strings.TrimSpace(nickname)
	if len(n) > 64 || len(n) < 2 {
		return errInvalidNickname(nil)
	}

	return nil
}

func mapParticipantScriptStatus(result any) error {
	status, ok := result.(string)
	if !ok {
		return errSerializationFailure("unexpected redis script response", nil)
	}

	switch status {
	case scriptStatusOK:
		return nil
	case scriptErrNicknameTaken:
		return errNicknameTaken(nil)
	case scriptErrParticipantConflict:
		return errParticipantConflict(nil)
	case scriptErrParticipantNotFound:
		return errParticipantNotFound(nil)
	default:
		return errSerializationFailure("unexpected redis script status", nil)
	}
}

func canonicalNicknameKey(nickname string) string {
	return strings.ToLower(strings.TrimSpace(nickname))
}
