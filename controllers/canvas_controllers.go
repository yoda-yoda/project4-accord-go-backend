package controllers

import (
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/fiber/v2"
)

type CanvasController struct {
	repo repository.CanvasRepositoryInterface
}

func NewCanvasController(repo repository.CanvasRepositoryInterface) *CanvasController {
	return &CanvasController{repo: repo}
}

func (cc *CanvasController) CreateCanvas(c *fiber.Ctx) error {
	var canvas models.Canvas
	if err := c.BodyParser(&canvas); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	objectID, err := cc.repo.SaveCanvas(canvas)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save canvas"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": objectID})
}

func (cc *CanvasController) GetCanvasByID(c *fiber.Ctx) error {
	id := c.Params("id")
	canvas, err := cc.repo.FindCanvasByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Canvas not found"})
	}
	return c.Status(fiber.StatusOK).JSON(canvas)
}

func (cc *CanvasController) GetCanvasesByTeamID(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	canvases, err := cc.repo.FindCanvasesByTeamID(teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find canvases"})
	}
	return c.Status(fiber.StatusOK).JSON(canvases)
}

func (cc *CanvasController) UpdateCanvasTitle(c *fiber.Ctx) error {
	id := c.Params("id")
	var request struct {
		NewTitle string `json:"new_title"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := cc.repo.UpdateCanvasTitle(id, request.NewTitle); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update canvas title"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (cc *CanvasController) DeleteCanvasByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := cc.repo.DeleteCanvasByID(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete canvas"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
