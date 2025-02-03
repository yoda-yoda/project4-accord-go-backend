package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"go-server/controllers"
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupNoteApp() *fiber.App {
	app := fiber.New()
	var repo repository.NoteRepositoryInterface = NewMockNoteRepository()

	noteController := controllers.NewNoteController(repo)

	app.Post("/notes", noteController.CreateNote)
	app.Get("/notes/:id", noteController.GetNoteByID)
	app.Get("/notes/team/:teamId", noteController.GetNotesByTeamID)
	app.Put("/notes/:id/title", noteController.UpdateNoteTitle)
	app.Delete("/notes/:id", noteController.DeleteNoteByID)

	return app
}

func TestCreateNote_Success(t *testing.T) {
	app := setupNoteApp()

	note := models.Note{
		Title:  "New Note",
		TeamID: "team123",
		Note:   "Some note content",
	}
	body, _ := json.Marshal(note)
	req := httptest.NewRequest("POST", "/notes", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, respBody["id"])
}

func TestCreateNote_InvalidJSON(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("POST", "/notes", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid JSON", respBody["error"])
}

func TestGetNoteByID_Success(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("GET", "/notes/12345", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var note models.Note
	_ = json.NewDecoder(resp.Body).Decode(&note)
	assert.Equal(t, "12345", note.ID)
	assert.Equal(t, "Test Note", note.Title)
}

func TestGetNoteByID_NotFound(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("GET", "/notes/not-found", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Note not found", respBody["error"])
}

func TestGetNotesByTeamID(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("GET", "/notes/team/team123", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var notes []models.Note
	_ = json.NewDecoder(resp.Body).Decode(&notes)
	// 위 setup에서 team123에 "Test Note"를 넣었으므로 1개는 있을 것
	// 추가로 POST 해도 되니 상황에 따라 테스트를 조절하세요.
	assert.Len(t, notes, 1)
	assert.Equal(t, "Test Note", notes[0].Title)
}

func TestUpdateNoteTitle_Success(t *testing.T) {
	app := setupNoteApp()

	reqBody := map[string]string{"new_title": "Updated Title"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/notes/12345/title", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "success", respBody["status"])

	// 실제로 변경되었는지 한번 더 확인 가능(목 리포지토리에서 재조회 등)
}

func TestUpdateNoteTitle_InvalidJSON(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("PUT", "/notes/12345/title", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid JSON", respBody["error"])
}

func TestDeleteNoteByID_Success(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("DELETE", "/notes/12345", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "success", respBody["status"])
}

func TestDeleteNoteByID_Failure(t *testing.T) {
	app := setupNoteApp()

	req := httptest.NewRequest("DELETE", "/notes/fail", nil)
	resp, err := app.Test(req, -1)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Failed to delete note", respBody["error"])
}
