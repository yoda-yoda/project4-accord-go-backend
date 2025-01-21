package controllers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

// SignalMessage: y-webrtc가 사용하는 시그널링 메시지 구조
// 예) {"type":"signal","to":"...","from":"...","room":"...","data":{...}}
type SignalMessage struct {
	Type string                 `json:"type"`
	To   string                 `json:"to"`
	From string                 `json:"from"`
	Room string                 `json:"room"`
	Data map[string]interface{} `json:"data"`
}

// WebSocketController: (예시) 방(room) 단위로 WebSocket 접속을 관리
type WebSocketController struct {
	mu    sync.Mutex
	rooms map[string]map[*websocket.Conn]bool
}

// NewWebSocketController: 간단 초기화
func NewWebSocketController() *WebSocketController {
	return &WebSocketController{
		rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

// HandleYWebRTC: y-webrtc 시그널링 전용 WebSocket 핸들러
func (wsc *WebSocketController) HandleYWebRTC(c *websocket.Conn) {
	// 1) Query "room" 파라미터
	roomID := c.Query("room")
	if roomID == "" {
		log.Println("[y-webrtc] No 'room' query param provided, closing.")
		_ = c.Close()
		return
	}

	// 2) 방에 접속
	wsc.joinRoom(roomID, c)
	defer wsc.leaveRoom(roomID, c)

	log.Printf("[y-webrtc] Client joined room=%s\n", roomID)

	// 3) 메시지 루프
	for {
		msgType, msg, err := c.ReadMessage()
		if err != nil {
			log.Printf("[y-webrtc] Read error: %v\n", err)
			break
		}
		if msgType != websocket.TextMessage {
			// y-webrtc는 주로 TextMessage 사용. 아닌 경우 무시
			continue
		}

		// y-webrtc에서 오는 메시지를 구조체로 파싱 (type, from, to, room, data...)
		var signal SignalMessage
		if err := json.Unmarshal(msg, &signal); err != nil {
			log.Printf("[y-webrtc] JSON parse error: %v\n", err)
			continue
		}

		// room 필드가 비어있다면 보완
		if signal.Room == "" {
			signal.Room = roomID
		}

		// 4) y-webrtc가 'type: "signal"'을 보냈다면, 그대로 방 전체에 브로드캐스트
		switch signal.Type {
		case "signal":
			// offer/answer/iceCandidate 등 협업에 필요한 핵심 메시지
			wsc.broadcastSignal(roomID, c, msg)

		// 필요하다면, "participants"나 "awareness" 등 다른 타입도 그대로 중계 가능
		case "participants":
			// 예: 참가자 목록 메시지도 as-is로 뿌리고 싶다면
			wsc.broadcastSignal(roomID, c, msg)

		default:
			// 그 외 타입은 무시하거나, 로그만 찍기
			log.Printf("[y-webrtc] Unknown type=%s message=%s\n", signal.Type, string(msg))
		}
	}
}

// joinRoom: 해당 roomID에 WebSocket 연결 추가
func (wsc *WebSocketController) joinRoom(roomID string, conn *websocket.Conn) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	if wsc.rooms[roomID] == nil {
		wsc.rooms[roomID] = make(map[*websocket.Conn]bool)
	}
	wsc.rooms[roomID][conn] = true
}

// leaveRoom: roomID에서 해당 WebSocket 연결 제거
func (wsc *WebSocketController) leaveRoom(roomID string, conn *websocket.Conn) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	if clients, ok := wsc.rooms[roomID]; ok {
		if _, exists := clients[conn]; exists {
			delete(clients, conn)
			_ = conn.Close()
			log.Printf("[y-webrtc] Client left room=%s\n", roomID)
		}
		if len(clients) == 0 {
			delete(wsc.rooms, roomID)
		}
	}
}

// broadcastSignal: 받은 메시지를 방 안의 다른 클라이언트에게 그대로 전송
func (wsc *WebSocketController) broadcastSignal(roomID string, sender *websocket.Conn, rawMessage []byte) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	clients, ok := wsc.rooms[roomID]
	if !ok {
		return
	}

	// 여기서는 rawMessage(msg) 그 자체를 그대로 중계
	for conn := range clients {
		// 자기 자신(sender)에게는 보내지 않음
		if conn == sender {
			continue
		}
		if err := conn.WriteMessage(websocket.TextMessage, rawMessage); err != nil {
			log.Println("[y-webrtc] Write error:", err)
			_ = conn.Close()
			delete(clients, conn)
		}
	}
}
