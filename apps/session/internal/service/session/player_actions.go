package session

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
	"github.com/google/uuid"
)

func (s *Service) HostConnect(ctx context.Context, cmd HostConnectParams) (HostConnectResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	hostUserID := strings.TrimSpace(cmd.HostUserID)

	if sessionID == "" || hostUserID == "" {
		return HostConnectResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return HostConnectResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return HostConnectResult{}, domain.NewForbidden("forbidden", "forbidden", nil)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return HostConnectResult{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, participants, leaderboardTopLimit)
	if err != nil {
		return HostConnectResult{}, err
	}

	return HostConnectResult{
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants, leaderboardTop),
	}, nil
}

func (s *Service) PlayerJoin(ctx context.Context, cmd PlayerJoinParams) (PlayerJoinResult, error) {
	roomCode := strings.TrimSpace(cmd.RoomCode)
	nickname := strings.TrimSpace(cmd.Nickname)

	if roomCode == "" || nickname == "" {
		return PlayerJoinResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	sessionID, err := s.roomCodeRepository.GetSessionID(ctx, roomCode)
	if err != nil {
		return PlayerJoinResult{}, s.mapRoomCodeError(err)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return PlayerJoinResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		return PlayerJoinResult{}, domain.NewConflict("game_already_finished", "game already finished", nil)
	}

	pID := uuid.NewString()
	token := uuid.NewString()
	now := time.Now().UTC()

	participant := domain.RuntimeParticipant{
		ParticipantID:    pID,
		ParticipantToken: token,
		Nickname:         nickname,
		Score:            0,
		Rank:             0,
		Connected:        true,
		JoinedAt:         now,
		LastSeenAt:       &now,
	}

	if err := s.participantRepository.Create(ctx, sessionID, participant); err != nil {
		return PlayerJoinResult{}, s.mapParticipantRepositoryError(err)
	}

	newScore, err := s.leaderboardRepository.AddScore(ctx, sessionID, pID, 0)
	if err != nil {
		return PlayerJoinResult{}, s.mapLeaderboardRepositoryError(err)
	}

	rank, err := s.leaderboardRepository.GetRank(ctx, sessionID, pID)
	if err != nil {
		return PlayerJoinResult{}, s.mapLeaderboardRepositoryError(err)
	}

	if err := s.participantRepository.UpdateScoreAndRank(ctx, sessionID, pID, newScore, rank); err != nil {
		return PlayerJoinResult{}, s.mapParticipantRepositoryError(err)
	}

	participant.Score = newScore
	participant.Rank = rank

	allParticipants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return PlayerJoinResult{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, allParticipants, leaderboardTopLimit)
	if err != nil {
		return PlayerJoinResult{}, err
	}

	return PlayerJoinResult{
		JoinedLobby: JoinedLobby{
			ParticipantID:    pID,
			ParticipantToken: token,
			Nickname:         nickname,
			RoomCode:         roomCode,
			Status:           string(snapshot.Runtime.Status),
		},
		LobbyUpdated: LobbyUpdated{
			PlayersCount: len(allParticipants),
		},
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, allParticipants, leaderboardTop),
	}, nil
}

func (s *Service) PlayerReconnect(ctx context.Context, cmd PlayerReconnectParams) (PlayerReconnectResult, error) {
	roomCode := strings.TrimSpace(cmd.RoomCode)
	token := strings.TrimSpace(cmd.ParticipantToken)

	if roomCode == "" || token == "" {
		return PlayerReconnectResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	sessionID, err := s.roomCodeRepository.GetSessionID(ctx, roomCode)
	if err != nil {
		return PlayerReconnectResult{}, s.mapRoomCodeError(err)
	}

	participant, err := s.participantRepository.GetByToken(ctx, sessionID, token)
	if err != nil {
		if errors.Is(err, redis.ErrParticipantNotFound) {
			return PlayerReconnectResult{}, domain.NewInvalidInput("invalid_participant_token", "invalid participant token", err)
		}

		return PlayerReconnectResult{}, s.mapParticipantRepositoryError(err)
	}

	if err := s.participantRepository.SetConnected(ctx, sessionID, participant.ParticipantID, true); err != nil {
		return PlayerReconnectResult{}, s.mapParticipantRepositoryError(err)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return PlayerReconnectResult{}, s.mapRedisError(err)
	}

	allParticipants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return PlayerReconnectResult{}, s.mapParticipantRepositoryError(err)
	}

	leaderboardTop, err := s.loadLeaderboardTop(ctx, sessionID, allParticipants, leaderboardTopLimit)
	if err != nil {
		return PlayerReconnectResult{}, err
	}

	return PlayerReconnectResult{
		ParticipantID:   participant.ParticipantID,
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, allParticipants, leaderboardTop),
	}, nil
}

