package domain

import "time"

type RuntimeParticipant struct {
	ParticipantID    string
	ParticipantToken string
	Nickname         string
	Score            int
	Rank             int
	Connected        bool
	JoinedAt         time.Time
	LastSeenAt       *time.Time
}
