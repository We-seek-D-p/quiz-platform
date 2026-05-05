package session

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/management"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
	"github.com/google/uuid"
)

type Service struct {
	managementRepository  ManagementRepository
	runtimeRepository     RuntimeRepository
	roomCodeRepository    RoomCodeRepository
	roomCodeGenerator     RoomCodeGenerator
	participantRepository ParticipantRepository
	answersRepository     AnswerRepository
	leaderboardRepository LeaderboardRepository
	revealDuration        time.Duration
}

func NewService(
	managementRepository ManagementRepository,
	runtimeRepository RuntimeRepository,
	roomCodeRepository RoomCodeRepository,
	roomCodeGenerator RoomCodeGenerator,
	participantRepository ParticipantRepository,
	answersRepository AnswerRepository,
	leaderboardRepository LeaderboardRepository,
	revealDuration time.Duration,
) *Service {
	return &Service{
		managementRepository:  managementRepository,
		runtimeRepository:     runtimeRepository,
		roomCodeRepository:    roomCodeRepository,
		roomCodeGenerator:     roomCodeGenerator,
		participantRepository: participantRepository,
		answersRepository:     answersRepository,
		leaderboardRepository: leaderboardRepository,
		revealDuration:        revealDuration,
	}
}

func (s *Service) InitSession(ctx context.Context, cmd InitSessionParams) (InitSessionResult, error) {
	if cmd.SessionID == "" || cmd.QuizID == "" || cmd.HostID == "" || cmd.IdempotencyKey == "" || cmd.CreatedAt.IsZero() {
		return InitSessionResult{}, ErrInvalidParams
	}

	existing, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
	if err == nil {
		if existing.QuizID != cmd.QuizID || existing.HostID != cmd.HostID {
			return InitSessionResult{}, ErrSessionRuntimeConflict
		}
		return InitSessionResult{Runtime: existing, Created: false}, nil
	}

	if !errors.Is(err, redis.ErrSessionNotFound) {
		return InitSessionResult{}, ErrRuntimeStoreUnavailable
	}

	bootstrap, err := s.managementRepository.GetSessionBootstrap(ctx, cmd.SessionID)
	if err != nil {
		return InitSessionResult{}, s.mapManagementError(err)
	}

	if bootstrap.SessionID != cmd.SessionID || bootstrap.QuizID != cmd.QuizID || bootstrap.HostID != cmd.HostID {
		return InitSessionResult{}, ErrSessionRuntimeConflict
	}

	var reservedCode string
	for i := 0; i < 50; i++ {
		code := s.roomCodeGenerator.Generate()
		ok, err := s.roomCodeRepository.Reserve(ctx, code, cmd.SessionID)
		if err != nil {
			return InitSessionResult{}, ErrRuntimeStoreUnavailable
		}
		if ok {
			reservedCode = code
			break
		}
	}
	if reservedCode == "" {
		return InitSessionResult{}, ErrRoomCodeUnavailable
	}

	runtime := domain.SessionRuntime{
		SessionID:     cmd.SessionID,
		QuizID:        bootstrap.QuizID,
		HostID:        bootstrap.HostID,
		RoomCode:      reservedCode,
		Status:        domain.RuntimeStatusLobby,
		InitializedAt: time.Now().UTC(),
		Progress: domain.RuntimeProgress{
			CurrentQuestionIndex: -1,
			TotalQuestions:       len(bootstrap.Quiz.Questions),
		},
	}

	err = s.runtimeRepository.Create(ctx, runtime, bootstrap.Quiz)
	if err != nil {
		_ = s.roomCodeRepository.Release(ctx, reservedCode)
		return InitSessionResult{}, s.mapRedisError(err)
	}

	return InitSessionResult{Runtime: runtime, Created: true}, nil
}

func (s *Service) GetSessionRuntime(ctx context.Context, cmd GetSessionRuntimeParams) (domain.SessionRuntime, error) {
	if cmd.SessionID == "" {
		return domain.SessionRuntime{}, ErrInvalidParams
	}

	res, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
	if err != nil {
		if errors.Is(err, redis.ErrSessionNotFound) {
			return domain.SessionRuntime{}, ErrSessionRuntimeNotFound
		}
		return domain.SessionRuntime{}, ErrRuntimeStoreUnavailable
	}

	return res, nil
}

