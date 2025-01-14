package controllers

import (
	"encoding/json"
	"log"

	"go-server/repository"
	service "go-server/services"

	"github.com/gofiber/websocket/v2"
)

type WebSocketController struct {
	participantService *service.ParticipantService
}

func NewWebSocketController(service *service.ParticipantService) *WebSocketController {
	return &WebSocketController{participantService: service}
}

func (wsc *WebSocketController) HandleWebSocket(c *websocket.Conn) {
	defer c.Close()
	log.Println("WebSocket connected")

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		// 메시지 파싱
		var data struct {
			Type        string                 `json:"type"`
			TeamID      string                 `json:"team_id"`
			DataType    string                 `json:"data_type"` // note, canvas, voice
			Participant repository.Participant `json:"participant"`
		}
		if err := json.Unmarshal(msg, &data); err != nil {
			log.Println("Invalid message format:", err)
			continue
		}

		// 메시지 처리
		switch data.Type {
		case "addParticipant":
			err := wsc.participantService.AddParticipant(data.TeamID, data.DataType, data.Participant)
			if err != nil {
				log.Println("Error adding participant:", err)
			}
		case "removeParticipant":
			err := wsc.participantService.RemoveParticipant(data.TeamID, data.DataType, data.Participant.ID)
			if err != nil {
				log.Println("Error removing participant:", err)
			}
		default:
			log.Println("Unknown message type:", data.Type)
		}
	}
}
