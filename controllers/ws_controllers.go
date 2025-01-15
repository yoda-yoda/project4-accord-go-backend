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
	iceCandidateQueue  map[*websocket.Conn][]webrtc.ICECandidateInit
	participantService *service.ParticipantService
}

func NewWebSocketController(service *service.ParticipantService) *WebSocketController {
	return &WebSocketController{participantService: service}
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
			log.Println("Processing WebRTC Offer...")
			sdp, ok := payload["sdp"].(string)
			if !ok {
				log.Println("Invalid SDP format")
				continue
			}

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
				break
			}

			// 연결 저장
			wsc.connections[c] = peerConnection

			// Offer 설정
			if err := peerConnection.SetRemoteDescription(offer); err != nil {
				log.Println("Failed to set remote description:", err)
				break
			}

			// Answer 생성
			answer, err := peerConnection.CreateAnswer(nil)
			if err != nil {
				log.Println("Failed to create Answer:", err)
				break
			}

			// LocalDescription 설정
			if err := peerConnection.SetLocalDescription(answer); err != nil {
				log.Println("Failed to set local description:", err)
				break
			}

			// Answer 전송
			response := map[string]interface{}{
				"type": "answer",
				"sdp":  answer.SDP,
			}
			responseJSON, _ := json.Marshal(response)
			c.WriteMessage(websocket.TextMessage, responseJSON)

			// 큐에 있는 ICE Candidate 처리
			if candidates, exists := wsc.iceCandidateQueue[c]; exists {
				for _, candidate := range candidates {
					if err := peerConnection.AddICECandidate(candidate); err != nil {
						log.Println("Failed to add ICE Candidate from queue:", err)
					}
				}
				delete(wsc.iceCandidateQueue, c)
			}

		case "iceCandidate":
			candidateMap, ok := payload["candidate"].(map[string]interface{})
			if !ok {
				log.Println("Invalid candidate format")
				continue
			}

			sdpMLineIndex := uint16(candidateMap["sdpMLineIndex"].(float64))
			candidate := webrtc.ICECandidateInit{
				Candidate:        candidateMap["candidate"].(string),
				SDPMid:           func(s string) *string { return &s }(candidateMap["sdpMid"].(string)),
				SDPMLineIndex:    &sdpMLineIndex,
				UsernameFragment: func(s string) *string { return &s }(candidateMap["usernameFragment"].(string)),
			}

			if pc, exists := wsc.connections[c]; exists && pc != nil {
				if err := pc.AddICECandidate(candidate); err != nil {
					log.Println("Failed to add ICE Candidate:", err)
				}
				log.Println("Added ICE Candidate:", candidate)
			} else {
				// 큐에 추가
				wsc.iceCandidateQueue[c] = append(wsc.iceCandidateQueue[c], candidate)
				log.Println("Queued ICE Candidate:", candidate)
			}

		default:
			log.Println("Unknown message type:", payload["type"])
		}
	}
}
