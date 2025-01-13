package configs

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectMongo() *mongo.Client {
	clientOptions := options.Client().ApplyURI("mongodb://admin:adminpassword@localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("MongoDB connection error:", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB ping error:", err)
	}

	log.Println("Connected to MongoDB!")
	return client
}
