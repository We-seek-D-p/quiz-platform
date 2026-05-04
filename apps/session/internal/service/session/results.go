package session

import "time"

type SnapshotQuestionOptionDTO struct {
	ID   string
	Text string
}

type SnapshotParticipantDTO struct {
	ParticipantID string
	Nickname      string
	Score         int
	Rank          int
	Connected     bool
}

type SnapshotLeaderboardEntryDTO struct {
	ParticipantID string
	Nickname      string
	Score         int
	Rank          int
}

type SnapshotQuestionDTO struct {
	ID            string
	Text          string
	SelectionType string
	Options       []SnapshotQuestionOptionDTO
}

type SnapshotQuestionRevealDTO struct {
	QuestionID       string
	CorrectOptionIDs []string
	RevealDuration   int
	RevealUntil      time.Time
}

type SnapshotDTO struct {
	SessionID             string
	RoomCode              string
	Status                string
	CurrentQuestionIndex  int
	TotalQuestions        int
	DeadlineAt            *time.Time
	RevealUntil           *time.Time
	Participants          []SnapshotParticipantDTO
	CurrentQuestion       *SnapshotQuestionDTO
	LeaderboardTop        []SnapshotLeaderboardEntryDTO
	CurrentQuestionReveal *SnapshotQuestionRevealDTO
}

type FinishedDTO struct {
	LeaderboardTop []SnapshotLeaderboardEntryDTO
}

type JoinedLobbyDTO struct {
	ParticipantID    string
	ParticipantToken string
	Nickname         string
	RoomCode         string
	Status           string
}

type LobbyUpdatedDTO struct {
	PlayersCount int
}

type QuestionOpenedDTO struct {
	QuestionIndex  int
	TotalQuestions int
	Question       SnapshotQuestionDTO
	DeadlineAt     time.Time
}

type AnswerAcceptedDTO struct {
	QuestionID string
	AcceptedAt time.Time
}

type QuestionProgressDTO struct {
	QuestionID    string
	AnsweredCount int
	TotalPlayers  int
}

type AnswerRevealDTO struct {
	QuestionID            string
	CorrectOptionIDs      []string
	YourSelectedOptionIDs []string
	YourResult            string
	ScoreDelta            int
	TotalScore            int
	YourRank              int
	LeaderboardTop        []SnapshotLeaderboardEntryDTO
	RevealDurationSec     int
	RevealUntil           time.Time
}

type QuestionRevealHostDTO struct {
	QuestionID        string
	CorrectOptionIDs  []string
	AnsweredCount     int
	TotalPlayers      int
	LeaderboardTop    []SnapshotLeaderboardEntryDTO
	RevealDurationSec int
	RevealUntil       time.Time
}

type HostConnectResult struct {
	SessionSnapshot SnapshotDTO
}

type PlayerJoinResult struct {
	JoinedLobby     JoinedLobbyDTO
	LobbyUpdated    LobbyUpdatedDTO
	SessionSnapshot SnapshotDTO
}

type PlayerReconnectResult struct {
	SessionSnapshot SnapshotDTO
}

type StartGameResult struct {
	QuestionOpened   QuestionOpenedDTO
	SessionSnapshot  SnapshotDTO
	PersistedStatus  string
	PersistedEventID string
}

type SubmitAnswerResult struct {
	AnswerAccepted AnswerAcceptedDTO
	HostProgress   *QuestionProgressDTO
}

type FinishGameResult struct {
	SessionFinished FinishedDTO
	PersistedStatus string
	PersistedAt     time.Time
}
