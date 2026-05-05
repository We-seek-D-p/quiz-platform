package session

import "time"

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