func (s *Service) DeleteSessionRuntime(ctx context.Context, cmd DeleteSessionRuntimeParams) error {
	if cmd.SessionID == "" {
		return ErrInvalidParams
	}

	err := s.runtimeRepository.Delete(ctx, cmd.SessionID)
	if err != nil {
		if errors.Is(err, redis.ErrSessionNotFound) {
			return nil
		}
		return ErrRuntimeStoreUnavailable
	}

	return nil
}

func (s *Service) HostConnect(ctx context.Context, cmd HostConnectParams) (HostConnectResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	hostUserID := strings.TrimSpace(cmd.HostUserID)

	if sessionID == "" || hostUserID == "" {
		return HostConnectResult{}, ErrInvalidParams
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return HostConnectResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.HostID != hostUserID {
		return HostConnectResult{}, ErrForbidden
	}

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return HostConnectResult{}, ErrInternal
	}

	return HostConnectResult{
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants),
	}, nil
}

func (s *Service) PlayerJoin(ctx context.Context, cmd PlayerJoinParams) (PlayerJoinResult, error) {
	roomCode := strings.TrimSpace(cmd.RoomCode)
	nickname := strings.TrimSpace(cmd.Nickname)

	if roomCode == "" || nickname == "" {
		return PlayerJoinResult{}, ErrInvalidParams
	}

	sessionID, err := s.roomCodeRepository.GetSessionID(ctx, roomCode)
	if err != nil {
		return PlayerJoinResult{}, ErrRoomNotFound
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return PlayerJoinResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		return PlayerJoinResult{}, ErrGameAlreadyFinished
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
		return PlayerJoinResult{}, err
	}

	rank, err := s.leaderboardRepository.AddScore(ctx, sessionID, pID, 0)
	if err == nil {
		_ = s.participantRepository.UpdateScoreAndRank(ctx, sessionID, pID, 0, rank)
		participant.Rank = rank
	}

	allParticipants, _ := s.participantRepository.List(ctx, sessionID)

	return PlayerJoinResult{
		JoinedLobby: JoinedLobbyDTO{
			ParticipantID:    pID,
			ParticipantToken: token,
			Nickname:         nickname,
			RoomCode:         roomCode,
			Status:           string(snapshot.Runtime.Status),
		},
		LobbyUpdated: LobbyUpdatedDTO{
			PlayersCount: len(allParticipants),
		},
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, allParticipants),
	}, nil
}

func (s *Service) PlayerReconnect(ctx context.Context, cmd PlayerReconnectParams) (PlayerReconnectResult, error) {
	roomCode := strings.TrimSpace(cmd.RoomCode)
	token := strings.TrimSpace(cmd.ParticipantToken)

	if roomCode == "" || token == "" {
		return PlayerReconnectResult{}, ErrInvalidParams
	}

	sessionID, err := s.roomCodeRepository.GetSessionID(ctx, roomCode)
	if err != nil {
		return PlayerReconnectResult{}, ErrRoomNotFound
	}

	participant, err := s.participantRepository.GetByToken(ctx, sessionID, token)
	if err != nil {
		return PlayerReconnectResult{}, ErrInvalidParticipantToken
	}

	_ = s.participantRepository.SetConnected(ctx, sessionID, participant.ParticipantID, true)

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return PlayerReconnectResult{}, s.mapRedisError(err)
	}

	allParticipants, _ := s.participantRepository.List(ctx, sessionID)

	return PlayerReconnectResult{
		SessionSnapshot: s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, allParticipants),
	}, nil
}

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

	participants, _ := s.participantRepository.List(ctx, sessionID)
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

func (s *Service) calculateDeadline(start time.Time, seconds int) *time.Time {
	deadline := start.Add(time.Duration(seconds) * time.Second)
	return &deadline
}

