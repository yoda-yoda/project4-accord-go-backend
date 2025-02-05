package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"go-server/models"

	"github.com/fasthttp/websocket"
	adaptor "github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	fiberws "github.com/gofiber/websocket/v2"
	"github.com/stretchr/testify/assert"
)

// MockParticipantRepository 구조체 정의
type MockParticipantRepository struct {
	participants map[string]map[string]models.Participant // key: "teamID:kind", value: map[participantID]Participant
	mu           sync.Mutex
}

func NewMockParticipantRepository() *MockParticipantRepository {
	return &MockParticipantRepository{
		participants: make(map[string]map[string]models.Participant),
	}
}

func (m *MockParticipantRepository) AddParticipant(ctx context.Context, teamID, kind string, p models.Participant) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", teamID, kind)
	if _, exists := m.participants[key]; !exists {
		m.participants[key] = make(map[string]models.Participant)
	}
	m.participants[key][p.ID] = p
	return nil
}

func (m *MockParticipantRepository) RemoveParticipant(ctx context.Context, teamID, kind, participantID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", teamID, kind)
	if participants, exists := m.participants[key]; exists {
		delete(participants, participantID)
		if len(participants) == 0 {
			delete(m.participants, key)
		}
	}
	return nil
}

func (m *MockParticipantRepository) GetParticipants(ctx context.Context, teamID, kind string) ([]models.Participant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s:%s", teamID, kind)
	participants := make([]models.Participant, 0)
	if pMap, exists := m.participants[key]; exists {
		for _, p := range pMap {
			participants = append(participants, p)
		}
	}
	return participants, nil
}

func TestFiberWebSocket(t *testing.T) {
	app := fiber.New()
	app.Get("/ws", fiberws.New(func(c *fiberws.Conn) {
		defer c.Close()
		// 메시지 읽고 에코 등등
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				return
			}
			_ = c.WriteMessage(mt, msg)
		}
	}))

	ts := httptest.NewServer(adaptor.FiberApp(app))
	defer ts.Close()

	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"

	// fasthttp/websocket Dialer 사용
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	err = conn.WriteMessage(websocket.TextMessage, []byte("hello!"))
	assert.NoError(t, err)

	_, resp, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "hello!", string(resp))
}

func TestParticipantsController_AddAndBroadcastParticipant(t *testing.T) {
	mockRepo := NewMockParticipantRepository()
	controller := NewParticipantsController(mockRepo)

	// Fiber + Fiber 웹소켓 설정
	app := fiber.New()

	// WebSocket 업그레이드 체크 미들웨어
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(controller.HandleWebSocket))

	// httptest.NewServer로 Fiber 앱 감싸기
	ts := httptest.NewServer(adaptor.FiberApp(app))
	defer ts.Close()

	// WebSocket 연결 (fasthttp/websocket Dialer)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(wsURL, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// 메시지: addParticipant
	testParticipant := map[string]string{
		"action":         "addParticipant",
		"team_id":        "team1",
		"kind":           "canvas",
		"participant":    "user1",
		"name":           "Test User",
		"profilePicture": "test.jpg",
		"color":          "#ffffff",
	}
	msg, _ := json.Marshal(testParticipant)

	err = conn.WriteMessage(websocket.TextMessage, msg)
	assert.NoError(t, err)

	// 응답 수신
	_, resp, err := conn.ReadMessage()
	fmt.Println(string(resp))
	assert.NoError(t, err)

	var response map[string]interface{}
	_ = json.Unmarshal(resp, &response)

	assert.Equal(t, "updateParticipants", response["action"])
	participants := response["participants"].([]interface{})
	assert.Len(t, participants, 1)

	// Mock Repo 상태 확인
	parts, err := mockRepo.GetParticipants(context.Background(), "team1", "canvas")
	assert.NoError(t, err)
	assert.Len(t, parts, 1)
	assert.Equal(t, "user1", parts[0].ID)
}

