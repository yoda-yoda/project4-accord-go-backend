package service

import (
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
)

type WebSocketService struct {
	Rooms map[string]map[*websocket.Conn]bool // 방 관리
	mu    sync.Mutex                          // 동시성 제어를 위한 Mutex
}

func NewWebSocketService() *WebSocketService {
	return &WebSocketService{
		Rooms: make(map[string]map[*websocket.Conn]bool),
	}
}

// 클라이언트를 특정 방에 추가
func (s *WebSocketService) Subscribe(room string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Rooms[room]; !exists {
		s.Rooms[room] = make(map[*websocket.Conn]bool)
	}
	s.Rooms[room][conn] = true
	log.Printf("Client subscribed to room: %s\n", room)
}

// 방의 모든 클라이언트에 메시지 브로드캐스트
func (s *WebSocketService) Publish(room string, message []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clients, exists := s.Rooms[room]; exists {
		for client := range clients {
			if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Println("Error sending message:", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

// 방에서 클라이언트 제거
func (s *WebSocketService) RemoveClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for room, clients := range s.Rooms {
		if _, exists := clients[conn]; exists {
			delete(clients, conn)
			log.Printf("Client removed from room: %s\n", room)

			// 방에 클라이언트가 없으면 방 삭제
			if len(clients) == 0 {
				delete(s.Rooms, room)
			}
		}
	}
}
