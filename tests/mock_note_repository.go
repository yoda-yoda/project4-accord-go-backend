package tests

import (
	"errors"
	"sync"
	"time"

	"go-server/models"
	"go-server/utils"
)

type MockNoteRepository struct {
	data map[string]models.Note
	mu   sync.RWMutex
}

func NewMockNoteRepository() *MockNoteRepository {
	return &MockNoteRepository{
		data: make(map[string]models.Note),
	}
}

func (m *MockNoteRepository) SaveNote(note models.Note) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if note.Title == "fail" {
		return "", errors.New("failed to save note")
	}
	if note.ID == "" {
		note.ID = utils.GenerateID()
	}
	note.CreatedAt = time.Now()
	m.data[note.ID] = note
	return note.ID, nil
}

// FindNoteByID는 주어진 ID로 note를 찾습니다.
func (m *MockNoteRepository) FindNoteByID(id string) (models.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	note, ok := m.data[id]
	if !ok {
		return models.Note{}, errors.New("note not found")
	}
	return note, nil
}

func (m *MockNoteRepository) FindNotesByTeamID(teamID string) ([]models.Note, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var notes []models.Note
	for _, note := range m.data {
		if note.TeamID == teamID {
			notes = append(notes, note)
		}
	}
	return notes, nil
}

func (m *MockNoteRepository) UpdateNoteTitle(id, newTitle string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "fail" {
		return errors.New("failed to update note title")
	}
	note, ok := m.data[id]
	if !ok {
		return errors.New("note not found")
	}
	note.Title = newTitle
	m.data[id] = note
	return nil
}

func (m *MockNoteRepository) DeleteNoteByID(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "fail" {
		return errors.New("failed to delete note")
	}
	_, ok := m.data[id]
	if !ok {
		return errors.New("note not found")
	}
	delete(m.data, id)
	return nil
}
