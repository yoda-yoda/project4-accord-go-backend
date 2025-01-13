package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
)

func NoteRoutes(app *fiber.App, controller *controllers.NoteController) {
	app.Post("/note", controller.CreateNote)
	app.Get("/note/:teamId", controller.GetNoteByTeamID)
}
