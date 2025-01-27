package controllers

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/pion/webrtc/v4"
)

type AudioSocketController struct {
	mu          sync.Mutex
	teams       map[string]map[*websocket.Conn]*webrtc.PeerConnection
	teamsTracks map[string]map[*websocket.Conn][]*webrtc.TrackLocalStaticRTP
}

func NewAudioSocketController() *AudioSocketController {
	return &AudioSocketController{
		teams:       make(map[string]map[*websocket.Conn]*webrtc.PeerConnection),
		teamsTracks: make(map[string]map[*websocket.Conn][]*webrtc.TrackLocalStaticRTP),
	}
}

func (wsc *AudioSocketController) HandleWebRTC(c *websocket.Conn) {
	defer func() {
		wsc.cleanupConnection(c)
		c.Close()
	}()

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
			wsc.handleOffer(c, payload)

		case "answer":
			wsc.handleAnswer(c, payload)

		case "iceCandidate":
			wsc.handleICECandidate(c, payload)
		}
	}
}

func (asc *AudioSocketController) cleanupConnection(c *websocket.Conn) {
	asc.mu.Lock()
	defer asc.mu.Unlock()

	// 모든 팀을 돌면서 해당 conn이 있는지 찾고 제거
	for teamID, connMap := range asc.teams {
		if pc, ok := connMap[c]; ok {
			pc.Close()
			delete(connMap, c)
			if len(connMap) == 0 {
				delete(asc.teams, teamID)
			}

			// 트랙 맵도 제거
			if trackMap, ok2 := asc.teamsTracks[teamID]; ok2 {
				delete(trackMap, c)
				if len(trackMap) == 0 {
					delete(asc.teamsTracks, teamID)
				}
			}
			break
		}
	}
}

// 클라이언트가 처음 보낸 Offer를 처리 -> 서버가 Answer
func (wsc *AudioSocketController) handleOffer(c *websocket.Conn, payload map[string]interface{}) {
	teamID, _ := payload["teamId"].(string)
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

	// SFU: 오디오 sendrecv
	_, err = peerConnection.AddTransceiverFromKind(
		webrtc.RTPCodecTypeAudio,
		webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionSendrecv},
	)
	if err != nil {
		log.Println("AddTransceiverFromKind failed:", err)
	}

	// 1) OnNegotiationNeeded: 서버가 새 트랙을 추가하면 이 콜백이 뜸 -> 서버가 re-offer
	peerConnection.OnNegotiationNeeded(func() {
		log.Println("[OnNegotiationNeeded] => CreateOffer from server side")
		wsc.handleServerNegotiation(c, peerConnection)
	})

	// (2) OnTrack -> 같은 팀의 다른 피어들에게만 RTP 중계
	peerConnection.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Printf("Got remote track from team=%s conn=%p, trackID=%s\n", teamID, c, remoteTrack.ID())

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
		)
		if err != nil {
			log.Println("Failed to create local track:", err)
			return
		}

		wsc.mu.Lock()
		// 같은 팀만 순회
		for otherConn, otherPC := range wsc.teams[teamID] {
			if otherConn == c {
				continue
			}
			if sender, addErr := otherPC.AddTrack(localTrack); addErr != nil {
				log.Printf("AddTrack error: %v\n", addErr)
			} else {
				log.Printf("Forward track to team=%s conn=%p via sender=%v\n", teamID, otherConn, sender)
			}
		}
		// 이 conn이 소유한 localTrack 목록에 저장
		wsc.teamsTracks[teamID][c] = append(wsc.teamsTracks[teamID][c], localTrack)
		wsc.mu.Unlock()

		// RTP를 localTrack으로 계속 포워딩
		go func() {
			rtpBuf := make([]byte, 1400)
			for {
				n, _, readErr := remoteTrack.Read(rtpBuf)
				if readErr != nil {
					log.Println("remoteTrack read error:", readErr)
					return
				}
				if _, writeErr := localTrack.Write(rtpBuf[:n]); writeErr != nil {
					log.Println("localTrack write error:", writeErr)
					return
				}
			}
		}()
	})

	// (3) 팀 맵에 등록
	wsc.mu.Lock()
	if wsc.teams[teamID] == nil {
		wsc.teams[teamID] = make(map[*websocket.Conn]*webrtc.PeerConnection)
	}
	if wsc.teamsTracks[teamID] == nil {
		wsc.teamsTracks[teamID] = make(map[*websocket.Conn][]*webrtc.TrackLocalStaticRTP)
	}

	wsc.teams[teamID][c] = peerConnection
	wsc.teamsTracks[teamID][c] = []*webrtc.TrackLocalStaticRTP{}
	wsc.mu.Unlock()

	// 4) SetRemoteDescription(offer) → CreateAnswer → SetLocalDescription(answer)
	if err = peerConnection.SetRemoteDescription(offer); err != nil {
		log.Println("Failed to set remote description:", err)
		return
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create Answer:", err)
		return
	}
	if err = peerConnection.SetLocalDescription(answer); err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	// 5) answer -> 클라이언트로
	response := map[string]interface{}{
		"type": "answer",
		"sdp":  answer.SDP,
	}
	resBytes, _ := json.Marshal(response)
	if err := c.WriteMessage(websocket.TextMessage, resBytes); err != nil {
		log.Println("WriteMessage error:", err)
	}

	// 6) 이미 존재하던 다른 사람들의 track도 이 유저에게 addTrack (재협상 필요)
	wsc.mu.Lock()
	for otherConn, otherLocalTracks := range wsc.teamsTracks[teamID] {
		if otherConn == c {
			continue
		}
		for _, lt := range otherLocalTracks {
			if sender, err := peerConnection.AddTrack(lt); err != nil {
				log.Println("AddTrack for existing track error:", err)
			} else {
				log.Printf("conn=%p: added existing track from %p (sender=%v)\n", c, otherConn, sender)
			}
		}
	}
	wsc.mu.Unlock()
	// 여기서 AddTrack이 일어나므로 -> peerConnection.OnNegotiationNeeded 콜백이 발생
	// -> handleServerNegotiation(...)에서 re-offer를 보냄
}

