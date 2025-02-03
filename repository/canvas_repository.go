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

type CanvasRepositoryInterface interface {
	SaveCanvas(canvas models.Canvas) (string, error)
	FindCanvasByID(id string) (models.Canvas, error)
	FindCanvasesByTeamID(teamID string) ([]models.Canvas, error)
	UpdateCanvasTitle(id, newTitle string) error
	DeleteCanvasByID(id string) error
}

type CanvasRepository struct {
	collection *mongo.Collection
}

func NewCanvasRepository(collection *mongo.Collection) *CanvasRepository {
	return &CanvasRepository{collection: collection}
}

func (r *CanvasRepository) SaveCanvas(canvas models.Canvas) (string, error) {
	canvas.CreatedAt = time.Now()

	var filter bson.M
	var objectID primitive.ObjectID
	var err error

	if canvas.ID != "" {
		objectID, err = primitive.ObjectIDFromHex(canvas.ID)
		if err != nil {
			return "", err
		}
		filter = bson.M{"_id": objectID}
	} else {
		objectID = primitive.NewObjectID()
		canvas.ID = objectID.Hex()
		filter = bson.M{"_id": objectID}
	}

	update := bson.M{
		"$set": bson.M{
			"team_id":    canvas.TeamID,
			"title":      canvas.Title,
			"canvas":     canvas.Canvas,
			"created_at": canvas.CreatedAt,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err = r.collection.UpdateOne(context.Background(), filter, update, opts)
	if err != nil {
		return "", err
	}
	return objectID.Hex(), nil
}

func (r *CanvasRepository) FindCanvasByID(id string) (models.Canvas, error) {
	var canvas models.Canvas
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return canvas, err
	}
	err = r.collection.FindOne(context.Background(), bson.M{"_id": objectID}).Decode(&canvas)
	return canvas, err
}

func (r *CanvasRepository) FindCanvasesByTeamID(teamID string) ([]models.Canvas, error) {
	var canvases []models.Canvas
	projection := bson.M{"canvas": 0}
	opts := options.Find().SetProjection(projection)
	cursor, err := r.collection.Find(context.Background(), bson.M{"team_id": teamID}, opts)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(context.Background(), &canvases); err != nil {
		return nil, err
	}
	return canvases, nil
}

func (r *CanvasRepository) UpdateCanvasTitle(id, newTitle string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectID}
	update := bson.M{"$set": bson.M{"title": newTitle}}
	_, err = r.collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	return err
}

func (r *CanvasRepository) DeleteCanvasByID(id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": objectID}
	_, err = r.collection.DeleteOne(context.Background(), filter)
	return err
}
