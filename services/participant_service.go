package service

import "go-server/repository"

type ParticipantService struct {
	repo *repository.RedisParticipantRepository
}

func NewParticipantService(repo *repository.RedisParticipantRepository) *ParticipantService {
	return &ParticipantService{repo: repo}
}

func (ps *ParticipantService) AddParticipant(teamID, dataType string, participant repository.Participant) error {
	return ps.repo.AddParticipant(teamID, dataType, participant)
}

func (ps *ParticipantService) GetParticipants(teamID, dataType string) ([]repository.Participant, error) {
	return ps.repo.GetParticipants(teamID, dataType)
}

func (ps *ParticipantService) RemoveParticipant(teamID, dataType, participantID string) error {
	return ps.repo.RemoveParticipant(teamID, dataType, participantID)
}
