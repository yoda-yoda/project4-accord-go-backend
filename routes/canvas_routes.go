package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
)

func CanvasRoutes(app *fiber.App, canvasController *controllers.CanvasController) {
	app.Post("/canvas", canvasController.CreateCanvas)
	app.Get("/canvas/:id", canvasController.GetCanvasByID)
	app.Get("/canvases/:teamId", canvasController.GetCanvasesByTeamID)
	app.Put("/canvas/:id/title", canvasController.UpdateCanvasTitle)
	app.Delete("/canvas/:id", canvasController.DeleteCanvasByID)
}
