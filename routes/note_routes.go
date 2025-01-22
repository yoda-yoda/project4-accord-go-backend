package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
)

func NoteRoutes(app *fiber.App, noteController *controllers.NoteController) {
	app.Post("/note", noteController.CreateNote)
	app.Get("/note/:id", noteController.GetNoteByID)
	app.Get("/notes/:teamId", noteController.GetNotesByTeamID)
	app.Put("/note/:id/title", noteController.UpdateNoteTitle)
	app.Delete("/note/:id", noteController.DeleteNoteByID)
}
