package ws

import (
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/service/session"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/transport/ws/dto"
)

// mapSnapshot converts a service snapshot into a public WebSocket DTO.
func mapSnapshot(snapshot session.Snapshot) dto.Snapshot {
	participants := make([]dto.SnapshotParticipant, len(snapshot.Participants))
	for i, participant := range snapshot.Participants {
		participants[i] = dto.SnapshotParticipant{
			ParticipantID: participant.ParticipantID,
			Nickname:      participant.Nickname,
			Score:         participant.Score,
			Rank:          participant.Rank,
			Connected:     participant.Connected,
		}
	}

	mapped := dto.Snapshot{
		SessionID:            snapshot.Runtime.SessionID,
		RoomCode:             snapshot.Runtime.RoomCode,
		Status:               string(snapshot.Runtime.Status),
		CurrentQuestionIndex: snapshot.Runtime.Progress.CurrentQuestionIndex,
		TotalQuestions:       snapshot.Runtime.Progress.TotalQuestions,
		DeadlineAt:           snapshot.Runtime.Progress.DeadlineAt,
		RevealUntil:          snapshot.Runtime.Progress.RevealUntil,
		Participants:         participants,
		LeaderboardTop:       mapLeaderboard(snapshot.LeaderboardTop),
	}

	if snapshot.CurrentQuestion != nil {
		mapped.CurrentQuestion = new(mapQuestion(*snapshot.CurrentQuestion))
	}

	if snapshot.CurrentQuestionReveal != nil {
		mapped.CurrentQuestionReveal = &dto.SnapshotQuestionReveal{
			QuestionID:       snapshot.CurrentQuestionReveal.QuestionID,
			CorrectOptionIDs: snapshot.CurrentQuestionReveal.CorrectOptionIDs,
			RevealDuration:   snapshot.CurrentQuestionReveal.RevealDuration,
			RevealUntil:      snapshot.CurrentQuestionReveal.RevealUntil,
		}
	}

	return mapped
}

// mapQuestion converts a service question snapshot into a public WebSocket DTO.
func mapQuestion(question session.SnapshotQuestion) dto.SnapshotQuestion {
	options := make([]dto.SnapshotQuestionOption, len(question.Options))
	for i, option := range question.Options {
		options[i] = dto.SnapshotQuestionOption{ID: option.ID, Text: option.Text}
	}

	return dto.SnapshotQuestion{
		ID:            question.ID,
		Text:          question.Text,
		SelectionType: question.SelectionType,
		Options:       options,
	}
}

// mapLeaderboard converts service leaderboard entries into public WebSocket DTOs.
func mapLeaderboard(entries []session.SnapshotLeaderboardEntry) []dto.SnapshotLeaderboardEntry {
	mapped := make([]dto.SnapshotLeaderboardEntry, len(entries))
	for i, entry := range entries {
		mapped[i] = dto.SnapshotLeaderboardEntry{
			ParticipantID: entry.ParticipantID,
			Nickname:      entry.Nickname,
			Score:         entry.Score,
			Rank:          entry.Rank,
		}
	}
	return mapped
}

// mapJoinedLobby converts a join result into a public WebSocket DTO.
func mapJoinedLobby(payload session.JoinedLobby) dto.JoinedLobby {
	return dto.JoinedLobby{
		ParticipantID:    payload.ParticipantID,
		ParticipantToken: payload.ParticipantToken,
		Nickname:         payload.Nickname,
		RoomCode:         payload.RoomCode,
		Status:           payload.Status,
	}
}

// mapLobbyUpdated converts a lobby update into a public WebSocket DTO.
func mapLobbyUpdated(payload session.LobbyUpdated) dto.LobbyUpdated {
	return dto.LobbyUpdated{PlayersCount: payload.PlayersCount}
}

// mapQuestionOpened converts a question opening event into a public WebSocket DTO.
func mapQuestionOpened(payload session.QuestionOpened) dto.QuestionOpened {
	return dto.QuestionOpened{
		QuestionIndex:  payload.QuestionIndex,
		TotalQuestions: payload.TotalQuestions,
		Question:       mapQuestion(payload.Question),
		DeadlineAt:     payload.DeadlineAt,
	}
}

// mapAnswerAccepted converts an accepted answer result into a public WebSocket DTO.
func mapAnswerAccepted(payload session.AnswerAccepted) dto.AnswerAccepted {
	return dto.AnswerAccepted{QuestionID: payload.QuestionID, AcceptedAt: payload.AcceptedAt}
}

// mapQuestionProgress converts host progress into a public WebSocket DTO.
func mapQuestionProgress(payload session.QuestionProgress) dto.QuestionProgress {
	return dto.QuestionProgress{
		QuestionID:    payload.QuestionID,
		AnsweredCount: payload.AnsweredCount,
		TotalPlayers:  payload.TotalPlayers,
	}
}

// mapAnswerReveal converts a player reveal payload into a public WebSocket DTO.
func mapAnswerReveal(payload session.AnswerReveal) dto.AnswerReveal {
	return dto.AnswerReveal{
		QuestionID:            payload.QuestionID,
		CorrectOptionIDs:      payload.CorrectOptionIDs,
		YourSelectedOptionIDs: payload.YourSelectedOptionIDs,
		YourResult:            payload.YourResult,
		ScoreDelta:            payload.ScoreDelta,
		TotalScore:            payload.TotalScore,
		YourRank:              payload.YourRank,
		LeaderboardTop:        mapLeaderboard(payload.LeaderboardTop),
		RevealDurationSec:     payload.RevealDurationSec,
		RevealUntil:           payload.RevealUntil,
	}
}

// mapQuestionRevealHost converts a host reveal payload into a public WebSocket DTO.
func mapQuestionRevealHost(payload session.QuestionRevealHost) dto.QuestionRevealHost {
	return dto.QuestionRevealHost{
		QuestionID:        payload.QuestionID,
		CorrectOptionIDs:  payload.CorrectOptionIDs,
		AnsweredCount:     payload.AnsweredCount,
		TotalPlayers:      payload.TotalPlayers,
		LeaderboardTop:    mapLeaderboard(payload.LeaderboardTop),
		RevealDurationSec: payload.RevealDurationSec,
		RevealUntil:       payload.RevealUntil,
	}
}

// mapLeaderboardReveal converts a player leaderboard reveal into a public WebSocket DTO.
func mapLeaderboardReveal(payload session.LeaderboardReveal) dto.LeaderboardReveal {
	return dto.LeaderboardReveal{
		QuestionID:        payload.QuestionID,
		LeaderboardTop:    mapLeaderboard(payload.LeaderboardTop),
		YourScore:         payload.YourScore,
		YourRank:          payload.YourRank,
		RevealDurationSec: payload.RevealDurationSec,
		RevealUntil:       payload.RevealUntil,
	}
}

// mapLeaderboardRevealHost converts a host leaderboard reveal into a public WebSocket DTO.
func mapLeaderboardRevealHost(payload session.LeaderboardRevealHost) dto.LeaderboardRevealHost {
	return dto.LeaderboardRevealHost{
		QuestionID:        payload.QuestionID,
		LeaderboardTop:    mapLeaderboard(payload.LeaderboardTop),
		RevealDurationSec: payload.RevealDurationSec,
		RevealUntil:       payload.RevealUntil,
	}
}

// mapFinished converts final session results into a public WebSocket DTO.
func mapFinished(payload session.Finished) dto.Finished {
	return dto.Finished{LeaderboardTop: mapLeaderboard(payload.LeaderboardTop)}
}
