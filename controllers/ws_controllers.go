package controllers

import (
	"encoding/json"
	"log"

	"go-server/repository"
	service "go-server/services"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/webrtc/v4"
)

type WebSocketController struct {
	connections        map[*websocket.Conn]*webrtc.PeerConnection
	dataChannels       map[*websocket.Conn]*webrtc.DataChannel
	iceCandidateQueue  map[*websocket.Conn][]webrtc.ICECandidateInit
	participantService *service.ParticipantService
	wsService          *service.WebSocketService
}

func NewWebSocketController(
	participantService *service.ParticipantService,
	wsService *service.WebSocketService, // 추가 의존성
) *WebSocketController {
	return &WebSocketController{
		participantService: participantService,
		wsService:          wsService, // 추가된 필드 초기화
		connections:        make(map[*websocket.Conn]*webrtc.PeerConnection),
		dataChannels:       make(map[*websocket.Conn]*webrtc.DataChannel),
	}
}

func (wsc *WebSocketController) HandleWebSocket(c *websocket.Conn) {
	var peerConnection *webrtc.PeerConnection
	defer func() {
		if peerConnection != nil {
			peerConnection.Close()
		}
		c.Close()
	}()
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

func (wsc *WebSocketController) HandleWebRTC(c *websocket.Conn) {
	if wsc.connections == nil {
		wsc.connections = make(map[*websocket.Conn]*webrtc.PeerConnection)
		wsc.iceCandidateQueue = make(map[*websocket.Conn][]webrtc.ICECandidateInit)
	}

	defer func() {
		if pc, ok := wsc.connections[c]; ok {
			pc.Close()
			delete(wsc.connections, c)
		}
		delete(wsc.iceCandidateQueue, c)
		c.Close()
	}()

	log.Println("WebRTC WebSocket connected")

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		switch payload["type"] {
		case "offer":
			handleOffer(wsc, c, payload)

		case "iceCandidate":
			handleIceCandidate(wsc, c, payload)
		}
	}
}

func handleOffer(wsc *WebSocketController, c *websocket.Conn, payload map[string]interface{}) {
	sdp, _ := payload["sdp"].(string)

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  sdp,
	}

	peerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{
				URLs:       []string{"turn:127.0.0.1:3478"},
				Username:   "user",
				Credential: "pass",
			},
		},
	})
	if err != nil {
		log.Println("Failed to create PeerConnection:", err)
		return
	}

	wsc.connections[c] = peerConnection

	// Handle DataChannel creation
	peerConnection.OnDataChannel(func(d *webrtc.DataChannel) {
		log.Printf("New DataChannel %s\n", d.Label())

		// 저장
		wsc.dataChannels[c] = d

		// Handle DataChannel messages
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Received message: %s\n", string(msg.Data))

			// Broadcast message to other clients
			for conn, channel := range wsc.dataChannels {
				if conn != c && channel != nil {
					err := channel.SendText(string(msg.Data))
					if err != nil {
						log.Printf("Failed to send message to client: %v", err)
					}
				}
			}
		})
	})

	err = peerConnection.SetRemoteDescription(offer)
	if err != nil {
		log.Println("Failed to set remote description:", err)
		return
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create Answer:", err)
		return
	}

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	response := map[string]interface{}{
		"type": "answer",
		"sdp":  answer.SDP,
	}
	responseMsg, err := json.Marshal(response)
	if err != nil {
		log.Println("Marshal error:", err)
		return
	}
	if err := c.WriteMessage(websocket.TextMessage, responseMsg); err != nil {
		log.Println("Write error:", err)
		return
	}
}

func handleIceCandidate(wsc *WebSocketController, c *websocket.Conn, payload map[string]interface{}) {
	candidateMap, _ := payload["candidate"].(map[string]interface{})
	candidate := webrtc.ICECandidateInit{
		Candidate: candidateMap["candidate"].(string),
		SDPMid:    func(s string) *string { return &s }(candidateMap["sdpMid"].(string)),
	}

	pc, exists := wsc.connections[c]
	if exists {
		pc.AddICECandidate(candidate)
	}
}

func (wsc *WebSocketController) HandleWebSocketForYjs(c *websocket.Conn) {
	defer func() {
		wsc.wsService.RemoveClient(c)
		c.Close()
	}()

	log.Println("WebSocket connected")

	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			log.Println("Error parsing message:", err)
			continue
		}

		switch payload["type"] {
		case "subscribe":
			room, ok := payload["room"].(string)
			if !ok || room == "" {
				log.Println("Invalid room in subscribe message")
				continue
			}
			wsc.wsService.Subscribe(room, c)

		case "publish":
			room, ok := payload["room"].(string)
			if !ok || room == "" {
				log.Println("Invalid room in publish message")
				continue
			}
			message, ok := payload["message"].(string)
			if !ok {
				log.Println("Invalid message in publish message")
				continue
			}
			wsc.wsService.Publish(room, []byte(message))
		case "ping":
			message, ok := payload["message"].(string)
			if !ok {
				log.Println("Invalid message in publish message")
				continue
			}
			wsc.wsService.Publish("pong", []byte(message))

		default:
			log.Println("Unknown message type:", payload["type"])
		}
	}
}
