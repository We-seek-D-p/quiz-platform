package session

import (
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

func (s *Service) calculateDeadline(start time.Time, seconds int) *time.Time {
	t := start.Add(time.Duration(seconds) * time.Second)
	return &t

}

func (s *Service) validateAnswerPayload(q domain.QuestionSnapshot, selected []string) error {
	validIDs := make(map[string]struct{})
	for _, opt := range q.Options {
		validIDs[opt.ID] = struct{}{}
	}

	seen := make(map[string]struct{}, len(selected))

	for _, id := range selected {
		if _, duplicate := seen[id]; duplicate {
			return ErrSelectionCountInvalid
		}
		seen[id] = struct{}{}

		if _, ok := validIDs[id]; !ok {
			return ErrOptionNotInQuestion
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

func (s *Service) collectCorrectOptionIDs(q domain.QuestionSnapshot) []string {
	correctOptionIDs := make([]string, 0, len(q.Options))
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correctOptionIDs = append(correctOptionIDs, opt.ID)
		}
	}

	return correctOptionIDs
}

func (s *Service) checkIsCorrect(q domain.QuestionSnapshot, selected []string) bool {
	correctIDs := s.collectCorrectOptionIDs(q)

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
