package main

import (
	"go-server/configs"
	"go-server/controllers"
	"go-server/repository"
	"go-server/routes"
	"go-server/server"
	"go-server/utils"
	"log"
	"os"

	fiberprometheus "github.com/ansrivas/fiberprometheus/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {

	os.Setenv("CONSUL_ADDRESS", "http://localhost:8500")
	err := configs.RegisterService(
		"go-server",
		"go-server",
		"localhost",
		4000,
		"http://localhost:4000/health",
	)
	if err != nil {
		log.Fatalf("Consul service registration failed: %v", err)
	}

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

	store := utils.NewPublicKeyStore()

	app := fiber.New()

	p := fiberprometheus.New("go-server")

	p.RegisterAt(app, "/metrics")

	app.Use(p.Middleware)

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.NoteRoutes(app, noteController, store)
	routes.WebSocketRoutes(app, participantsController, audioController)
	routes.CanvasRoutes(app, canvasController, store)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "UP",
		})
	})

	go func() {
		if err := server.RunGRPCServer(store); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
		log.Println("gRPC server started")
	}()

	log.Println("Starting server on port 4000...")
	if err := app.Listen(":4000"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
