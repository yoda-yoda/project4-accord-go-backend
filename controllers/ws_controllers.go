package controllers

import (
	"encoding/json"
	"log"
	"strings"

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
	noteService        *service.NoteService
}

func NewWebSocketController(
	participantService *service.ParticipantService,
	wsService *service.WebSocketService,
	noteService *service.NoteService,
) *WebSocketController {
	return &WebSocketController{
		participantService: participantService,
		wsService:          wsService,
		noteService:        noteService,
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

		label := d.Label()
		var teamID string
		if parts := strings.Split(label, "/"); len(parts) == 2 && parts[0] == "note" {
			teamID = parts[1]
		}

		// 저장
		wsc.dataChannels[c] = d

		// Handle DataChannel messages
		d.OnMessage(func(msg webrtc.DataChannelMessage) {
			log.Printf("Received message: %s\n", string(msg.Data))

			// 1) JSON 파싱
			var m map[string]interface{}
			if err := json.Unmarshal(msg.Data, &m); err != nil {
				log.Println("Failed to unmarshal JSON from data channel:", err)
				return
			}

			// 예: { "type": "note", "version": 3, "steps": [...], "clientID": "xxx" }
			msgType, _ := m["type"].(string)
			if msgType == "note" {
				// 2) 서버 관점에서 필요한 로직(추가 저장이나 버전 관리 등)이 있으면 처리
				change := service.Change{
					Type: "note",
				}
				// version
				if v, ok := m["version"].(float64); ok {
					change.Version = int(v)
				}
				// clientID
				if cid, ok := m["clientID"].(string); ok {
					change.ClientID = cid
				}
				// steps
				if steps, ok := m["steps"].([]interface{}); ok {
					for _, s := range steps {
						// s는 {"stepType":"replace","from":0,"to":2,"slice":{...}}
						stepBytes, _ := json.Marshal(s)
						var step service.Step
						if err := json.Unmarshal(stepBytes, &step); err == nil {
							change.Steps = append(change.Steps, step)
						}
					}
				}

				updatedDoc, newVersion, err := wsc.noteService.HandleNoteChange(teamID, change)
				if err != nil {
					log.Println("Error applying note change:", err)
					return
				}

				// 3) 보내준 클라이언트(혹은 전체)에게 "ackSteps" 형태로 응답

				// ACK
				ack := map[string]interface{}{
					"type":     "ackSteps",
					"version":  newVersion,
					"clientID": change.ClientID,
					"doc":      updatedDoc,
				}
				ackBytes, err := json.Marshal(ack)
				if err == nil {
					// **지금 메시지를 보낸 원 클라이언트**에게도 돌려주기
					// (여기선 `d`가 해당 DataChannel)
					if sendErr := d.SendText(string(ackBytes)); sendErr != nil {
						log.Println("Failed to send ack to the sender:", sendErr)
					}

					// 그리고 다른 클라이언트에게도 보내고 싶다면:
					for conn, channel := range wsc.dataChannels {
						if conn != c && channel != nil {
							_ = channel.SendText(string(ackBytes))
						}
					}
				} else {
					log.Println("Failed to marshal ackSteps:", err)
				}
			} else {
				// note가 아닌 경우(기존 broadcast 로직)
				for conn, channel := range wsc.dataChannels {
					if conn != c && channel != nil {
						err := channel.SendText(string(msg.Data))
						if err != nil {
							log.Printf("Failed to send message to client: %v", err)
						}
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
