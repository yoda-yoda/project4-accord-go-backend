package repository

import (
	"context"
	"go-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type NoteRepository struct {
	collection *mongo.Collection
}

func NewNoteRepository(collection *mongo.Collection) *NoteRepository {
	return &NoteRepository{collection: collection}
}

func (r *NoteRepository) SaveNote(note models.Note) (string, error) {
	note.CreatedAt = time.Now()

	var filter bson.M
	var objectID primitive.ObjectID
	var err error

	if note.ID != "" {
		objectID, err = primitive.ObjectIDFromHex(note.ID)
		if err != nil {
			return "", err
		}
		filter = bson.M{"_id": objectID}
	} else {
		objectID = primitive.NewObjectID()
		note.ID = objectID.Hex()
		filter = bson.M{"_id": objectID}
	}

	update := bson.M{
		"$set": bson.M{
			"team_id":    note.TeamID,
			"title":      note.Title,
			"note":       note.Note,
			"created_at": note.CreatedAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return "", err
	}
	return objectID.Hex(), nil
}

func (r *NoteRepository) FindNoteByID(id string) (models.Note, error) {
	var note models.Note
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return note, err
	}
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&note)
	return note, err
}

func (r *NoteRepository) FindNotesByTeamID(teamID string) ([]models.Note, error) {
	var notes []models.Note
	projection := bson.M{"note": 0} // note 필드를 제외
	opts := options.Find().SetProjection(projection)
	cursor, err := r.collection.Find(context.Background(), bson.M{"team_id": teamID}, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &notes); err != nil {
		return nil, err
	}
	return notes, nil
}

func (r *NoteRepository) UpdateNoteTitle(id, newTitle string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"title": newTitle}}
	_, err = r.collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	return err
}

func (r *NoteRepository) DeleteNoteByID(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.DeleteOne(context.Background(), bson.M{"_id": objectID})
	return err
}
