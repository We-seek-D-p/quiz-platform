package management

import "time"

type ReportSessionResultsRequest struct {
	EventID      string                           `json:"event_id"`
	FinishReason string                           `json:"finish_reason"`
	FinishedAt   time.Time                        `json:"finished_at"`
	Participants []ReportSessionResultParticipant `json:"participants"`
}

type ReportSessionResultParticipant struct {
	ParticipantID string `json:"participant_id"`
	Nickname      string `json:"nickname"`
	Score         int    `json:"score"`
	Rank          int    `json:"rank"`
}
