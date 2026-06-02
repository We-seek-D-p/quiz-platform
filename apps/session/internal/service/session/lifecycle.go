package session

import (
	"context"
	"strings"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

func (s *Service) InitSession(ctx context.Context, cmd InitSessionParams) (InitSessionResult, error) {
	if cmd.SessionID == "" || cmd.QuizID == "" || cmd.HostID == "" || cmd.IdempotencyKey == "" || cmd.CreatedAt.IsZero() {
		return InitSessionResult{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	existing, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
	if err == nil {
		if existing.QuizID != cmd.QuizID || existing.HostID != cmd.HostID {
			return InitSessionResult{}, domain.NewConflict("session_runtime_conflict", "session runtime conflict", nil)
		}
		return InitSessionResult{Runtime: existing, Created: false}, nil
	}

	if !isAppErrorCode(err, "session_runtime_not_found") {
		return InitSessionResult{}, err
	}

	bootstrap, err := s.managementRepository.GetSessionBootstrap(ctx, cmd.SessionID)
	if err != nil {
		return InitSessionResult{}, s.mapManagementError(err)
	}

	if bootstrap.SessionID != cmd.SessionID || bootstrap.QuizID != cmd.QuizID || bootstrap.HostID != cmd.HostID {
		return InitSessionResult{}, domain.NewConflict("session_runtime_conflict", "session runtime conflict", nil)
	}

	var reservedCode string
	for i := 0; i < 50; i++ {
		code := s.roomCodeGenerator.Generate()
		ok, err := s.roomCodeRepository.Reserve(ctx, code, cmd.SessionID)
		if err != nil {
			return InitSessionResult{}, domain.NewInternal("internal_error", "runtime storage unavailable", err)
		}
		if ok {
			reservedCode = code
			break
		}
	}
	if reservedCode == "" {
		return InitSessionResult{}, domain.NewInternal("room_code_unavailable", "room code unavailable", nil)
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
		return InitSessionResult{}, err
	}

	return InitSessionResult{Runtime: runtime, Created: true}, nil
}

func (s *Service) GetSessionRuntime(ctx context.Context, cmd GetSessionRuntimeParams) (domain.SessionRuntime, error) {
	if cmd.SessionID == "" {
		return domain.SessionRuntime{}, domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	res, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
	if err != nil {
		return domain.SessionRuntime{}, err
	}

	return res, nil
}

func (s *Service) DeleteSessionRuntime(ctx context.Context, cmd DeleteSessionRuntimeParams) error {
	if cmd.SessionID == "" {
		return domain.NewInvalidInput("invalid_payload", "invalid payload", nil)
	}

	err := s.runtimeRepository.Delete(ctx, cmd.SessionID)
	if err != nil {
		if isAppErrorCode(err, "session_runtime_not_found") {
			return nil
		}
		return err
	}

	return nil
}

// CheckSessionLiveness verifies that a session exists and still accepts socket clients.
func (s *Service) CheckSessionLiveness(ctx context.Context, sessionID string) error {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return domain.NewInvalidInput("invalid_payload", "session_id is required", nil)
	}

	snapshot, err := s.runtimeRepository.GetSnapshot(ctx, sessionID)
	if err != nil {
		return err
	}

	if snapshot.Runtime.Status == domain.RuntimeStatusFinished {
		return domain.NewConflict("game_already_finished", "game is already finished", nil)
	}

	return nil
}

// CheckRoomLiveness resolves a room code and verifies that its session is active.
func (s *Service) CheckRoomLiveness(ctx context.Context, roomCode string) error {
	roomCode = strings.TrimSpace(roomCode)
	if roomCode == "" {
		return domain.NewInvalidInput("invalid_payload", "room_code is required", nil)
	}

	sessionID, err := s.roomCodeRepository.GetSessionID(ctx, roomCode)
	if err != nil {
		return err
	}

	return s.CheckSessionLiveness(ctx, sessionID)
}
