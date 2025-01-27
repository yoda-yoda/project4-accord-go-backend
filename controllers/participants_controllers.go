package controllers

import (
	"encoding/json"
	"log"
	"sync"

	"go-server/configs"
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/websocket/v2"
)

type ParticipantsController struct {
	participantRepo *repository.ParticipantRepository
	connections     map[*websocket.Conn]string
	rooms           map[string]map[*websocket.Conn]bool
	mu              sync.Mutex
}

func NewParticipantsController(participantRepo *repository.ParticipantRepository) *ParticipantsController {
	return &ParticipantsController{
		participantRepo: participantRepo,
		connections:     make(map[*websocket.Conn]string),
		rooms:           make(map[string]map[*websocket.Conn]bool),
	}
}

func (pc *ParticipantsController) HandleWebSocket(c *websocket.Conn) {
	log.Println("New participant connected")
	log.Println("Remote addr:", c.RemoteAddr())
	log.Println("Local addr:", c.LocalAddr())

	defer func() {
		pc.mu.Lock()
		participant := pc.connections[c]
		delete(pc.connections, c)
		pc.mu.Unlock()

		// Remove participant from Redis
		var payload map[string]string
		if err := json.Unmarshal([]byte(participant), &payload); err == nil {
			log.Println("Payload:", payload)
			pc.participantRepo.RemoveParticipant(configs.Ctx, payload["team_id"], payload["kind"], payload["participant"])
			pc.broadcastParticipants(payload["team_id"], payload["kind"])
			// 추가: 오디오 참여자 업데이트
			if payload["kind"] == "audio" {
				pc.broadcastAudioParticipants(payload["team_id"], payload["kind"])
			}
		} else {
			log.Println("Failed to unmarshal participant:", err)
		}

		c.Close()
	}()

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var payload map[string]string
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Println("Unmarshal error:", err)
			continue
		}

		switch payload["action"] {
		case "addParticipant":
			pc.handleAddParticipant(c, payload)
		case "removeParticipant":
			pc.handleRemoveParticipant(c, payload)
		case "getParticipants":
			pc.handleGetParticipants(c, payload)
		case "addAudioParticipant":
			pc.handleAddAudioParticipant(c, payload)
		case "removeAudioParticipant":
			pc.handleRemoveAudioParticipant(c, payload)
		case "getAudioParticipants":
			pc.handleGetAudioParticipants(c, payload)
		}
	}
}

func (pc *ParticipantsController) handleAddAudioParticipant(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]
	participant := models.Participant{
		ID:             payload["participant"],
		Name:           payload["name"],
		ProfilePicture: payload["profilePicture"],
		Color:          payload["color"],
	}

	err := pc.participantRepo.AddParticipant(configs.Ctx, teamID, kind, participant)
	if err != nil {
		log.Println("Failed to add audio participant:", err)
		return
	}

	payloadStr, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal payload:", err)
		return
	}

	pc.mu.Lock()
	pc.connections[c] = string(payloadStr)
	roomKey := teamID + ":" + kind
	if pc.rooms[roomKey] == nil {
		pc.rooms[roomKey] = make(map[*websocket.Conn]bool)
	}
	pc.rooms[roomKey][c] = true
	pc.mu.Unlock()

	pc.broadcastAudioParticipants(teamID, kind)
}

func (pc *ParticipantsController) handleRemoveAudioParticipant(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]
	participantID := payload["participant"]

	err := pc.participantRepo.RemoveParticipant(configs.Ctx, teamID, kind, participantID)
	if err != nil {
		log.Println("Failed to remove audio participant:", err)
		return
	}

	pc.broadcastAudioParticipants(teamID, kind)
}

func (pc *ParticipantsController) handleGetAudioParticipants(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]

	participants, err := pc.participantRepo.GetParticipants(configs.Ctx, teamID, kind)
	if err != nil {
		log.Println("Failed to get audio participants:", err)
		return
	}

	response := map[string]interface{}{
		"action":            "getAudioParticipants",
		"audioParticipants": participants,
	}
	responseMsg, err := json.Marshal(response)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	if err := c.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
		log.Println("Write error:", err)
	}
}

func (pc *ParticipantsController) broadcastAudioParticipants(teamID, kind string) {
	participants, err := pc.participantRepo.GetParticipants(configs.Ctx, teamID, kind)
	if err != nil {
		log.Println("Failed to get audio participants:", err)
		return
	}

	response := map[string]interface{}{
		"action":            "updateAudioParticipants",
		"audioParticipants": participants,
	}
	responseMsg, err := json.Marshal(response)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	roomKey := teamID + ":" + kind
	for conn := range pc.rooms[roomKey] {
		if err := conn.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
			log.Println("Write error:", err)
		}
	}
}

func (pc *ParticipantsController) handleAddParticipant(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]
	participant := models.Participant{
		ID:             payload["participant"],
		Name:           payload["name"],
		ProfilePicture: payload["profilePicture"],
		Color:          payload["color"],
	}

	err := pc.participantRepo.AddParticipant(configs.Ctx, teamID, kind, participant)
	if err != nil {
		log.Println("Failed to add participant:", err)
		return
	}

	payloadStr, err := json.Marshal(payload)
	if err != nil {
		log.Println("Failed to marshal payload:", err)
		return
	}

	pc.mu.Lock()
	pc.connections[c] = string(payloadStr)
	roomKey := teamID + ":" + kind
	if pc.rooms[roomKey] == nil {
		pc.rooms[roomKey] = make(map[*websocket.Conn]bool)
	}
	pc.rooms[roomKey][c] = true
	pc.mu.Unlock()

	pc.broadcastParticipants(teamID, kind)
}

func (pc *ParticipantsController) handleRemoveParticipant(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]
	participantID := payload["participant"]

	err := pc.participantRepo.RemoveParticipant(configs.Ctx, teamID, kind, participantID)
	if err != nil {
		log.Println("Failed to remove participant:", err)
		return
	}

	pc.broadcastParticipants(teamID, kind)
}

func (pc *ParticipantsController) handleGetParticipants(c *websocket.Conn, payload map[string]string) {
	teamID := payload["team_id"]
	kind := payload["kind"]

	participants, err := pc.participantRepo.GetParticipants(configs.Ctx, teamID, kind)
	if err != nil {
		log.Println("Failed to get participants:", err)
		return
	}

	response := map[string]interface{}{
		"action":       "getParticipants",
		"participants": participants,
	}
	responseMsg, err := json.Marshal(response)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	if err := c.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
		log.Println("Write error:", err)
	}
}

func (pc *ParticipantsController) broadcastParticipants(teamID, kind string) {
	participants, err := pc.participantRepo.GetParticipants(configs.Ctx, teamID, kind)
	if err != nil {
		log.Println("Failed to get participants:", err)
		return
	}

	response := map[string]interface{}{
		"action":       "updateParticipants",
		"participants": participants,
	}
	responseMsg, err := json.Marshal(response)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}

	pc.mu.Lock()
	defer pc.mu.Unlock()

	roomKey := teamID + ":" + kind
	for conn := range pc.rooms[roomKey] {
		if err := conn.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
			log.Println("Write error:", err)
		}
	}
}
