package routes

import (
	"go-server/controllers"
	middleware "go-server/middlewares"
	"go-server/utils"

	"github.com/gofiber/fiber/v2"
)

func CanvasRoutes(app *fiber.App, canvasController *controllers.CanvasController, store *utils.PublicKeyStore) {

	canvasGroup := app.Group("/canvas", middleware.JWTParser(store))

	canvasGroup.Post("/", canvasController.CreateCanvas)
	canvasGroup.Get("/:id", canvasController.GetCanvasByID)
	canvasGroup.Get("/team/:teamId", canvasController.GetCanvasesByTeamID)
	canvasGroup.Put("/:id/title", canvasController.UpdateCanvasTitle)
	canvasGroup.Delete("/:id", canvasController.DeleteCanvasByID)
}
