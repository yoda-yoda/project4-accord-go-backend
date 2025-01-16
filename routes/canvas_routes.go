package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
)

func CanvasRoutes(app *fiber.App, canvasController *controllers.CanvasController) {
	app.Post("/canvas", canvasController.CreateCanvas)
	app.Get("/canvas/:teamId", canvasController.GetCanvasByTeamID)
}
