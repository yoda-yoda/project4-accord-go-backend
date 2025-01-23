package routes

import (
	"go-server/controllers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

func WebSocketRoutes(app *fiber.App, participanstController *controllers.ParticipantsController, audioController *controllers.AudioSocketController) {
	app.Get("/ws", websocket.New(participanstController.HandleWebSocket))
	app.Get("/webrtc/audio", websocket.New(audioController.HandleWebRTC))
}
