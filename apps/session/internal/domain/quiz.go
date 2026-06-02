package domain

type SelectionType string

const (
	SelectionTypeSingle   SelectionType = "single"
	SelectionTypeMultiple SelectionType = "multiple"
)

type QuizSnapshot struct {
	Title     string
	Questions []QuestionSnapshot
}

type QuestionSnapshot struct {
	ID               string
	Text             string
	SelectionType    SelectionType
	TimeLimitSeconds int
	OrderIndex       int
	Options          []OptionSnapshot
}

// IsSingle flags if the internal step constraint expects exactly one chosen variable element.
func (q QuestionSnapshot) IsSingle() bool {
	return q.SelectionType == SelectionTypeSingle
}

// IsMultiple flags if multiselection rule schemas are enabled for evaluation parameters.
func (q QuestionSnapshot) IsMultiple() bool {
	return q.SelectionType == SelectionTypeMultiple
}

// ValidateSelectionCount verifies if the absolute payload slice length maps cleanly inside question layout bounds.
func (q QuestionSnapshot) ValidateSelectionCount(count int) bool {
	if q.IsSingle() && count != 1 {
		return false
	}
	if q.IsMultiple() && (count < 1 || count > len(q.Options)) {
		return false
	}
	return true
}

type OptionSnapshot struct {
	ID         string
	Text       string
	OrderIndex int
	IsCorrect  bool
}
