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

type OptionSnapshot struct {
	ID         string
	Text       string
	OrderIndex int
	IsCorrect  bool
}
