package main

import (
	"go-server/configs"
	"go-server/controllers"
	"go-server/repository"
	"go-server/routes"
	service "go-server/services"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	redisClient := configs.ConnectRedis()
	participantRepo := repository.NewRedisParticipantRepository(redisClient)
	participantService := service.NewParticipantService(participantRepo)
	wsService := service.NewWebSocketService()

	client := configs.ConnectMongo()
	collection := client.Database("mydb").Collection("notes")

	noteRepo := repository.NewNoteRepository(collection)
	noteController := controllers.NewNoteController(noteRepo)

	wsController := controllers.NewWebSocketController(participantService, wsService)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.NoteRoutes(app, noteController)
	routes.WebSocketRoutes(app, wsController)

	log.Println("Starting server on port 4000...")
	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
