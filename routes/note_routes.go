package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
)

func NoteRoutes(app *fiber.App, controller *controllers.NoteController) {
	app.Post("/note", controller.CreateNote)
	app.Get("/notes/:teamId", controller.GetNotesByTeamID)
	app.Get("/notes/:teamId/:title", controller.GetNoteByTeamIDAndTitle)
	app.Put("/notes/:teamId/:oldTitle/:newTitle", controller.UpdateNoteTitle)
}