// SetParticipantConnected updates runtime presence for a connected player.
func (s *Service) SetParticipantConnected(ctx context.Context, sessionID, participantID string, connected bool) error {
	sessionID = strings.TrimSpace(sessionID)
	participantID = strings.TrimSpace(participantID)
	if sessionID == "" || participantID == "" {
		return domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	if err := s.participantRepository.SetConnected(ctx, sessionID, participantID, connected); err != nil {
		return s.mapParticipantRepositoryError(err)
	}

	return nil
}

func (s *Service) SubmitAnswer(ctx context.Context, cmd SubmitAnswerParams) (SubmitAnswerResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	pID := strings.TrimSpace(cmd.ParticipantID)
	qID := strings.TrimSpace(cmd.QuestionID)

	if sessionID == "" || pID == "" || qID == "" {
		return SubmitAnswerResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}
	if len(cmd.SelectedOptionIDs) == 0 {
		return SubmitAnswerResult{}, domain.NewInvalidInput("invalid_answer_payload", "invalid answer payload", nil)
	}

	seen := make(map[string]struct{}, len(cmd.SelectedOptionIDs))
	for _, optionID := range cmd.SelectedOptionIDs {
		normalized := strings.TrimSpace(optionID)
		if normalized == "" {
			return SubmitAnswerResult{}, domain.NewInvalidInput("option_not_in_question", "option not found", nil)
		}
		if _, exists := seen[normalized]; exists {
			return SubmitAnswerResult{}, domain.NewInvalidInput("selection_count_invalid", "selection count invalid", nil)
		}
		seen[normalized] = struct{}{}
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return SubmitAnswerResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusQuestionOpen {
		return SubmitAnswerResult{}, domain.NewConflict("question_not_active", "question not active", nil)
	}

	currIdx := snapshot.Runtime.Progress.CurrentQuestionIndex
	if currIdx < 0 || currIdx >= len(snapshot.Quiz.Questions) {
		return SubmitAnswerResult{}, domain.NewConflict("question_not_active", "question not active", nil)
	}

	currentQuestion := snapshot.Quiz.Questions[currIdx]
	if currentQuestion.ID != qID {
		return SubmitAnswerResult{}, domain.NewConflict("question_not_active", "question not active", nil)
	}

	if err := s.validateAnswerPayload(currentQuestion, cmd.SelectedOptionIDs); err != nil {
		return SubmitAnswerResult{}, err
	}

	if _, err := s.participantRepository.GetByID(ctx, sessionID, pID); err != nil {
		return SubmitAnswerResult{}, s.mapParticipantRepositoryError(err)
	}

	now := time.Now().UTC()
	answer := domain.RuntimeAnswer{
		ParticipantID:     pID,
		SelectedOptionIDs: cmd.SelectedOptionIDs,
		SubmittedAt:       now,
	}

	result := "wrong"
	if s.checkIsCorrect(currentQuestion, cmd.SelectedOptionIDs) {
		result = "correct"
	}

	openedAt := now
	if snapshot.Runtime.Progress.DeadlineAt != nil {
		openedAt = snapshot.Runtime.Progress.DeadlineAt.Add(-time.Duration(currentQuestion.TimeLimitSeconds) * time.Second)
	}
	delta := domain.CalculatePoints(currentQuestion, cmd.SelectedOptionIDs, openedAt, now)
	if delta > 0 {
		result = "correct"
	}

	answer.Result = result
	answer.ScoreDelta = delta

	if err := s.answersRepository.SubmitOnce(ctx, sessionID, qID, answer); err != nil {
		return SubmitAnswerResult{}, s.mapAnswerRepositoryError(err)
	}

	newScore, err := s.leaderboardRepository.AddScore(ctx, sessionID, pID, delta)
	if err != nil {
		return SubmitAnswerResult{}, s.mapLeaderboardRepositoryError(err)
	}

	rank, err := s.leaderboardRepository.GetRank(ctx, sessionID, pID)
	if err != nil {
		return SubmitAnswerResult{}, s.mapLeaderboardRepositoryError(err)
	}

	if err := s.participantRepository.UpdateScoreAndRank(ctx, sessionID, pID, newScore, rank); err != nil {
		return SubmitAnswerResult{}, s.mapParticipantRepositoryError(err)
	}

	answered, err := s.answersRepository.ListByQuestion(ctx, sessionID, qID)
	if err != nil {
		return SubmitAnswerResult{}, s.mapAnswerRepositoryError(err)
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return SubmitAnswerResult{}, s.mapParticipantRepositoryError(err)
	}

	return SubmitAnswerResult{
		AnswerAccepted: AnswerAccepted{
			QuestionID: qID,
			AcceptedAt: answer.SubmittedAt,
		},
		HostProgress: &QuestionProgress{
			QuestionID:    qID,
			AnsweredCount: len(answered),
			TotalPlayers:  len(participants),
		},
	}, nil
}
