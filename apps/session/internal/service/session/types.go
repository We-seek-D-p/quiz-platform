package session

import (
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

type SnapshotQuestionOption struct {
	ID   string
	Text string
}

type SnapshotParticipant struct {
	ParticipantID string
	Nickname      string
	Score         int
	Rank          int
	Connected     bool
}

type SnapshotLeaderboardEntry struct {
	ParticipantID string
	Nickname      string
	Score         int
	Rank          int
}

type SnapshotQuestion struct {
	ID            string
	Text          string
	SelectionType string
	Options       []SnapshotQuestionOption
}

type SnapshotQuestionReveal struct {
	QuestionID       string
	CorrectOptionIDs []string
	RevealDuration   int
	RevealUntil      time.Time
}

type Snapshot struct {
	Runtime               domain.SessionRuntime
	Quiz                  domain.QuizSnapshot
	Participants          []domain.RuntimeParticipant
	LeaderboardTop        []SnapshotLeaderboardEntry
	CurrentQuestion       *SnapshotQuestion
	CurrentQuestionReveal *SnapshotQuestionReveal
}

type Finished struct {
	LeaderboardTop []SnapshotLeaderboardEntry
}

type JoinedLobby struct {
	ParticipantID    string
	ParticipantToken string
	Nickname         string
	RoomCode         string
	Status           string
}

type LobbyUpdated struct {
	PlayersCount int
}

type QuestionOpened struct {
	QuestionIndex  int
	TotalQuestions int
	Question       SnapshotQuestion
	DeadlineAt     time.Time
}

type AnswerAccepted struct {
	QuestionID string
	AcceptedAt time.Time
}

type QuestionProgress struct {
	QuestionID    string
	AnsweredCount int
	TotalPlayers  int
}

type AnswerReveal struct {
	QuestionID            string
	CorrectOptionIDs      []string
	YourSelectedOptionIDs []string
	YourResult            string
	ScoreDelta            int
	TotalScore            int
	YourRank              int
	LeaderboardTop        []SnapshotLeaderboardEntry
	RevealDurationSec     int
	RevealUntil           time.Time
}

type ParticipantAnswerReveal struct {
	ParticipantID string
	Payload       AnswerReveal
}

type QuestionRevealHost struct {
	QuestionID        string
	CorrectOptionIDs  []string
	AnsweredCount     int
	TotalPlayers      int
	LeaderboardTop    []SnapshotLeaderboardEntry
	RevealDurationSec int
	RevealUntil       time.Time
}

type RevealTransitionResult struct {
	SessionSnapshot Snapshot
	HostReveal      QuestionRevealHost
	PlayerReveals   []ParticipantAnswerReveal
}

type LeaderboardReveal struct {
	QuestionID        string
	LeaderboardTop    []SnapshotLeaderboardEntry
	YourScore         int
	YourRank          int
	RevealDurationSec int
	RevealUntil       time.Time
}

type ParticipantLeaderboardReveal struct {
	ParticipantID string
	Payload       LeaderboardReveal
}

type LeaderboardRevealHost struct {
	QuestionID        string
	LeaderboardTop    []SnapshotLeaderboardEntry
	RevealDurationSec int
	RevealUntil       time.Time
}

type LeaderboardTransitionResult struct {
	SessionSnapshot Snapshot
	HostReveal      LeaderboardRevealHost
	PlayerReveals   []ParticipantLeaderboardReveal
}

type HostConnectResult struct {
	SessionSnapshot Snapshot
}

type PlayerJoinResult struct {
	JoinedLobby     JoinedLobby
	LobbyUpdated    LobbyUpdated
	SessionSnapshot Snapshot
}

type PlayerReconnectResult struct {
	ParticipantID   string
	SessionSnapshot Snapshot
}

type StartGameResult struct {
	QuestionOpened   QuestionOpened
	SessionSnapshot  Snapshot
	PersistedStatus  string
	PersistedEventID string
}

type SubmitAnswerResult struct {
	AnswerAccepted AnswerAccepted
	HostProgress   *QuestionProgress
}

type FinishGameResult struct {
	SessionFinished Finished
	PersistedStatus string
	PersistedAt     time.Time
}