func TestParticipantsController_RemoveParticipant(t *testing.T) {
	mockRepo := NewMockParticipantRepository()
	controller := NewParticipantsController(mockRepo)

	app := fiber.New()
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(controller.HandleWebSocket))

	ts := httptest.NewServer(adaptor.FiberApp(app))
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	dialer := websocket.Dialer{}

	// 첫 번째 연결(참가자 추가)
	conn1, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn1.Close()

	addMsg := map[string]string{
		"action":         "addParticipant",
		"team_id":        "team1",
		"kind":           "canvas",
		"participant":    "user1",
		"name":           "User One",
		"profilePicture": "pic1.jpg",
		"color":          "red",
	}
	addBytes, _ := json.Marshal(addMsg)
	conn1.WriteMessage(websocket.TextMessage, addBytes)
	conn1.ReadMessage() // 응답 소진

	// 두 번째 연결(참가자 제거)
	conn2, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn2.Close()

	removeMsg := map[string]string{
		"action":      "removeParticipant",
		"team_id":     "team1",
		"kind":        "canvas",
		"participant": "user1",
	}
	removeBytes, _ := json.Marshal(removeMsg)
	conn2.WriteMessage(websocket.TextMessage, removeBytes)

	// 응답 수신
	_, resp, err := conn2.ReadMessage()
	assert.NoError(t, err)

	var response map[string]interface{}
	_ = json.Unmarshal(resp, &response)
	assert.Equal(t, "updateParticipants", response["action"])
	assert.Empty(t, response["participants"])

	// 리포지토리 확인
	parts, _ := mockRepo.GetParticipants(context.Background(), "team1", "canvas")
	assert.Len(t, parts, 0)
}

func TestParticipantsController_DisconnectCleanup(t *testing.T) {
	// 여기서는 "disconnect 시에 해당 참가자가 제거되었는지"를 확인하는 예시
	// 실제 구현에서 controller.HandleWebSocket 내부에
	// defer/cleanup 로직을 넣어야 동작합니다.
	mockRepo := NewMockParticipantRepository()
	controller := NewParticipantsController(mockRepo)

	app := fiber.New()
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(controller.HandleWebSocket))

	ts := httptest.NewServer(adaptor.FiberApp(app))
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)

	// 참가자 추가
	addMsg := map[string]string{
		"action":         "addParticipant",
		"team_id":        "team1",
		"kind":           "note",
		"participant":    "user2",
		"name":           "User Two",
		"profilePicture": "pic2.jpg",
		"color":          "blue",
	}
	addBytes, _ := json.Marshal(addMsg)
	conn.WriteMessage(websocket.TextMessage, addBytes)
	conn.ReadMessage() // 응답 소진

	// 연결 종료
	conn.Close()
	time.Sleep(100 * time.Millisecond)

	// 참가자 제거 되었는지 확인
	parts, _ := mockRepo.GetParticipants(context.Background(), "team1", "note")
	assert.Len(t, parts, 0)
}

func TestParticipantsController_AudioParticipants(t *testing.T) {
	mockRepo := NewMockParticipantRepository()
	controller := NewParticipantsController(mockRepo)

	app := fiber.New()
	app.Use("/ws", func(c *fiber.Ctx) error {
		if fiberws.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})
	app.Get("/ws", fiberws.New(controller.HandleWebSocket))

	ts := httptest.NewServer(adaptor.FiberApp(app))
	defer ts.Close()

	url := "ws" + strings.TrimPrefix(ts.URL, "http") + "/ws"
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(url, nil)
	assert.NoError(t, err)
	defer conn.Close()

	// 오디오 참가자 추가
	audioMsg := map[string]string{
		"action":         "addAudioParticipant",
		"team_id":        "team2",
		"kind":           "audio",
		"participant":    "user3",
		"name":           "Audio User",
		"profilePicture": "audio.jpg",
		"color":          "green",
	}
	audioBytes, _ := json.Marshal(audioMsg)
	conn.WriteMessage(websocket.TextMessage, audioBytes)

	// 응답 (updateAudioParticipants) 확인
	_, resp, err := conn.ReadMessage()
	assert.NoError(t, err)
	var response map[string]interface{}
	_ = json.Unmarshal(resp, &response)
	assert.Equal(t, "updateAudioParticipants", response["action"])
	audioParticipants := response["audioParticipants"].([]interface{})
	assert.Len(t, audioParticipants, 1)

	// 오디오 참가자 조회
	getMsg := map[string]string{
		"action":  "getAudioParticipants",
		"team_id": "team2",
		"kind":    "audio",
	}
	getBytes, _ := json.Marshal(getMsg)
	conn.WriteMessage(websocket.TextMessage, getBytes)

	_, resp2, err := conn.ReadMessage()
	assert.NoError(t, err)
	var getResp map[string]interface{}
	_ = json.Unmarshal(resp2, &getResp)
	assert.Equal(t, "getAudioParticipants", getResp["action"])
	assert.Len(t, getResp["audioParticipants"], 1)
}