func (s *Service) SubmitAnswer(ctx context.Context, cmd SubmitAnswerParams) (SubmitAnswerResult, error) {
	sessionID := strings.TrimSpace(cmd.SessionID)
	pID := strings.TrimSpace(cmd.ParticipantID)
	qID := strings.TrimSpace(cmd.QuestionID)

	if sessionID == "" || pID == "" || qID == "" {
		return SubmitAnswerResult{}, ErrInvalidParams
	}
	if len(cmd.SelectedOptionIDs) == 0 {
		return SubmitAnswerResult{}, ErrInvalidAnswerPayload
	}

	seen := make(map[string]struct{}, len(cmd.SelectedOptionIDs))
	for _, optionID := range cmd.SelectedOptionIDs {
		normalized := strings.TrimSpace(optionID)
		if normalized == "" {
			return SubmitAnswerResult{}, ErrInvalidAnswerPayload
		}
		if _, exists := seen[normalized]; exists {
			return SubmitAnswerResult{}, ErrInvalidAnswerPayload
		}
		seen[normalized] = struct{}{}
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return SubmitAnswerResult{}, s.mapRedisError(err)
	}

	if snapshot.Runtime.Status != domain.RuntimeStatusQuestionOpen {
		return SubmitAnswerResult{}, ErrQuestionNotActive
	}

	currIdx := snapshot.Runtime.Progress.CurrentQuestionIndex
	if currIdx < 0 || currIdx >= len(snapshot.Quiz.Questions) {
		return SubmitAnswerResult{}, ErrQuestionNotActive
	}

	currentQuestion := snapshot.Quiz.Questions[currIdx]
	if currentQuestion.ID != qID {
		return SubmitAnswerResult{}, ErrQuestionNotActive
	}

	if err := s.validateAnswerPayload(currentQuestion, cmd.SelectedOptionIDs); err != nil {
		return SubmitAnswerResult{}, err
	}

	answer := domain.RuntimeAnswer{
		ParticipantID:     pID,
		SelectedOptionIDs: cmd.SelectedOptionIDs,
		SubmittedAt:       time.Now().UTC(),
	}
	if err := s.answersRepository.SubmitOnce(ctx, sessionID, qID, answer); err != nil {
		return SubmitAnswerResult{}, ErrAnswerAlreadySubmitted
	}

	delta := 0
	if s.checkIsCorrect(currentQuestion, cmd.SelectedOptionIDs) {
		delta = 1
	}

	newScore, _ := s.leaderboardRepository.AddScore(ctx, sessionID, pID, delta)
	rank, _ := s.leaderboardRepository.GetRank(ctx, sessionID, pID)
	_ = s.participantRepository.UpdateScoreAndRank(ctx, sessionID, pID, newScore, rank)

	answered, _ := s.answersRepository.ListByQuestion(ctx, sessionID, qID)
	participants, _ := s.participantRepository.List(ctx, sessionID)

	return SubmitAnswerResult{
		AnswerAccepted: AnswerAcceptedDTO{
			QuestionID: qID,
			AcceptedAt: answer.SubmittedAt,
		},
		HostProgress: &QuestionProgressDTO{
			QuestionID:    qID,
			AnsweredCount: len(answered),
			TotalPlayers:  len(participants),
		},
	}, nil
}

func (s *Service) validateAnswerPayload(q domain.QuestionSnapshot, selected []string) error {
	validIDs := make(map[string]struct{})
	for _, opt := range q.Options {
		validIDs[opt.ID] = struct{}{}
	}

	for _, id := range selected {
		if _, ok := validIDs[id]; !ok {
			return ErrInvalidAnswerPayload
		}
	}

	st := string(q.SelectionType)
	if st == "single" && len(selected) != 1 {
		return ErrSelectionCountInvalid
	}
	if st == "multiple" && len(selected) < 1 {
		return ErrSelectionCountInvalid
	}
	return nil
}

