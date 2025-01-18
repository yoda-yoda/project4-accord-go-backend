package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WebSocketRoutes(app *fiber.App, wsController *controllers.WebSocketController, audioController *controllers.AudioSocketController) {
	app.Get("/ws", websocket.New(wsController.HandleWebSocket))
	app.Get("/webrtc", websocket.New(wsController.HandleWebRTC))
	app.Get("/webrtc/audio", websocket.New(audioController.HandleWebRTC))
}
