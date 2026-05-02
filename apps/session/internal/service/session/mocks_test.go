package session

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/domain"
)

var (
	_ ManagementRepository = (*managementRepositoryMock)(nil)
	_ RuntimeRepository    = (*runtimeRepositoryMock)(nil)
	_ RoomCodeRepository   = (*roomCodeRepositoryMock)(nil)
	_ RoomCodeGenerator    = (*roomCodeGeneratorMock)(nil)
)

type managementRepositoryMock struct {
	GetSessionBootstrapFn  func(ctx context.Context, sessionID string) (domain.SessionBootstrap, error)
	ReportSessionStatusFn  func(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error
	ReportSessionResultsFn func(ctx context.Context, sessionID string, results domain.SessionResults) error
}

func (m *managementRepositoryMock) GetSessionBootstrap(ctx context.Context, sessionID string) (domain.SessionBootstrap, error) {
	if m.GetSessionBootstrapFn != nil {
		return m.GetSessionBootstrapFn(ctx, sessionID)
	}

	return domain.SessionBootstrap{}, nil
}

func (m *managementRepositoryMock) ReportSessionStatus(ctx context.Context, sessionID string, update domain.SessionStatusUpdate) error {
	if m.ReportSessionStatusFn != nil {
		return m.ReportSessionStatusFn(ctx, sessionID, update)
	}

	return nil
}

func (m *managementRepositoryMock) ReportSessionResults(ctx context.Context, sessionID string, results domain.SessionResults) error {
	if m.ReportSessionResultsFn != nil {
		return m.ReportSessionResultsFn(ctx, sessionID, results)
	}

	return nil
}

type runtimeRepositoryMock struct {
	CreateFn func(ctx context.Context, runtime domain.SessionRuntime, quiz domain.QuizSnapshot) error
	GetFn    func(ctx context.Context, sessionID string) (domain.SessionRuntime, error)
	DeleteFn func(ctx context.Context, sessionID string) error
}

func (m *runtimeRepositoryMock) Create(ctx context.Context, runtime domain.SessionRuntime, quiz domain.QuizSnapshot) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, runtime, quiz)
	}

	return nil
}

func (m *runtimeRepositoryMock) Get(ctx context.Context, sessionID string) (domain.SessionRuntime, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, sessionID)
	}

	return domain.SessionRuntime{}, nil
}

func (m *runtimeRepositoryMock) Delete(ctx context.Context, sessionID string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, sessionID)
	}

	return nil
}

type roomCodeRepositoryMock struct {
	ReserveFn func(ctx context.Context, roomCode string, sessionID string) (bool, error)
	ReleaseFn func(ctx context.Context, roomCode string) error
}

func (m *roomCodeRepositoryMock) Reserve(ctx context.Context, roomCode string, sessionID string) (bool, error) {
	if m.ReserveFn != nil {
		return m.ReserveFn(ctx, roomCode, sessionID)
	}

	return false, nil
}

func (m *roomCodeRepositoryMock) Release(ctx context.Context, roomCode string) error {
	if m.ReleaseFn != nil {
		return m.ReleaseFn(ctx, roomCode)
	}

	return nil
}

type roomCodeGeneratorMock struct {
	GenerateFn func() string
}

func (m *roomCodeGeneratorMock) Generate() string {
	if m.GenerateFn != nil {
		return m.GenerateFn()
	}

	return "00000000"
}
