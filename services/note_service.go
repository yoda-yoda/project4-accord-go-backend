package service

import (
	"encoding/json"
	"fmt"
	"go-server/models"
	"go-server/repository"
	"log"
	"time"
)

// NoteService: CRDT + MongoDB 연동
type NoteService struct {
	noteRepo *repository.NoteRepository
}

func NewNoteService(noteRepo *repository.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}

// HandleNoteChange: 클라이언트 -> Change -> CRDT 적용 -> DB -> 결과 반환
func (ns *NoteService) HandleNoteChange(teamID string, change Change) (string, int, error) {
	// 1) DB에서 기존 노트 읽기
	note, err := ns.noteRepo.FindNoteByTeamID(teamID)
	if err != nil {
		// 찾지 못하면 새로 생성
		log.Printf("No existing note for teamID=%s -> create new.\n", teamID)
		note = models.Note{
			TeamID:  teamID,
			Version: 0,
			Note: `{
			"type": "doc",
			"content": []
			}`, // 빈 문서
		}
	}

	// 2) CRDT 초기화
	crdtObj := NewCRDT()
	crdtObj.Version = note.Version

	// 기존 노트 JSON -> map[string]interface{}
	docMap, err := parseJSON(note.Note)
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse existing note: %v", err)
	}
	crdtObj.Document = docMap

	// 3) 변경 적용
	change.Timestamp = time.Now()
	if applyErr := crdtObj.ApplyChange(change); applyErr != nil {
		return "", 0, fmt.Errorf("applyChange error: %v", applyErr)
	}

	// 만약 실제로 적용된(버전이 올라갔는지) 확인
	if crdtObj.Version > note.Version {
		// 4) CRDT(Document) 다시 JSON 문자열로
		updatedStr, err := toJSON(crdtObj.Document)
		if err != nil {
			return "", 0, fmt.Errorf("failed to convert doc to JSON: %v", err)
		}

		note.Note = updatedStr
		note.Version = crdtObj.Version
		note.CreatedAt = time.Now()

		// DB 저장
		if err := ns.noteRepo.SaveNote(note); err != nil {
			return "", 0, err
		}

		return updatedStr, note.Version, nil
	}

	// 변경 없음
	return note.Note, note.Version, nil
}

// parseJSON: 문자열 -> map
func parseJSON(s string) (map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return nil, err
	}
	return m, nil
}

// toJSON: map -> 문자열
func toJSON(m map[string]interface{}) (string, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
