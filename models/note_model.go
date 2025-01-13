package models

import "time"

type Note struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	TeamID    string    `bson:"team_id" json:"team_id"`
	Note      string    `bson:"note" json:"note"` // 대용량 JSON 데이터
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
