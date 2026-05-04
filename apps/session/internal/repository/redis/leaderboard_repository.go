package redis

import (
	"context"
	"errors"
	"fmt"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type LeaderboardRepository struct {
	client *goredis.Client
}

func NewLeaderboardRepository(client *goredis.Client) *LeaderboardRepository {
	return &LeaderboardRepository{client: client}
}

func (r *LeaderboardRepository) AddScore(ctx context.Context, sessionID, participantID string, delta int) (int, error) {
	key := sessionLeaderboardKey(sessionID)

	score, err := r.client.ZIncrBy(ctx, key, float64(delta), participantID).Result()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return int(score), nil
}

func (r *LeaderboardRepository) SetScore(ctx context.Context, sessionID, participantID string, score int) error {
	key := sessionLeaderboardKey(sessionID)

	if err := r.client.ZAdd(ctx, key, goredis.Z{Score: float64(score), Member: participantID}).Err(); err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return nil
}

func (r *LeaderboardRepository) GetScore(ctx context.Context, sessionID, participantID string) (int, error) {
	key := sessionLeaderboardKey(sessionID)

	score, err := r.client.ZScore(ctx, key, participantID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return 0, ErrLeaderboardEntryNotFound
		}

		return 0, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return int(score), nil
}

func (r *LeaderboardRepository) GetRank(ctx context.Context, sessionID, participantID string) (int, error) {
	key := sessionLeaderboardKey(sessionID)

	rank, err := r.client.ZRevRank(ctx, key, participantID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return 0, ErrLeaderboardEntryNotFound
		}

		return 0, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	return int(rank) + 1, nil
}

func (r *LeaderboardRepository) GetTop(ctx context.Context, sessionID string, limit int) ([]domain.LeaderboardEntry, error) {
	if limit <= 0 {
		return []domain.LeaderboardEntry{}, nil
	}

	key := sessionLeaderboardKey(sessionID)
	items, err := r.client.ZRevRangeWithScores(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	top := make([]domain.LeaderboardEntry, 0, len(items))
	for i, item := range items {
		participantID, ok := item.Member.(string)
		if !ok {
			continue
		}

		top = append(top, domain.LeaderboardEntry{
			ParticipantID: participantID,
			Score:         int(item.Score),
			Rank:          i + 1,
		})
	}

	return top, nil
}
