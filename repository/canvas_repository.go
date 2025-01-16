package repository

import (
	"context"
	"go-server/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CanvasRepository struct {
	collection *mongo.Collection
}

func NewCanvasRepository(collection *mongo.Collection) *CanvasRepository {
	return &CanvasRepository{collection: collection}
}

func (r *CanvasRepository) SaveCanvas(canvas models.Canvas) error {
	canvas.CreatedAt = time.Now()
	_, err := r.collection.InsertOne(context.Background(), canvas)
	return err
}

func (r *CanvasRepository) FindCanvasByTeamID(teamID string) (models.Canvas, error) {
	var canvas models.Canvas
	err := r.collection.FindOne(context.Background(), bson.M{"team_id": teamID}).Decode(&canvas)
	return canvas, err
}
