package routes

import (
	"go-server/controllers"
	middleware "go-server/middlewares"
	"go-server/utils"

	"github.com/gofiber/fiber/v2"
)

func NoteRoutes(app *fiber.App, noteController *controllers.NoteController, store *utils.PublicKeyStore) {

	noteGroup := app.Group("/note", middleware.JWTParser(store))
	noteGroup.Post("/", noteController.CreateNote)
	noteGroup.Get("/:id", noteController.GetNoteByID)
	noteGroup.Get("/team/:teamId", noteController.GetNotesByTeamID)
	noteGroup.Put("/:id/title", noteController.UpdateNoteTitle)
	noteGroup.Delete("/:id", noteController.DeleteNoteByID)
}
