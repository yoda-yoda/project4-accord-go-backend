package repository

import (
	"context"
	"go-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type CanvasRepository struct {
	collection *mongo.Collection
}

func NewCanvasRepository(collection *mongo.Collection) *CanvasRepository {
	return &CanvasRepository{collection: collection}
}

func (r *CanvasRepository) SaveCanvas(canvas models.Canvas) error {
	canvas.CreatedAt = time.Now()
	filter := bson.M{"team_id": canvas.TeamID, "title": canvas.Title}
	update := bson.M{
		"$set": canvas,
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (r *CanvasRepository) FindCanvasByID(id string) (models.Canvas, error) {
	var canvas models.Canvas
	err := r.collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&canvas)
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

func (r *CanvasRepository) UpdateCanvasTitle(teamID, oldTitle, newTitle string) error {
	filter := bson.M{"team_id": teamID, "title": oldTitle}
	update := bson.M{"$set": bson.M{"title": newTitle}}
	_, err := r.collection.UpdateOne(context.Background(), filter, update, options.Update().SetUpsert(false))
	return err
}