func (wsc *AudioSocketController) handleServerNegotiation(c *websocket.Conn, pc *webrtc.PeerConnection) {
	offer, err := pc.CreateOffer(nil)
	if err != nil {
		log.Println("CreateOffer error:", err)
		return
	}
	if err := pc.SetLocalDescription(offer); err != nil {
		log.Println("SetLocalDescription error:", err)
		return
	}

	msg := map[string]interface{}{
		"type": "offer",
		"sdp":  offer.SDP,
	}
	data, _ := json.Marshal(msg)
	_ = c.WriteMessage(websocket.TextMessage, data)
	log.Println("[Server -> Client] re-offer sent")
}

// 클라이언트가 re-offer에 대한 answer(혹은 서버 offer에 대한 answer)를 보냈을 때
func (wsc *AudioSocketController) handleAnswer(c *websocket.Conn, payload map[string]interface{}) {
	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	for _, connMap := range wsc.teams {
		if pc, ok := connMap[c]; ok {
			sdp, _ := payload["sdp"].(string)
			answer := webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  sdp,
			}
			if err := pc.SetRemoteDescription(answer); err != nil {
				log.Println("handleAnswer: SetRemoteDescription error:", err)
			} else {
				log.Println("handleAnswer: remoteDescription set (answer)")
			}
			return
		}
	}
	log.Println("handleAnswer: no PeerConnection found for this client")
}

func (wsc *AudioSocketController) handleICECandidate(c *websocket.Conn, payload map[string]interface{}) {
	candidateMap, _ := payload["candidate"].(map[string]interface{})
	candidate := webrtc.ICECandidateInit{
		Candidate: candidateMap["candidate"].(string),
		SDPMid:    func(s string) *string { return &s }(candidateMap["sdpMid"].(string)),
	}

	wsc.mu.Lock()
	defer wsc.mu.Unlock()

	for _, connMap := range wsc.teams {
		if pc, ok := connMap[c]; ok {
			if pc != nil {
				if err := pc.AddICECandidate(candidate); err != nil {
					log.Println("AddICECandidate error:", err)
				}
			}
			return
		}
	}
}
