package repository

import (
	"context"
	"go-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type NoteRepository struct {
	collection *mongo.Collection
}

func NewNoteRepository(collection *mongo.Collection) *NoteRepository {
	return &NoteRepository{collection: collection}
}

func (r *NoteRepository) SaveNote(note models.Note) error {
	note.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(context.Background(), note)
	return err
}

func (r *NoteRepository) FindNoteByTeamID(teamID string) (models.Note, error) {
	var note models.Note
	err := r.collection.FindOne(context.Background(), bson.M{"team_id": teamID}).Decode(&note)
	return note, err
}
