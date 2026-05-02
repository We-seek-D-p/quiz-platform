package session

import (
	"context"
	"errors"
	"time"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/management"
	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/repository/redis"
)

type Service struct {
	managementRepository ManagementRepository
	runtimeRepository    RuntimeRepository
	roomCodeRepository   RoomCodeRepository
	roomCodeGenerator    RoomCodeGenerator
}

func NewService(
	managementRepository ManagementRepository,
	runtimeRepository RuntimeRepository,
	roomCodeRepository RoomCodeRepository,
	roomCodeGenerator RoomCodeGenerator,
) *Service {
	return &Service{
		managementRepository: managementRepository,
		runtimeRepository:    runtimeRepository,
		roomCodeRepository:   roomCodeRepository,
		roomCodeGenerator:    roomCodeGenerator,
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
