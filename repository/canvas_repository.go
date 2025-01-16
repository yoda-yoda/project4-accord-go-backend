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
	filter := bson.M{"team_id": canvas.TeamID}
	update := bson.M{
		"$set": canvas,
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.collection.UpdateOne(context.Background(), filter, update, opts)
	return err
}

func (r *CanvasRepository) FindCanvasByTeamID(teamID string) (models.Canvas, error) {
	var canvas models.Canvas
	err := r.collection.FindOne(context.Background(), bson.M{"team_id": teamID}).Decode(&canvas)
	return canvas, err
}
