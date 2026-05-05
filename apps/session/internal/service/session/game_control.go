package session

import (
	"context"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/google/uuid"
)

func (s *Service) StartGame(ctx context.Context, cmd StartGameParams) (StartGameResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	hostUserID := strings.TrimSpace(cmd.HostUserID)

	if sessionID == "" || hostUserID == "" {
		return StartGameResult{}, ErrInvalidParams
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return StartGameResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return StartGameResult{}, ErrForbidden
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusLobby {
		if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
			return StartGameResult{}, ErrGameAlreadyFinished
		}
		return StartGameResult{}, ErrGameAlreadyStarted
	}

	if len(snapshot.Quiz.Questions) == 0 {
		return StartGameResult{}, ErrInternal
	}

	now := time.Now().UTC()
	firstQuestion := snapshot.Quiz.Questions[0]

	snapshot.Runtime.Status = domain.RuntimeStatusQuestionOpen
	snapshot.Runtime.Progress.CurrentQuestionIndex = 0
	snapshot.Runtime.Progress.StartedAt = &now
	snapshot.Runtime.Progress.DeadlineAt = s.calculateDeadline(now, firstQuestion.TimeLimitSeconds)

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return StartGameResult{}, s.mapRedisError(err)
	}

	eventID := uuid.NewString()
	err = s.managementRepository.ReportSessionStatus(ctx, sessionID, domain.SessionStatusUpdate{
		Status:    domain.PersistedStatusInProgress,
		StartedAt: &now,
		EventID:   eventID,
	})
	if err != nil {
		return StartGameResult{}, s.mapManagementError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return StartGameResult{}, s.mapParticipantRepositoryError(err)
	}
	snapshotDTO := s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants)

	return StartGameResult{
		QuestionOpened: QuestionOpenedDTO{
			QuestionIndex:  0,
			TotalQuestions: len(snapshot.Quiz.Questions),
			Question:       *snapshotDTO.CurrentQuestion,
			DeadlineAt:     *snapshot.Runtime.Progress.DeadlineAt,
		},
		SessionSnapshot:  snapshotDTO,
		PersistedStatus:  string(domain.PersistedStatusInProgress),
		PersistedEventID: eventID,
	}, nil
}

func (s *Service) FinishGame(ctx context.Context, cmd FinishGameParams) (FinishGameResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	hostUserID := strings.TrimSpace(cmd.HostUserID)

	if sessionID == "" || hostUserID == "" {
		return FinishGameResult{}, ErrInvalidParams
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return FinishGameResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return FinishGameResult{}, ErrForbidden
	}

	_, participants, persistedAt, err := s.finishSession(ctx, snapshot, "manual")
	if err != nil {
		return FinishGameResult{}, err
	}

	return FinishGameResult{
		SessionFinished: FinishedDTO{
			LeaderboardTop: s.mapParticipantsToLeaderboard(participants),
		},
		PersistedStatus: string(domain.PersistedStatusFinished),
		PersistedAt:     persistedAt,
	}, nil
}

func (s *Service) finishSession(
	ctx context.Context,
	snapshot domain.SessionSnapshot,
	finishReason string,
) (domain.SessionRuntime, []domain.RuntimeParticipant, time.Time, error) {
	sessionID := snapshot.Runtime.SessionID

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		participants, err := s.participantRepository.List(ctx, sessionID)
		if err != nil {
			return domain.SessionRuntime{}, nil, time.Time{}, s.mapParticipantRepositoryError(err)
		}

		persistedAt := time.Now().UTC()
		if snapshot.Runtime.Progress.FinishedAt != nil {
			persistedAt = *snapshot.Runtime.Progress.FinishedAt
		}

		return snapshot.Runtime, participants, persistedAt, nil
	}

	now := time.Now().UTC()
	runtime := snapshot.Runtime
	runtime.Status = domain.RuntimeStatusFinished
	runtime.Progress.FinishedAt = &now
	runtime.Progress.DeadlineAt = nil
	runtime.Progress.RevealUntil = nil

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return domain.SessionRuntime{}, nil, time.Time{}, s.mapParticipantRepositoryError(err)
	}

	eventID := uuid.NewString()
	results := domain.SessionResults{
		EventID:      eventID,
		FinishReason: finishReason,
		FinishedAt:   now,
		Participants: s.mapToManagementResults(participants),
	}

	if err := s.managementRepository.ReportSessionResults(ctx, sessionID, results); err != nil {
		return domain.SessionRuntime{}, nil, time.Time{}, s.mapManagementError(err)
	}

	if err := s.runtimeRepository.UpdateRuntime(ctx, runtime); err != nil {
		return domain.SessionRuntime{}, nil, time.Time{}, s.mapRedisError(err)
	}

	return runtime, participants, now, nil
}

func (s *Service) CloseCurrentQuestionAndBuildReveal(ctx context.Context, sessionID string) (SnapshotDTO, error) {
	sessionID = strings.TrimSpace(sessionID)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return SnapshotDTO{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusQuestionOpen {
		return SnapshotDTO{}, ErrInvalidStateTransition
	}

	currIdx := snapshot.Runtime.Progress.CurrentQuestionIndex
	if currIdx < 0 || currIdx >= len(snapshot.Quiz.Questions) {
		return SnapshotDTO{}, ErrInvalidStateTransition
	}

	now := time.Now().UTC()
	snapshot.Runtime.Status = domain.RuntimeStatusAnswerReveal
	snapshot.Runtime.Progress.RevealUntil = new(time.Time)
	*snapshot.Runtime.Progress.RevealUntil = now.Add(s.revealDuration)
	snapshot.Runtime.Progress.DeadlineAt = nil

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return SnapshotDTO{}, s.mapRedisError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return SnapshotDTO{}, s.mapParticipantRepositoryError(err)
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants), nil
}

func (s *Service) AdvanceToNextQuestion(ctx context.Context, sessionID string) (SnapshotDTO, error) {
	sessionID = strings.TrimSpace(sessionID)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return SnapshotDTO{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusAnswerReveal {
		return SnapshotDTO{}, ErrInvalidStateTransition
	}

	nextIndex := snapshot.Runtime.Progress.CurrentQuestionIndex + 1
	if nextIndex >= len(snapshot.Quiz.Questions) {
		runtime, participants, _, err := s.finishSession(ctx, snapshot, "completed")
		if err != nil {
			return SnapshotDTO{}, err
		}

		return s.buildSessionSnapshot(runtime, snapshot.Quiz, participants), nil
	}

	now := time.Now().UTC()
	nextQuestion := snapshot.Quiz.Questions[nextIndex]

	snapshot.Runtime.Status = domain.RuntimeStatusQuestionOpen
	snapshot.Runtime.Progress.CurrentQuestionIndex = nextIndex
	snapshot.Runtime.Progress.DeadlineAt = s.calculateDeadline(now, nextQuestion.TimeLimitSeconds)
	snapshot.Runtime.Progress.RevealUntil = nil

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return SnapshotDTO{}, s.mapRedisError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return SnapshotDTO{}, s.mapParticipantRepositoryError(err)
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants), nil
}
