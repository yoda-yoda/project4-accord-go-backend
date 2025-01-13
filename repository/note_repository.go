package repository

import (
	"context"
	"go-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NoteRepository struct {
	collection *mongo.Collection
}

func NewNoteRepository(collection *mongo.Collection) *NoteRepository {
	return &NoteRepository{collection: collection}
}

func (r *NoteRepository) SaveNote(note models.Note) error {
	filter := bson.M{"team_id": note.TeamID} // team_id 기준으로 문서 검색
	update := bson.M{
		"$set": bson.M{
			"note":       note.Note,
			"created_at": time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(
		context.Background(),
		filter,
		update,
		options.Update().SetUpsert(true),
	)
	return err
}

func (r *NoteRepository) FindNoteByTeamID(teamID string) (models.Note, error) {
	var note models.Note
	err := r.collection.FindOne(context.Background(), bson.M{"team_id": teamID}).Decode(&note)
	return note, err
}
