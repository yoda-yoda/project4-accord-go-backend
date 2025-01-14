package repository

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type Participant struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RedisParticipantRepository struct {
	client *redis.Client
}

func NewRedisParticipantRepository(client *redis.Client) *RedisParticipantRepository {
	return &RedisParticipantRepository{client: client}
}

// Redis Key 생성
func generateKey(teamID string, dataType string) string {
	return "team:" + teamID + ":" + dataType + ":participants"
}

// 참여자 추가
func (r *RedisParticipantRepository) AddParticipant(teamID string, dataType string, participant Participant) error {
	ctx := context.Background()
	data, err := json.Marshal(participant)
	if err != nil {
		return err
	}
	return r.client.RPush(ctx, generateKey(teamID, dataType), data).Err()
}

// 참여자 목록 가져오기
func (r *RedisParticipantRepository) GetParticipants(teamID string, dataType string) ([]Participant, error) {
	ctx := context.Background()
	data, err := r.client.LRange(ctx, generateKey(teamID, dataType), 0, -1).Result()
	if err != nil {
		return nil, err
	}
	participants := []Participant{}
	for _, item := range data {
		var p Participant
		if err := json.Unmarshal([]byte(item), &p); err == nil {
			participants = append(participants, p)
		}
	}
	return participants, nil
}

// 참여자 제거
func (r *RedisParticipantRepository) RemoveParticipant(teamID string, dataType string, participantID string) error {
	ctx := context.Background()
	participants, err := r.GetParticipants(teamID, dataType)
	if err != nil {
		return err
	}
	updatedList := []Participant{}
	for _, p := range participants {
		if p.ID != participantID {
			updatedList = append(updatedList, p)
		}
	}
	if err := r.client.Del(ctx, generateKey(teamID, dataType)).Err(); err != nil {
		return err
	}
	for _, p := range updatedList {
		if err := r.AddParticipant(teamID, dataType, p); err != nil {
			return err
		}
	}
	return nil
}
