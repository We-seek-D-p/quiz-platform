package domain

import "time"

const (
	MaxScorePerQuestion    = 1000
	correctnessScoreWeight = 0.65
	speedScoreWeight       = 0.35
)

// CalculatePoints returns a 0..1000 score based on answer correctness and response speed.
func CalculatePoints(question QuestionSnapshot, selectedOptionIDs []string, openedAt, submittedAt time.Time) int {
	correctness := CalculateCorrectness(question, selectedOptionIDs)
	if correctness <= 0 {
		return 0
	}

	speed := calculateSpeedRatio(question.TimeLimitSeconds, openedAt, submittedAt)
	score := float64(MaxScorePerQuestion) * correctness * (correctnessScoreWeight + speedScoreWeight*speed)

	return int(score)
}

// CalculateCorrectness returns a 0..1 ratio and penalizes selected wrong options.
func CalculateCorrectness(question QuestionSnapshot, selectedOptionIDs []string) float64 {
	correctIDs := make(map[string]struct{})
	wrongIDs := make(map[string]struct{})

	for _, option := range question.Options {
		if option.IsCorrect {
			correctIDs[option.ID] = struct{}{}
			continue
		}
		wrongIDs[option.ID] = struct{}{}
	}

	if len(correctIDs) == 0 {
		return 0
	}

	selected := make(map[string]struct{}, len(selectedOptionIDs))
	for _, optionID := range selectedOptionIDs {
		selected[optionID] = struct{}{}
	}

	correctSelected := 0
	wrongSelected := 0
	for optionID := range selected {
		if _, ok := correctIDs[optionID]; ok {
			correctSelected++
			continue
		}
		if _, ok := wrongIDs[optionID]; ok {
			wrongSelected++
		}
	}

	correctRatio := float64(correctSelected) / float64(len(correctIDs))
	wrongRatio := 0.0
	if len(wrongIDs) > 0 {
		wrongRatio = float64(wrongSelected) / float64(len(wrongIDs))
	}

	return clamp01(correctRatio - wrongRatio)
}

// calculateSpeedRatio returns a 0..1 ratio where faster submissions score higher.
func calculateSpeedRatio(timeLimitSec int, openedAt, submittedAt time.Time) float64 {
	if timeLimitSec <= 0 {
		return 0
	}

	elapsed := submittedAt.Sub(openedAt).Seconds()
	if elapsed <= 0 {
		return 1
	}

	limit := float64(timeLimitSec)
	if elapsed >= limit {
		return 0
	}

	return clamp01(1 - elapsed/limit)
}

// clamp01 bounds a floating point ratio to the inclusive 0..1 range.
func clamp01(value float64) float64 {
	if value < 0 {
		return 0
	}
	if value > 1 {
		return 1
	}
	return value
}
