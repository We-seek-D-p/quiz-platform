package domain

type EventType string

const (
	EventQuestionOpened      EventType = "question_opened"
	EventAnswerRevealed      EventType = "answer_reveal"
	EventLeaderboardRevealed EventType = "leaderboard_reveal"
	EventGameFinished        EventType = "session_finished"
)

type SessionDomainEvent struct {
	SessionID string
	Type      EventType
	Payload   any
}
