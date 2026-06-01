package session

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

const leaderboardTopLimit = 10

// mapToManagementResults converts runtime participants into final Management results.
func (s *Service) mapToManagementResults(ps []domain.RuntimeParticipant) []domain.SessionResultParticipant {
	res := make([]domain.SessionResultParticipant, len(ps))
	for i, p := range ps {
		res[i] = domain.SessionResultParticipant{
			ParticipantID: p.ParticipantID,
			Nickname:      p.Nickname,
			Score:         p.Score,
			Rank:          p.Rank,
		}
	}
	return res
}

// loadLeaderboardTop reads and enriches leaderboard entries with participant nicknames.
func (s *Service) loadLeaderboardTop(
	ctx context.Context,
	sessionID string,
	participants []domain.RuntimeParticipant,
	limit int,
) ([]SnapshotLeaderboardEntry, error) {
	entries, err := s.leaderboardRepository.GetTop(ctx, sessionID, limit)
	if err != nil {
		return nil, err
	}

	return mapLeaderboardEntries(entries, participants), nil
}

// mapLeaderboardEntries attaches participant nicknames to raw leaderboard rows.
func mapLeaderboardEntries(entries []domain.LeaderboardEntry, participants []domain.RuntimeParticipant) []SnapshotLeaderboardEntry {
	nicknameByParticipant := make(map[string]string, len(participants))
	for _, participant := range participants {
		nicknameByParticipant[participant.ParticipantID] = participant.Nickname
	}

	mapped := make([]SnapshotLeaderboardEntry, 0, len(entries))
	for _, entry := range entries {
		mapped = append(mapped, SnapshotLeaderboardEntry{
			ParticipantID: entry.ParticipantID,
			Nickname:      nicknameByParticipant[entry.ParticipantID],
			Score:         entry.Score,
			Rank:          entry.Rank,
		})
	}

	return mapped
}

// buildSessionSnapshot assembles a service-level session snapshot without transport DTOs.
func (s *Service) buildSessionSnapshot(
	runtime domain.SessionRuntime,
	quiz domain.QuizSnapshot,
	participants []domain.RuntimeParticipant,
	leaderboardTop []SnapshotLeaderboardEntry,
) Snapshot {
	snapshot := Snapshot{
		Runtime:        runtime,
		Quiz:           quiz,
		Participants:   participants,
		LeaderboardTop: leaderboardTop,
	}

	currIdx := runtime.Progress.CurrentQuestionIndex
	if (runtime.Status == domain.RuntimeStatusQuestionOpen || runtime.Status == domain.RuntimeStatusAnswerReveal || runtime.Status == domain.RuntimeStatusLeaderboardReveal) &&
		currIdx >= 0 &&
		currIdx < len(quiz.Questions) {
		q := quiz.Questions[currIdx]

		snapshot.CurrentQuestion = &SnapshotQuestion{
			ID:            q.ID,
			Text:          q.Text,
			SelectionType: string(q.SelectionType),
			Options:       make([]SnapshotQuestionOption, len(q.Options)),
		}

		for i, opt := range q.Options {
			snapshot.CurrentQuestion.Options[i] = SnapshotQuestionOption{
				ID:   opt.ID,
				Text: opt.Text,
			}
		}

		if runtime.Status == domain.RuntimeStatusAnswerReveal && runtime.Progress.RevealUntil != nil {
			snapshot.CurrentQuestionReveal = &SnapshotQuestionReveal{
				QuestionID:       q.ID,
				CorrectOptionIDs: s.collectCorrectOptionIDs(q),
				RevealDuration:   int(s.answerRevealDuration.Seconds()),
				RevealUntil:      *runtime.Progress.RevealUntil,
			}
		}
	}

	return snapshot
}
