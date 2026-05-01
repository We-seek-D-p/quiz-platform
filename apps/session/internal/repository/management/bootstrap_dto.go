package management

type BootstrapResponse struct {
	Session      BootstrapSessionDTO `json:"session"`
	QuizSnapshot QuizSnapshotDTO     `json:"quiz_snapshot"`
}

type BootstrapSessionDTO struct {
	SessionID string `json:"session_id"`
	QuizID    string `json:"quiz_id"`
	HostID    string `json:"host_id"`
	Status    string `json:"status"`
}

type QuizSnapshotDTO struct {
	Title     string                `json:"title"`
	Questions []QuestionSnapshotDTO `json:"questions"`
}

type QuestionSnapshotDTO struct {
	ID               string              `json:"id"`
	Text             string              `json:"text"`
	SelectionType    string              `json:"selection_type"`
	TimeLimitSeconds int                 `json:"time_limit_seconds"`
	OrderIndex       int                 `json:"order_index"`
	Options          []OptionSnapshotDTO `json:"options"`
}

type OptionSnapshotDTO struct {
	ID         string `json:"id"`
	Text       string `json:"text"`
	OrderIndex int    `json:"order_index"`
	IsCorrect  bool   `json:"is_correct"`
}
