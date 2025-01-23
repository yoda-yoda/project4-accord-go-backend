package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go-server/models"

	"github.com/go-redis/redis/v8"
)

type ParticipantRepository struct {
	redisClient *redis.Client
}

func NewParticipantRepository(redisClient *redis.Client) *ParticipantRepository {
	return &ParticipantRepository{redisClient: redisClient}
}

func (pr *ParticipantRepository) AddParticipant(ctx context.Context, teamID, kind string, participant models.Participant) error {
	key := fmt.Sprintf("team:%s:%s:participants", teamID, kind)
	participantData, err := json.Marshal(participant)
	if err != nil {
		return err
	}
	return pr.redisClient.HSet(ctx, key, participant.ID, participantData).Err()
}

func (pr *ParticipantRepository) RemoveParticipant(ctx context.Context, teamID, kind, participantID string) error {
	key := fmt.Sprintf("team:%s:%s:participants", teamID, kind)
	return pr.redisClient.HDel(ctx, key, participantID).Err()
}

func (pr *ParticipantRepository) GetParticipants(ctx context.Context, teamID, kind string) ([]models.Participant, error) {
	key := fmt.Sprintf("team:%s:%s:participants", teamID, kind)
	entries, err := pr.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	participants := make([]models.Participant, 0, len(entries))
	for _, v := range entries {
		var p models.Participant
		if err := json.Unmarshal([]byte(v), &p); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}
		participants = append(participants, p)
	}
	return participants, nil
}
