package controllers

import (
	"bytes"
	"testing"
)

// mockConn은 단위 테스트용으로 WSConn 인터페이스를 구현합니다.
type mockConn struct {
	sentMessages [][]byte
	closed       bool
}

func (m *mockConn) WriteMessage(messageType int, data []byte) error {
	// 데이터 복사하여 저장
	m.sentMessages = append(m.sentMessages, append([]byte(nil), data...))
	return nil
}

func (m *mockConn) Close() error {
	m.closed = true
	return nil
}

// TestBroadcastSignal는 broadcastSignal 함수의 로직을 단위 테스트합니다.
func TestBroadcastSignal(t *testing.T) {
	wsc := NewWebSocketController()
	roomID := "testRoom"
	sender := &mockConn{}
	receiver := &mockConn{}

	// 방을 구성: sender와 receiver 두 클라이언트를 추가
	wsc.rooms = make(map[string]map[WSConn]bool)
	wsc.rooms[roomID] = map[WSConn]bool{
		sender:   true,
		receiver: true,
	}

	// 원본 메시지
	rawMessage := []byte("test message")

	// broadcastSignal 호출 (sender 제외한 다른 클라이언트에 메시지 전송)
	wsc.broadcastSignal(roomID, sender, rawMessage)

	// receiver가 메시지를 1건 수신했는지 확인
	if len(receiver.sentMessages) != 1 {
		t.Errorf("Expected receiver to get 1 message, got %d", len(receiver.sentMessages))
	} else if !bytes.Equal(receiver.sentMessages[0], rawMessage) {
		t.Errorf("Expected message %s, got %s", string(rawMessage), string(receiver.sentMessages[0]))
	}

	// sender에게는 메시지가 전송되지 않아야 함
	if len(sender.sentMessages) != 0 {
		t.Errorf("Expected sender to not receive own message, got %d messages", len(sender.sentMessages))
	}
}
