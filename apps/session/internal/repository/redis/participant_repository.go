package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

const participantCreateMaxRetries = 5

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

	if err := validateNickname(participant.Nickname); err != nil {
		return err
	}

	nicknameKey := canonicalNicknameKey(participant.Nickname)
	payload, err := json.Marshal(participant)
	if err != nil {
		return fmt.Errorf("marshal participant: %w", err)
	}

	for attempt := 0; attempt < participantCreateMaxRetries; attempt++ {
		err = r.client.Watch(ctx, func(tx *goredis.Tx) error {
			nicknameExists, err := tx.HExists(ctx, nicknameIndexKey, nicknameKey).Result()
			if err != nil {
				return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
			}
			if nicknameExists {
				return ErrNicknameTaken
			}

			tokenExists, err := tx.HExists(ctx, tokenIndexKey, participant.ParticipantToken).Result()
			if err != nil {
				return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
			}
			if tokenExists {
				return ErrParticipantConflict
			}

			participantExists, err := tx.HExists(ctx, participantsKey, participant.ParticipantID).Result()
			if err != nil {
				return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
			}
			if participantExists {
				return ErrParticipantConflict
			}

			_, err = tx.TxPipelined(ctx, func(pipe goredis.Pipeliner) error {
				pipe.HSet(ctx, participantsKey, participant.ParticipantID, payload)
				pipe.HSet(ctx, tokenIndexKey, participant.ParticipantToken, participant.ParticipantID)
				pipe.HSet(ctx, nicknameIndexKey, nicknameKey, participant.ParticipantID)
				return nil
			})

			return err
		}, participantsKey, tokenIndexKey, nicknameIndexKey)

		if err == nil {
			return nil
		}

		if errors.Is(err, ErrNicknameTaken) || errors.Is(err, ErrParticipantConflict) || errors.Is(err, ErrRedisUnavailable) {
			return err
		}

		if errors.Is(err, goredis.TxFailedErr) {
			continue
		}

		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return fmt.Errorf("%w: participant create retries exceeded", ErrRedisUnavailable)
}

func (r *ParticipantRepository) GetByToken(ctx context.Context, sessionID, participantToken string) (domain.RuntimeParticipant, error) {
	tokenIndexKey := sessionParticipantTokenIndexKey(sessionID)

	participantID, err := r.client.HGet(ctx, tokenIndexKey, participantToken).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeParticipant{}, ErrParticipantNotFound
		}

		return domain.RuntimeParticipant{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
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
			return domain.RuntimeParticipant{}, ErrParticipantNotFound
		}

		return domain.RuntimeParticipant{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return r.GetByID(ctx, sessionID, participantID)
}

func (r *ParticipantRepository) GetByID(ctx context.Context, sessionID, participantID string) (domain.RuntimeParticipant, error) {
	participantsKey := sessionParticipantsKey(sessionID)

	payload, err := r.client.HGet(ctx, participantsKey, participantID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeParticipant{}, ErrParticipantNotFound
		}

		return domain.RuntimeParticipant{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
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
		return nil, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
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
    participant.LastSeenAt = new(time.Now().UTC())

	return r.update(ctx, sessionID, participant)
}

func (r *ParticipantRepository) UpdateScoreAndRank(ctx context.Context, sessionID, participantID string, score, rank int) error {
	participant, err := r.GetByID(ctx, sessionID, participantID)
	if err != nil {
		return err
	}

	participant.Score = score
	participant.Rank = rank

	return r.update(ctx, sessionID, participant)
}

func (r *ParticipantRepository) update(ctx context.Context, sessionID string, participant domain.RuntimeParticipant) error {
	participantsKey := sessionParticipantsKey(sessionID)
	leaderboardKey := sessionLeaderboardKey(sessionID)

	payload, err := json.Marshal(participant)
	if err != nil {
		return fmt.Errorf("marshal participant: %w", err)
	}

	exists, err := r.client.HExists(ctx, participantsKey, participant.ParticipantID).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if !exists {
		return ErrParticipantNotFound
	}

	pipe := r.client.TxPipeline()
	pipe.HSet(ctx, participantsKey, participant.ParticipantID, payload)

	pipe.ZAdd(ctx, leaderboardKey, goredis.Z{
		Score:  float64(participant.Score),
		Member: participant.ParticipantID,
	})

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}

func unmarshalParticipant(payload string) (domain.RuntimeParticipant, error) {
	var participant domain.RuntimeParticipant
	if err := json.Unmarshal([]byte(payload), &participant); err != nil {
		return domain.RuntimeParticipant{}, fmt.Errorf("unmarshal participant: %w", err)
	}

	return participant, nil
}

func validateNickname(nickname string) error {
	n := strings.TrimSpace(nickname)
	if len(n) > 64 || len(n) < 2 {
		return ErrInvalidNickname
	}

	return nil
}

func canonicalNicknameKey(nickname string) string {
	return strings.ToLower(strings.TrimSpace(nickname))
}
