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
	filter := bson.M{"team_id": note.TeamID, "title": note.Title}
	update := bson.M{
		"$set": bson.M{
			"title":      note.Title,
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

func (r *NoteRepository) FindNotesByTeamID(teamID string) ([]models.Note, error) {
	var notes []models.Note
	cursor, err := r.collection.Find(context.Background(), bson.M{"team_id": teamID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepository) FindNoteByTeamIDAndTitle(teamID, title string) (models.Note, error) {
	var note models.Note
	err := r.collection.FindOne(context.Background(), bson.M{"team_id": teamID, "title": title}).Decode(&note)
	return note, err
}

func (r *NoteRepository) UpdateNoteTitle(teamID, oldTitle, newTitle string) error {
	filter := bson.M{"team_id": teamID, "title": oldTitle}
	update := bson.M{"$set": bson.M{"title": newTitle}}
	_, err := r.collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	return err
}
