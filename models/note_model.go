package models

import "time"

type Note struct {
	ID        string    `bson:"_id,omitempty" json:"id,omitempty"`
	TeamID    string    `bson:"team_id" json:"team_id"`
	Version   int       `bson:"version"`
	Note      string    `bson:"note" json:"note"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
