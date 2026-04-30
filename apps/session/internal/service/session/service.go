package session

import (
	"context"

	"github.com/We-seek-D-p/quiz-platform/apps/session/internal/models"
)

type Service struct {
	managementClient   ManagementClient
	runtimeRepository  RuntimeRepository
	roomCodeRepository RoomCodeRepository
	roomCodeGenerator  RoomCodeGenerator
}

func NewService(
	managementClient ManagementClient,
	runtimeRepository RuntimeRepository,
	roomCodeRepository RoomCodeRepository,
	roomCodeGenerator RoomCodeGenerator,
) *Service {
	return &Service{
		managementClient:   managementClient,
		runtimeRepository:  runtimeRepository,
		roomCodeRepository: roomCodeRepository,
		roomCodeGenerator:  roomCodeGenerator,
	}
}

func (s *Service) InitSession(ctx context.Context, cmd InitSessionParams) (models.SessionRuntime, error) {
	return models.SessionRuntime{}, ErrNotImplemented
}

func (s *Service) GetSessionRuntime(ctx context.Context, cmd GetSessionRuntimeParams) (models.SessionRuntime, error) {
	return models.SessionRuntime{}, ErrNotImplemented
}

func (s *Service) DeleteSessionRuntime(ctx context.Context, cmd DeleteSessionRuntimeParams) error {
	return ErrNotImplemented
}
