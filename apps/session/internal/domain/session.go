package domain

import "time"

type RuntimeStatus string

type PersistedStatus string

const (
	RuntimeStatusLobby        RuntimeStatus = "lobby"
	RuntimeStatusQuestionOpen RuntimeStatus = "question_open"
	RuntimeStatusAnswerReveal RuntimeStatus = "answer_reveal"
	RuntimeStatusFinished     RuntimeStatus = "finished"
)

const (
	PersistedStatusInitializing PersistedStatus = "initializing"
	PersistedStatusLobby        PersistedStatus = "lobby"
	PersistedStatusInProgress   PersistedStatus = "in_progress"
	PersistedStatusFinished     PersistedStatus = "finished"
	PersistedStatusInitFailed   PersistedStatus = "init_failed"
)

type SessionRuntime struct {
	SessionID     string
	QuizID        string
	HostID        string
	RoomCode      string
	Status        RuntimeStatus
	InitializedAt time.Time
}

type SessionBootstrap struct {
	SessionID string
	QuizID    string
	HostID    string
	Status    string
	Quiz      QuizSnapshot
}

type SessionStatusUpdate struct {
	Status    PersistedStatus
	StartedAt *time.Time
	EventID   string
}

type SessionResults struct {
	EventID      string
	FinishReason string
	FinishedAt   time.Time
	Participants []SessionResultParticipant
}

type SessionResultParticipant struct {
	ParticipantID string
	Nickname      string
	Score         int
	Rank          int
}
