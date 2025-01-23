package main

import (
	"go-server/configs"
	"go-server/controllers"
	"go-server/repository"
	"go-server/routes"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	configs.ConnectRedis()
	client := configs.ConnectMongo()
	redisClient := configs.GetRedisClient()

	collection := client.Database("mydb").Collection("notes")
	collectionCanvas := client.Database("mydb").Collection("canvases")

	noteRepo := repository.NewNoteRepository(collection)
	noteController := controllers.NewNoteController(noteRepo)

	participantRepo := repository.NewParticipantRepository(redisClient)
	participantsController := controllers.NewParticipantsController(participantRepo)

	audioController := controllers.NewAudioSocketController()

	canvasRepo := repository.NewCanvasRepository(collectionCanvas)
	canvasController := controllers.NewCanvasController(canvasRepo)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.NoteRoutes(app, noteController)
	routes.WebSocketRoutes(app, participantsController, audioController)
	routes.CanvasRoutes(app, canvasController)

	log.Println("Starting server on port 4000...")
	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
