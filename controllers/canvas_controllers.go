package controllers

import (
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/fiber/v2"
)

type CanvasController struct {
	repo *repository.CanvasRepository
}

func NewCanvasController(repo *repository.CanvasRepository) *CanvasController {
	return &CanvasController{repo: repo}
}

func (cc *CanvasController) CreateCanvas(c *fiber.Ctx) error {
	var canvas models.Canvas
	if err := c.BodyParser(&canvas); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := cc.repo.SaveCanvas(canvas); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save canvas"})
	}

	return c.Status(fiber.StatusCreated).JSON(canvas)
}

func (cc *CanvasController) GetCanvasByTeamID(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	canvas, err := cc.repo.FindCanvasByTeamID(teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find canvas"})
	}
	return c.Status(fiber.StatusOK).JSON(canvas)
}
