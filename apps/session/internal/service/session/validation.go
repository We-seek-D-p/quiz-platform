package session

import (
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

func (s *Service) calculateDeadline(start time.Time, seconds int) *time.Time {
	return new(start.Add(time.Duration(seconds) * time.Second))
}

func (s *Service) validateAnswerPayload(q domain.QuestionSnapshot, selected []string) error {
	validIDs := make(map[string]struct{})
	for _, opt := range q.Options {
		validIDs[opt.ID] = struct{}{}
	}

	for _, id := range selected {
		if _, ok := validIDs[id]; !ok {
			return ErrInvalidAnswerPayload
		}
	}

	st := string(q.SelectionType)
	if st == "single" && len(selected) != 1 {
		return ErrSelectionCountInvalid
	}
	if st == "multiple" && len(selected) < 1 {
		return ErrSelectionCountInvalid
	}
	return nil
}

func (s *Service) checkIsCorrect(q domain.QuestionSnapshot, selected []string) bool {
	var correctIDs []string
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correctIDs = append(correctIDs, opt.ID)
		}
	}

	if len(correctIDs) != len(selected) {
		return false
	}

	selectedMap := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		selectedMap[id] = struct{}{}
	}

	for _, id := range correctIDs {
		if _, ok := selectedMap[id]; !ok {
			return false
		}
	}
	return true
}
