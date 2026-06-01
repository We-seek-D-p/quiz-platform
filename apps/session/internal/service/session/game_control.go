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
		return StartGameResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return StartGameResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return StartGameResult{}, domain.NewForbidden("forbidden", "forbidden", nil)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusLobby {
		if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
			return StartGameResult{}, domain.NewConflict("game_already_finished", "game already finished", nil)
		}
		return StartGameResult{}, domain.NewConflict("game_already_started", "game already started", nil)
	}

	if len(snapshot.Quiz.Questions) == 0 {
		return StartGameResult{}, domain.NewInternal("internal_error", "internal error", nil)
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

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return StartGameResult{}, err
	}

	sessionSnapshot := s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop)

	return StartGameResult{
		QuestionOpened: QuestionOpened{
			QuestionIndex:  0,
			TotalQuestions: len(snapshot.Quiz.Questions),
			Question:       *sessionSnapshot.CurrentQuestion,
			DeadlineAt:     *snapshot.Runtime.Progress.DeadlineAt,
		},
		SessionSnapshot:  sessionSnapshot,
		PersistedStatus:  string(domain.PersistedStatusInProgress),
		PersistedEventID: eventID,
	}, nil
}

func (s *Service) FinishGame(ctx context.Context, cmd FinishGameParams) (FinishGameResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	hostUserID := strings.TrimSpace(cmd.HostUserID)

	if sessionID == "" || hostUserID == "" {
		return FinishGameResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return FinishGameResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return FinishGameResult{}, domain.NewForbidden("forbidden", "forbidden", nil)
	}

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		return FinishGameResult{}, domain.NewConflict("game_already_finished", "game already finished", nil)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusLobby &&
		snapshot.Runtime.Status != domain.RuntimeStatusQuestionOpen &&
		snapshot.Runtime.Status != domain.RuntimeStatusAnswerReveal {
		return FinishGameResult{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	_, participants, persistedAt, err := s.finishSession(ctx, snapshot, domain.FinishReasonManual)
	if err != nil {
		return FinishGameResult{}, err
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, len(participants))
	if err != nil {
		return FinishGameResult{}, err
	}

	return FinishGameResult{
		SessionFinished: Finished{
			LeaderboardTop: leaderboardTop,
		},
		PersistedStatus: string(domain.PersistedStatusFinished),
		PersistedAt:     persistedAt,
	}, nil
}

func (s *Service) finishSession(
	ctx context.Context,
	snapshot domain.SessionSnapshot,
	finishReason domain.FinishReason,
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

func (s *Service) CloseCurrentQuestionAndBuildReveal(ctx context.Context, sessionID string) (RevealTransitionResult, error) {
	sessionID = strings.TrimSpace(sessionID)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return RevealTransitionResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusQuestionOpen {
		return RevealTransitionResult{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	currIdx := snapshot.Runtime.Progress.CurrentQuestionIndex
	if currIdx < 0 || currIdx >= len(snapshot.Quiz.Questions) {
		return RevealTransitionResult{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	now := time.Now().UTC()
	snapshot.Runtime.Status = domain.RuntimeStatusAnswerReveal
	snapshot.Runtime.Progress.RevealUntil = new(time.Time)
	*snapshot.Runtime.Progress.RevealUntil = now.Add(s.answerRevealDuration)
	snapshot.Runtime.Progress.DeadlineAt = nil

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return RevealTransitionResult{}, s.mapRedisError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return RevealTransitionResult{}, s.mapParticipantRepositoryError(err)
	}

	answers, err := s.answersRepository.ListByQuestion(ctx, sessionID, snapshot.Quiz.Questions[currIdx].ID)
	if err != nil {
		return RevealTransitionResult{}, s.mapAnswerRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return RevealTransitionResult{}, err
	}

	sessionSnapshot := s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop)
	revealResult := s.buildRevealTransitionResult(snapshot, participants, answers, leaderboardTop)
	revealResult.SessionSnapshot = sessionSnapshot

	return revealResult, nil
}

func (s *Service) buildRevealTransitionResult(
	snapshot domain.SessionSnapshot,
	participants []domain.RuntimeParticipant,
	answers []domain.RuntimeAnswer,
	leaderboardTop []SnapshotLeaderboardEntry,
) RevealTransitionResult {
	currQuestion := snapshot.Quiz.Questions[snapshot.Runtime.Progress.CurrentQuestionIndex]
	correctOptionIDs := s.collectCorrectOptionIDs(currQuestion)
	answerByParticipant := make(map[string]domain.RuntimeAnswer, len(answers))

	for _, answer := range answers {
		answerByParticipant[answer.ParticipantID] = answer
	}

	playerReveals := make([]ParticipantAnswerReveal, 0, len(participants))
	for _, participant := range participants {
		participantAnswer, hasAnswer := answerByParticipant[participant.ParticipantID]
		var yourSelectedOptionIDs []string
		yourResult := "wrong"
		scoreDelta := 0

		if hasAnswer {
			yourSelectedOptionIDs = participantAnswer.SelectedOptionIDs
			yourResult = participantAnswer.Result
			scoreDelta = participantAnswer.ScoreDelta
		}

		playerReveals = append(playerReveals, ParticipantAnswerReveal{
			ParticipantID: participant.ParticipantID,
			Payload: AnswerReveal{
				QuestionID:            currQuestion.ID,
				CorrectOptionIDs:      correctOptionIDs,
				YourSelectedOptionIDs: yourSelectedOptionIDs,
				YourResult:            yourResult,
				ScoreDelta:            scoreDelta,
				TotalScore:            participant.Score,
				YourRank:              participant.Rank,
				LeaderboardTop:        leaderboardTop,
				RevealDurationSec:     int(s.answerRevealDuration.Seconds()),
				RevealUntil:           *snapshot.Runtime.Progress.RevealUntil,
			},
		})
	}

	return RevealTransitionResult{
		HostReveal: QuestionRevealHost{
			QuestionID:        currQuestion.ID,
			CorrectOptionIDs:  correctOptionIDs,
			AnsweredCount:     len(answers),
			TotalPlayers:      len(participants),
			LeaderboardTop:    leaderboardTop,
			RevealDurationSec: int(s.answerRevealDuration.Seconds()),
			RevealUntil:       *snapshot.Runtime.Progress.RevealUntil,
		},
		PlayerReveals: playerReveals,
	}
}

func (s *Service) AdvanceToLeaderboardReveal(ctx context.Context, sessionID string) (LeaderboardTransitionResult, error) {
	sessionID = strings.TrimSpace(sessionID)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return LeaderboardTransitionResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusAnswerReveal {
		return LeaderboardTransitionResult{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	currIdx := snapshot.Runtime.Progress.CurrentQuestionIndex
	if currIdx < 0 || currIdx >= len(snapshot.Quiz.Questions) {
		return LeaderboardTransitionResult{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	now := time.Now().UTC()
	snapshot.Runtime.Status = domain.RuntimeStatusLeaderboardReveal
	snapshot.Runtime.Progress.RevealUntil = new(time.Time)
	*snapshot.Runtime.Progress.RevealUntil = now.Add(s.leaderboardRevealDuration)

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return LeaderboardTransitionResult{}, s.mapRedisError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return LeaderboardTransitionResult{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return LeaderboardTransitionResult{}, err
	}

	sessionSnapshot := s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop)
	currQuestion := snapshot.Quiz.Questions[currIdx]

	playerReveals := make([]ParticipantLeaderboardReveal, 0, len(participants))
	for _, participant := range participants {
		playerReveals = append(playerReveals, ParticipantLeaderboardReveal{
			ParticipantID: participant.ParticipantID,
			Payload: LeaderboardReveal{
				QuestionID:        currQuestion.ID,
				LeaderboardTop:    leaderboardTop,
				YourScore:         participant.Score,
				YourRank:          participant.Rank,
				RevealDurationSec: int(s.leaderboardRevealDuration.Seconds()),
				RevealUntil:       *snapshot.Runtime.Progress.RevealUntil,
			},
		})
	}

	return LeaderboardTransitionResult{
		SessionSnapshot: sessionSnapshot,
		HostReveal: LeaderboardRevealHost{
			QuestionID:        currQuestion.ID,
			LeaderboardTop:    leaderboardTop,
			RevealDurationSec: int(s.leaderboardRevealDuration.Seconds()),
			RevealUntil:       *snapshot.Runtime.Progress.RevealUntil,
		},
		PlayerReveals: playerReveals,
	}, nil
}

func (s *Service) AdvanceAfterLeaderboardReveal(ctx context.Context, sessionID string) (Snapshot, error) {
	sessionID = strings.TrimSpace(sessionID)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return Snapshot{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusLeaderboardReveal {
		return Snapshot{}, domain.NewConflict("invalid_state_transition", "invalid state transition", nil)
	}

	nextIndex := snapshot.Runtime.Progress.CurrentQuestionIndex + 1
	if nextIndex >= len(snapshot.Quiz.Questions) {
		runtime, participants, _, err := s.finishSession(ctx, snapshot, domain.FinishReasonCompleted)
		if err != nil {
			return Snapshot{}, err
		}

		leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, len(participants))
		if err != nil {
			return Snapshot{}, err
		}

		return s.buildSessionSnapshot(runtime, snapshot.Quiz, participants, leaderboardTop), nil
	}

	now := time.Now().UTC()
	nextQuestion := snapshot.Quiz.Questions[nextIndex]

	snapshot.Runtime.Status = domain.RuntimeStatusQuestionOpen
	snapshot.Runtime.Progress.CurrentQuestionIndex = nextIndex
	snapshot.Runtime.Progress.DeadlineAt = s.calculateDeadline(now, nextQuestion.TimeLimitSeconds)
	snapshot.Runtime.Progress.RevealUntil = nil

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return Snapshot{}, s.mapRedisError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return Snapshot{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return Snapshot{}, err
	}

	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop), nil
}
