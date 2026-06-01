package session

import (
	"context"
	"errors"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/client/management"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
)

const leaderboardTopLimit = 10

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

func (s *Service) buildParticipantNicknameIndex(participants []domain.RuntimeParticipant) map[string]string {
	index := make(map[string]string, len(participants))
	for _, participant := range participants {
		index[participant.ParticipantID] = participant.Nickname
	}

	return index
}

func (s *Service) mapLeaderboardEntriesToSnapshot(
	entries []domain.LeaderboardEntry,
	participants []domain.RuntimeParticipant,
) []SnapshotLeaderboardEntryDTO {
	nicknameByParticipant := s.buildParticipantNicknameIndex(participants)
	mapped := make([]SnapshotLeaderboardEntryDTO, 0, len(entries))

	for _, entry := range entries {
		mapped = append(mapped, SnapshotLeaderboardEntryDTO{
			ParticipantID: entry.ParticipantID,
			Nickname:      nicknameByParticipant[entry.ParticipantID],
			Score:         entry.Score,
			Rank:          entry.Rank,
		})
	}

	return mapped
}

func (s *Service) loadLeaderboardTop(
	ctx context.Context,
	sessionID string,
	participants []domain.RuntimeParticipant,
	limit int,
) ([]SnapshotLeaderboardEntryDTO, error) {
	entries, err := s.leaderboardRepository.GetTop(ctx, sessionID, limit)
	if err != nil {
		return nil, s.mapLeaderboardRepositoryError(err)
	}

	return s.mapLeaderboardEntriesToSnapshot(entries, participants), nil
}

func (s *Service) mapManagementError(err error) error {
	if errors.Is(err, management.ErrSessionNotFound) {
		return ErrSessionNotFound
	}
	if errors.Is(err, management.ErrAlreadyFinished) {
		return ErrSessionAlreadyFinished
	}
	return ErrBootstrapFetchFailed
}

func (s *Service) mapRedisError(err error) error {
	if errors.Is(err, redis.ErrSessionNotFound) {
		return ErrSessionRuntimeNotFound
	}
	if errors.Is(err, redis.ErrSessionConflict) {
		return ErrSessionRuntimeConflict
	}
	return ErrRuntimeStoreUnavailable
}

func (s *Service) mapRoomCodeError(err error) error {
	if errors.Is(err, redis.ErrRoomNotFound) {
		return ErrRoomNotFound
	}

	return ErrRuntimeStoreUnavailable
}

func (s *Service) mapParticipantRepositoryError(err error) error {
	if errors.Is(err, redis.ErrNicknameTaken) {
		return ErrNicknameTaken
	}
	if errors.Is(err, redis.ErrParticipantNotFound) {
		return ErrParticipantNotFound
	}
	if errors.Is(err, redis.ErrInvalidNickname) {
		return ErrInvalidParams
	}

	return ErrRuntimeStoreUnavailable
}

func (s *Service) mapAnswerRepositoryError(err error) error {
	if errors.Is(err, redis.ErrAnswerAlreadySubmitted) {
		return ErrAnswerAlreadySubmitted
	}

	return ErrRuntimeStoreUnavailable
}

func (s *Service) mapLeaderboardRepositoryError(err error) error {
	if errors.Is(err, redis.ErrLeaderboardEntryNotFound) {
		return ErrParticipantNotFound
	}

	return ErrRuntimeStoreUnavailable
}

func (s *Service) buildSessionSnapshot(
	runtime domain.SessionRuntime,
	quiz domain.QuizSnapshot,
	participants []domain.RuntimeParticipant,
	leaderboardTop []SnapshotLeaderboardEntryDTO,
) SnapshotDTO {
	dto := SnapshotDTO{
		SessionID:            runtime.SessionID,
		RoomCode:             runtime.RoomCode,
		Status:               string(runtime.Status),
		CurrentQuestionIndex: runtime.Progress.CurrentQuestionIndex,
		TotalQuestions:       runtime.Progress.TotalQuestions,
		DeadlineAt:           runtime.Progress.DeadlineAt,
		RevealUntil:          runtime.Progress.RevealUntil,
		LeaderboardTop:       leaderboardTop,
	}

	dto.Participants = make([]SnapshotParticipantDTO, len(participants))
	for i, p := range participants {
		dto.Participants[i] = SnapshotParticipantDTO{
			ParticipantID: p.ParticipantID,
			Nickname:      p.Nickname,
			Score:         p.Score,
			Rank:          p.Rank,
			Connected:     p.Connected,
		}
	}

	currIdx := runtime.Progress.CurrentQuestionIndex
	if (runtime.Status == domain.RuntimeStatusQuestionOpen || runtime.Status == domain.RuntimeStatusAnswerReveal) &&
		currIdx >= 0 &&
		currIdx < len(quiz.Questions) {
		q := quiz.Questions[currIdx]

		dto.CurrentQuestion = &SnapshotQuestionDTO{
			ID:            q.ID,
			Text:          q.Text,
			SelectionType: string(q.SelectionType),
			Options:       make([]SnapshotQuestionOptionDTO, len(q.Options)),
		}

		for i, opt := range q.Options {
			dto.CurrentQuestion.Options[i] = SnapshotQuestionOptionDTO{
				ID:   opt.ID,
				Text: opt.Text,
			}
		}

		if runtime.Status == domain.RuntimeStatusAnswerReveal && runtime.Progress.RevealUntil != nil {
			correctIDs := s.collectCorrectOptionIDs(q)
			dto.CurrentQuestionReveal = &SnapshotQuestionRevealDTO{
				QuestionID:       q.ID,
				CorrectOptionIDs: correctIDs,
				RevealDuration:   int(s.revealDuration.Seconds()),
				RevealUntil:      *runtime.Progress.RevealUntil,
			}
		}
	}

	return dto
}
