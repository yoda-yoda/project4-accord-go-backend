package tests

import (
	"errors"
	"sync"
	"time"

	"go-server/models"
	"go-server/utils"
)

type MockCanvasRepository struct {
	data map[string]models.Canvas
	mu   sync.RWMutex
}

func NewMockCanvasRepository() *MockCanvasRepository {
	return &MockCanvasRepository{
		data: make(map[string]models.Canvas),
	}
}

func (m *MockCanvasRepository) SaveCanvas(canvas models.Canvas) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if canvas.Title == "fail" {
		return "", errors.New("failed to save canvas")
	}
	if canvas.ID == "" {
		canvas.ID = utils.GenerateID()
	}
	canvas.CreatedAt = time.Now()
	m.data[canvas.ID] = canvas
	return canvas.ID, nil
}

func (m *MockCanvasRepository) FindCanvasByID(id string) (models.Canvas, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	canvas, ok := m.data[id]
	if !ok {
		return models.Canvas{}, errors.New("canvas not found")
	}
	return canvas, nil
}

func (m *MockCanvasRepository) FindCanvasesByTeamID(teamID string) ([]models.Canvas, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var canvases []models.Canvas
	for _, canvas := range m.data {
		if canvas.TeamID == teamID {
			canvases = append(canvases, canvas)
		}
	}
	return canvases, nil
}

func (m *MockCanvasRepository) UpdateCanvasTitle(id, newTitle string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "fail" {
		return errors.New("failed to update canvas title")
	}
	canvas, ok := m.data[id]
	if !ok {
		return errors.New("canvas not found")
	}
	canvas.Title = newTitle
	m.data[id] = canvas
	return nil
}

func (m *MockCanvasRepository) DeleteCanvasByID(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "fail" {
		return errors.New("failed to delete canvas")
	}
	_, ok := m.data[id]
	if !ok {
		return errors.New("canvas not found")
	}
	delete(m.data, id)
	return nil
}
