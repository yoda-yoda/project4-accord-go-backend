// go-server/repository/participant_repository.go
package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"go-server/models"

	"github.com/go-redis/redis/v8"
)

// ParticipantRepository manages participants using Redis
type ParticipantRepository struct {
	redisClient *redis.Client
}

// NewParticipantRepository creates a new ParticipantRepository
func NewParticipantRepository(redisClient *redis.Client) *ParticipantRepository {
	return &ParticipantRepository{redisClient: redisClient}
}

// validKind checks if the provided kind is valid
func validKind(kind string) bool {
	switch kind {
	case models.KindCanvas, models.KindNote, models.KindAudio:
		return true
	default:
		return false
	}
}

// generateKey generates the Redis key for storing participants
func generateKey(teamID, kind string) (string, error) {
	if !validKind(kind) {
		return "", fmt.Errorf("invalid kind: %s", kind)
	}
	return fmt.Sprintf("team:%s:%s:participants", teamID, kind), nil
}

// AddParticipant adds a participant to a specific team and kind
func (pr *ParticipantRepository) AddParticipant(ctx context.Context, teamID, kind string, participant models.Participant) error {
	key, err := generateKey(teamID, kind)
	if err != nil {
		return err
	}

	participantData, err := json.Marshal(participant)
	if err != nil {
		return fmt.Errorf("failed to marshal participant: %w", err)
	}

	if err := pr.redisClient.HSet(ctx, key, participant.ID, participantData).Err(); err != nil {
		return fmt.Errorf("failed to HSet participant in Redis: %w", err)
	}

	log.Printf("Added participant %s to team %s, kind %s\n", participant.ID, teamID, kind)
	return nil
}

// RemoveParticipant removes a participant from a specific team and kind
func (pr *ParticipantRepository) RemoveParticipant(ctx context.Context, teamID, kind, participantID string) error {
	key, err := generateKey(teamID, kind)
	if err != nil {
		return err
	}

	if err := pr.redisClient.HDel(ctx, key, participantID).Err(); err != nil {
		return fmt.Errorf("failed to HDel participant from Redis: %w", err)
	}

	log.Printf("Removed participant %s from team %s, kind %s\n", participantID, teamID, kind)
	return nil
}

// GetParticipants retrieves all participants from a specific team and kind
func (pr *ParticipantRepository) GetParticipants(ctx context.Context, teamID, kind string) ([]models.Participant, error) {
	key, err := generateKey(teamID, kind)
	if err != nil {
		return nil, err
	}

	entries, err := pr.redisClient.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to HGetAll participants from Redis: %w", err)
	}

	participants := make([]models.Participant, 0, len(entries))
	for _, v := range entries {
		var p models.Participant
		if err := json.Unmarshal([]byte(v), &p); err != nil {
			log.Printf("Unmarshal error for participant data: %v\n", err)
			continue
		}
		participants = append(participants, p)
	}

	log.Printf("Retrieved %d participants from team %s, kind %s\n", len(participants), teamID, kind)
	return participants, nil
}
