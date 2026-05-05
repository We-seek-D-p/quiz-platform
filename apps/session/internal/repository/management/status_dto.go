package management

import "time"

type ReportSessionStatusRequest struct {
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EventID   string     `json:"event_id"`
}
