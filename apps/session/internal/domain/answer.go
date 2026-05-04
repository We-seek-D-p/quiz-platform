package domain

import "time"

type RuntimeAnswer struct {
	ParticipantID     string
	SelectedOptionIDs []string
	SubmittedAt       time.Time
	Result            string
	ScoreDelta        int
}