func (s *Service) checkIsCorrect(q domain.QuestionSnapshot, selected []string) bool {
	var correctIDs []string
	for _, opt := range q.Options {
		if opt.IsCorrect {
			correctIDs = append(correctIDs, opt.ID)
		}
	}

	if len(correctIDs) != len(selected) {
		return false
	}

	selectedMap := make(map[string]struct{}, len(selected))
	for _, id := range selected {
		selectedMap[id] = struct{}{}
	}

	for _, id := range correctIDs {
		if _, ok := selectedMap[id]; !ok {
			return false
		}
	}
	return true
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

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		participants, _ := s.participantRepository.List(ctx, sessionID)
		return FinishGameResult{
			SessionFinished: FinishedDTO{
				LeaderboardTop: s.mapParticipantsToLeaderboard(participants),
			},
			PersistedStatus: string(domain.PersistedStatusFinished),
			PersistedAt:     *snapshot.Runtime.Progress.FinishedAt,
		}, nil
	}

	now := time.Now().UTC()
	snapshot.Runtime.Status = domain.RuntimeStatusFinished
	snapshot.Runtime.Progress.FinishedAt = &now
	snapshot.Runtime.Progress.DeadlineAt = nil
	snapshot.Runtime.Progress.RevealUntil = nil

	participants, err := s.participantRepository.List(ctx, sessionID)
	if err != nil {
		return FinishGameResult{}, ErrInternal
	}

	eventID := uuid.NewString()
	results := domain.SessionResults{
		EventID:      eventID,
		FinishReason: "manual",
		FinishedAt:   now,
		Participants: s.mapToManagementResults(participants),
	}

	if err := s.managementRepository.ReportSessionResults(ctx, sessionID, results); err != nil {
		return FinishGameResult{}, s.mapManagementError(err)
	}

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return FinishGameResult{}, s.mapRedisError(err)
	}

	return FinishGameResult{
		SessionFinished: FinishedDTO{
			LeaderboardTop: s.mapParticipantsToLeaderboard(participants),
		},
		PersistedStatus: string(domain.PersistedStatusFinished),
		PersistedAt:     now,
	}, nil
}

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

func (s *Service) mapParticipantsToLeaderboard(ps []domain.RuntimeParticipant) []SnapshotLeaderboardEntryDTO {
	res := make([]SnapshotLeaderboardEntryDTO, len(ps))
	for i, p := range ps {
		res[i] = SnapshotLeaderboardEntryDTO{
			ParticipantID: p.ParticipantID,
			Nickname:      p.Nickname,
			Score:         p.Score,
			Rank:          p.Rank,
		}
	}
	return res
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
	if errors.Is(err, redis.ErrSessionConflict) {
		return ErrSessionRuntimeConflict
	}
	return ErrRuntimeStoreUnavailable
}

func (s *Service) buildSessionSnapshot(
	runtime domain.SessionRuntime,
	quiz domain.QuizSnapshot,
	participants []domain.RuntimeParticipant,
) SnapshotDTO {
	dto := SnapshotDTO{
		SessionID:            runtime.SessionID,
		RoomCode:             runtime.RoomCode,
		Status:               string(runtime.Status),
		CurrentQuestionIndex: runtime.Progress.CurrentQuestionIndex,
		TotalQuestions:       runtime.Progress.TotalQuestions,
		DeadlineAt:           runtime.Progress.DeadlineAt,
		RevealUntil:          runtime.Progress.RevealUntil,
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
			var correctIDs []string
			for _, opt := range q.Options {
				if opt.IsCorrect {
					correctIDs = append(correctIDs, opt.ID)
				}
			}
			dto.CurrentQuestionReveal = &SnapshotQuestionRevealDTO{
				QuestionID:       q.ID,
				CorrectOptionIDs: correctIDs,
				RevealUntil:      *runtime.Progress.RevealUntil,
			}
		}
	}

	return dto
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

	now := time.Now().UTC()
	revealUntil := now.Add(s.revealDuration)

	snapshot.Runtime.Status = domain.RuntimeStatusAnswerReveal
	snapshot.Runtime.Progress.RevealUntil = &revealUntil
	snapshot.Runtime.Progress.DeadlineAt = nil

	if err := s.runtimeRepository.UpdateRuntime(ctx, snapshot.Runtime); err != nil {
		return SnapshotDTO{}, s.mapRedisError(err)
	}

	participants, _ := s.participantRepository.List(ctx, sessionID)
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
		return SnapshotDTO{}, ErrInvalidStateTransition
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

	participants, _ := s.participantRepository.List(ctx, sessionID)
	return s.buildSessionSnapshot(snapshot.Runtime, snapshot.Quiz, participants), nil
}
