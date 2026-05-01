package domain

import "time"

type RuntimeStatus string

const (
	RuntimeStatusLobby        RuntimeStatus = "lobby"
	RuntimeStatusQuestionOpen RuntimeStatus = "question_open"
	RuntimeStatusAnswerReveal RuntimeStatus = "answer_reveal"
	RuntimeStatusFinished     RuntimeStatus = "finished"
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
