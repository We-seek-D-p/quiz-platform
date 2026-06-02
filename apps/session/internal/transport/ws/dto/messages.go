package dto

import "time"

type SnapshotQuestionOption struct {
	ID   string `json:"id"`
	Text string `json:"text"`
}

type SnapshotParticipant struct {
	ParticipantID string `json:"participant_id"`
	Nickname      string `json:"nickname"`
	Score         int    `json:"score"`
	Rank          int    `json:"rank"`
	Connected     bool   `json:"connected"`
}

type SnapshotLeaderboardEntry struct {
	ParticipantID string `json:"participant_id"`
	Nickname      string `json:"nickname"`
	Score         int    `json:"score"`
	Rank          int    `json:"rank"`
}

type SnapshotQuestion struct {
	ID            string                   `json:"id"`
	Text          string                   `json:"text"`
	SelectionType string                   `json:"selection_type"`
	Options       []SnapshotQuestionOption `json:"options"`
}

type SnapshotQuestionReveal struct {
	QuestionID       string    `json:"question_id"`
	CorrectOptionIDs []string  `json:"correct_option_ids"`
	RevealDuration   int       `json:"reveal_duration"`
	RevealUntil      time.Time `json:"reveal_until"`
}

type Snapshot struct {
	SessionID             string                     `json:"session_id"`
	RoomCode              string                     `json:"room_code"`
	Status                string                     `json:"status"`
	CurrentQuestionIndex  int                        `json:"current_question_index"`
	TotalQuestions        int                        `json:"total_questions"`
	DeadlineAt            *time.Time                 `json:"deadline_at"`
	RevealUntil           *time.Time                 `json:"reveal_until"`
	Participants          []SnapshotParticipant      `json:"participants"`
	CurrentQuestion       *SnapshotQuestion          `json:"current_question"`
	LeaderboardTop        []SnapshotLeaderboardEntry `json:"leaderboard_top"`
	CurrentQuestionReveal *SnapshotQuestionReveal    `json:"current_question_reveal"`
}

type JoinedLobby struct {
	ParticipantID    string `json:"participant_id"`
	ParticipantToken string `json:"participant_token"`
	Nickname         string `json:"nickname"`
	RoomCode         string `json:"room_code"`
	Status           string `json:"status"`
}

type LobbyUpdated struct {
	PlayersCount int `json:"players_count"`
}

type QuestionOpened struct {
	QuestionIndex  int              `json:"question_index"`
	TotalQuestions int              `json:"total_questions"`
	Question       SnapshotQuestion `json:"question"`
	DeadlineAt     time.Time        `json:"deadline_at"`
}

type AnswerAccepted struct {
	QuestionID string    `json:"question_id"`
	AcceptedAt time.Time `json:"accepted_at"`
}

type QuestionProgress struct {
	QuestionID    string `json:"question_id"`
	AnsweredCount int    `json:"answered_count"`
	TotalPlayers  int    `json:"total_players"`
}

type AnswerReveal struct {
	QuestionID            string                     `json:"question_id"`
	CorrectOptionIDs      []string                   `json:"correct_option_ids"`
	YourSelectedOptionIDs []string                   `json:"your_selected_option_ids"`
	YourResult            string                     `json:"your_result"`
	ScoreDelta            int                        `json:"score_delta"`
	TotalScore            int                        `json:"total_score"`
	YourRank              int                        `json:"your_rank"`
	LeaderboardTop        []SnapshotLeaderboardEntry `json:"leaderboard_top"`
	RevealDurationSec     int                        `json:"reveal_duration_sec"`
	RevealUntil           time.Time                  `json:"reveal_until"`
}

type QuestionRevealHost struct {
	QuestionID        string                     `json:"question_id"`
	CorrectOptionIDs  []string                   `json:"correct_option_ids"`
	AnsweredCount     int                        `json:"answered_count"`
	TotalPlayers      int                        `json:"total_players"`
	LeaderboardTop    []SnapshotLeaderboardEntry `json:"leaderboard_top"`
	RevealDurationSec int                        `json:"reveal_duration_sec"`
	RevealUntil       time.Time                  `json:"reveal_until"`
}

type LeaderboardReveal struct {
	QuestionID        string                     `json:"question_id"`
	LeaderboardTop    []SnapshotLeaderboardEntry `json:"leaderboard_top"`
	YourScore         int                        `json:"your_score"`
	YourRank          int                        `json:"your_rank"`
	RevealDurationSec int                        `json:"reveal_duration_sec"`
	RevealUntil       time.Time                  `json:"reveal_until"`
}

type LeaderboardRevealHost struct {
	QuestionID        string                     `json:"question_id"`
	LeaderboardTop    []SnapshotLeaderboardEntry `json:"leaderboard_top"`
	RevealDurationSec int                        `json:"reveal_duration_sec"`
	RevealUntil       time.Time                  `json:"reveal_until"`
}

type Finished struct {
	LeaderboardTop []SnapshotLeaderboardEntry `json:"leaderboard_top"`
}
