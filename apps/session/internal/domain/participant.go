package domain

import (
	"strings"
	"time"
)

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

// CanonicalNickname filters and formats raw string vectors to secure uniqueness checks inside cache maps.
func (p RuntimeParticipant) CanonicalNickname() string {
	return strings.ToLower(strings.TrimSpace(p.Nickname))
}
