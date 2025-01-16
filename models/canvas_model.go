package models

import "time"

type Canvas struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	TeamID    string    `bson:"team_id" json:"team_id"`
	Canvas    string    `bson:"canvas" json:"canvas"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
