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

func (s *Service) InitSession(ctx context.Context, cmd InitSessionParams) (domain.SessionRuntime, bool, error) {
    if cmd.SessionID == "" || cmd.QuizID == "" || cmd.HostID == "" || cmd.IdempotencyKey == "" || cmd.CreatedAt.IsZero() {
        return domain.SessionRuntime{}, false, ErrSessionRuntimeConflict
    }

    existing, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
    if err == nil {
        if existing.QuizID != cmd.QuizID || existing.HostID != cmd.HostID {
            return domain.SessionRuntime{}, false, ErrSessionRuntimeConflict
        }
        return domain.SessionRuntime{
            SessionID:     existing.SessionID,
            QuizID:        existing.QuizID,
            HostID:        existing.HostID,
            RoomCode:      existing.RoomCode,
            Status:        existing.Status,
            InitializedAt: existing.InitializedAt,
        }, false, nil
    }

    if !errors.Is(err, redis.ErrSessionNotFound) {
        return domain.SessionRuntime{}, false, ErrRuntimeStoreUnavailable
    }

    bootstrap, err := s.managementRepository.GetSessionBootstrap(ctx, cmd.SessionID)
    if err != nil {
        return domain.SessionRuntime{}, false, s.mapManagementError(err)
    }

    var reservedCode string
    for i := 0; i < 50; i++ {
        code := s.roomCodeGenerator.Generate()
        ok, err := s.roomCodeRepository.Reserve(ctx, code, cmd.SessionID)
        if err != nil {
            return domain.SessionRuntime{}, false, ErrRuntimeStoreUnavailable
        }
        if ok {
            reservedCode = code
            break
        }
    }
    if reservedCode == "" {
        return domain.SessionRuntime{}, false, ErrRoomCodeUnavailable
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
        return domain.SessionRuntime{}, false, s.mapRedisError(err)
    }

    return runtime, true, nil
}

func (s *Service) GetSessionRuntime(ctx context.Context, cmd GetSessionRuntimeParams) (domain.SessionRuntime, error) {
    if cmd.SessionID == "" {
        return domain.SessionRuntime{}, ErrSessionRuntimeNotFound
    }

    res, err := s.runtimeRepository.Get(ctx, cmd.SessionID)
    if err != nil {
        if errors.Is(err, redis.ErrSessionNotFound) {
            return domain.SessionRuntime{}, ErrSessionRuntimeNotFound
        }
        return domain.SessionRuntime{}, ErrRuntimeStoreUnavailable
    }

    // ПУНКТ 3: Возвращаем результат маппингом напрямую (можно через ToRuntime хелпер)
    return domain.SessionRuntime{
        SessionID:     res.SessionID,
        QuizID:        res.QuizID,
        HostID:        res.HostID,
        RoomCode:      res.RoomCode,
        Status:        res.Status,
        InitializedAt: res.InitializedAt,
    }, nil
}

func (s *Service) DeleteSessionRuntime(ctx context.Context, cmd DeleteSessionRuntimeParams) error {
    if cmd.SessionID == "" {
        return nil
    }

    err := s.runtimeRepository.Delete(ctx, cmd.SessionID)
    if err != nil {
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
