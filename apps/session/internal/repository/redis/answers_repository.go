package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type AnswersRepository struct {
	client *goredis.Client
}

func NewAnswersRepository(client *goredis.Client) *AnswersRepository {
	return &AnswersRepository{client: client}
}

func (r *AnswersRepository) SubmitOnce(ctx context.Context, sessionID, questionID string, answer domain.RuntimeAnswer) error {
	key := sessionAnswersKey(sessionID, questionID)

	answer = normalizeRuntimeAnswer(answer)

	payload, err := json.Marshal(answer)
	if err != nil {
		return fmt.Errorf("marshal answer: %w", err)
	}

	created, err := r.client.HSetNX(ctx, key, answer.ParticipantID, payload).Result()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}
	if !created {
		return ErrAnswerAlreadySubmitted
	}

	return nil
}

func (r *AnswersRepository) GetByParticipant(ctx context.Context, sessionID, questionID, participantID string) (domain.RuntimeAnswer, error) {
	key := sessionAnswersKey(sessionID, questionID)

	payload, err := r.client.HGet(ctx, key, participantID).Result()
	if err != nil {
		if errors.Is(err, goredis.Nil) {
			return domain.RuntimeAnswer{}, ErrAnswerNotFound
		}

		return domain.RuntimeAnswer{}, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	answer, err := unmarshalRuntimeAnswer(payload)
	if err != nil {
		return domain.RuntimeAnswer{}, err
	}

	return answer, nil
}

func (r *AnswersRepository) ListByQuestion(ctx context.Context, sessionID, questionID string) ([]domain.RuntimeAnswer, error) {
	key := sessionAnswersKey(sessionID, questionID)

	payloads, err := r.client.HVals(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRedisUnavailable, err)
	}

	answers := make([]domain.RuntimeAnswer, 0, len(payloads))
	for _, payload := range payloads {
		answer, err := unmarshalRuntimeAnswer(payload)
		if err != nil {
			return nil, err
		}

		answers = append(answers, answer)
	}

	return answers, nil
}

func unmarshalRuntimeAnswer(payload string) (domain.RuntimeAnswer, error) {
	var answer domain.RuntimeAnswer
	if err := json.Unmarshal([]byte(payload), &answer); err != nil {
		return domain.RuntimeAnswer{}, fmt.Errorf("unmarshal answer: %w", err)
	}

	return answer, nil
}

func normalizeRuntimeAnswer(answer domain.RuntimeAnswer) domain.RuntimeAnswer {
	answer.ParticipantID = strings.TrimSpace(answer.ParticipantID)

	unique := make(map[string]struct{}, len(answer.SelectedOptionIDs))
	normalized := make([]string, 0, len(answer.SelectedOptionIDs))

	for _, optionID := range answer.SelectedOptionIDs {
		optionID = strings.TrimSpace(optionID)
		if optionID == "" {
			continue
		}

		if _, exists := unique[optionID]; exists {
			continue
		}

		unique[optionID] = struct{}{}
		normalized = append(normalized, optionID)
	}

	sort.Strings(normalized)
	answer.SelectedOptionIDs = normalized

	return answer
}
