package main

import (
	"go-server/configs"
	"go-server/controllers"
	"go-server/repository"
	"go-server/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	client := configs.ConnectMongo()
	collection := client.Database("mydb").Collection("notes")

	noteRepo := repository.NewNoteRepository(collection)
	noteController := controllers.NewNoteController(noteRepo)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:3000",
		AllowMethods: "GET,POST,HEAD,PUT,DELETE,PATCH,OPTIONS",
	}))

	routes.NoteRoutes(app, noteController)

	app.Listen(":4000")
}
