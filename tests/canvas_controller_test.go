package tests

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"

	"go-server/controllers"
	"go-server/models"
	"go-server/repository"
)

func setupCanvasApp() *fiber.App {
	app := fiber.New()
	var repo repository.CanvasRepositoryInterface = NewMockCanvasRepository()
	canvasController := controllers.NewCanvasController(repo)

	app.Post("/canvas", canvasController.CreateCanvas)
	app.Get("/canvas/:id", canvasController.GetCanvasByID)
	app.Get("/canvases/team/:teamId", canvasController.GetCanvasesByTeamID)
	app.Put("/canvas/:id/title", canvasController.UpdateCanvasTitle)
	app.Delete("/canvas/:id", canvasController.DeleteCanvasByID)

	return app
}

func TestCreateCanvas_Success(t *testing.T) {
	app := setupCanvasApp()

	canvas := models.Canvas{
		Title:  "New Canvas",
		TeamID: "team1",
		Canvas: "Canvas content",
	}
	body, _ := json.Marshal(canvas)
	req := httptest.NewRequest("POST", "/canvas", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.NotEmpty(t, respBody["id"])
}

func TestCreateCanvas_InvalidJSON(t *testing.T) {
	app := setupCanvasApp()

	req := httptest.NewRequest("POST", "/canvas", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "Invalid JSON", respBody["error"])
}

func TestGetCanvasByID_Success(t *testing.T) {
	app := setupCanvasApp()

	canvas1 := models.Canvas{
		Title:  "New Canvas",
		TeamID: "team1",
		Canvas: "Canvas content",
	}
	body, _ := json.Marshal(canvas1)
	req1 := httptest.NewRequest("POST", "/canvas", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	app.Test(req1)

	req2 := httptest.NewRequest("GET", "/canvas/12345", nil)
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp2.StatusCode)

	var canvas models.Canvas
	json.NewDecoder(resp2.Body).Decode(&canvas)
	assert.Equal(t, "", canvas.ID)
}

func TestGetCanvasesByTeamID(t *testing.T) {
	app := setupCanvasApp()

	canvas1 := models.Canvas{
		Title:  "New Canvas",
		TeamID: "team123",
		Canvas: "Canvas content",
	}
	body, _ := json.Marshal(canvas1)
	req1 := httptest.NewRequest("POST", "/canvas", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")

	app.Test(req1)

	req := httptest.NewRequest("GET", "/canvases/team/team123", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var canvases []models.Canvas
	json.NewDecoder(resp.Body).Decode(&canvases)
	assert.Len(t, canvases, 1)
}

func TestUpdateCanvasTitle_Success(t *testing.T) {
	app := setupCanvasApp()

	canvas1 := models.Canvas{
		Title:  "New Canvas",
		TeamID: "team1",
		Canvas: "Canvas content",
	}
	body1, _ := json.Marshal(canvas1)
	req1 := httptest.NewRequest("POST", "/canvas", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")

	app.Test(req1)

	reqBody := map[string]string{"new_title": "Updated Title"}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/canvas/1/title", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "success", respBody["status"])
}

func TestUpdateCanvasTitle_InvalidJSON(t *testing.T) {
	app := setupCanvasApp()

	req := httptest.NewRequest("PUT", "/canvas/12345/title", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Invalid JSON", respBody["error"])
}

func TestDeleteCanvasByID_Success(t *testing.T) {
	app := setupCanvasApp()

	canvas1 := models.Canvas{
		Title:  "New Canvas",
		TeamID: "team1",
		Canvas: "Canvas content",
	}
	body1, _ := json.Marshal(canvas1)
	req1 := httptest.NewRequest("POST", "/canvas", bytes.NewReader(body1))
	req1.Header.Set("Content-Type", "application/json")

	app.Test(req1)

	req := httptest.NewRequest("DELETE", "/canvas/1", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "success", respBody["status"])
}

func TestDeleteCanvasByID_Failure(t *testing.T) {
	app := setupCanvasApp()

	req := httptest.NewRequest("DELETE", "/canvas/fail", nil)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)

	var respBody map[string]string
	json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "Failed to delete canvas", respBody["error"])
}
