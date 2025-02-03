package controllers

import (
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/fiber/v2"
)

type NoteController struct {
	repo repository.NoteRepositoryInterface
}

func NewNoteController(repo repository.NoteRepositoryInterface) *NoteController {
	return &NoteController{repo: repo}
}

func (nc *NoteController) CreateNote(c *fiber.Ctx) error {
	var note models.Note
	if err := c.BodyParser(&note); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	objectID, err := nc.repo.SaveNote(note)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save note"})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"id": objectID})
}

func (nc *NoteController) GetNoteByID(c *fiber.Ctx) error {
	id := c.Params("id")
	note, err := nc.repo.FindNoteByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Note not found"})
	}
	return c.Status(fiber.StatusOK).JSON(note)
}

func (nc *NoteController) GetNotesByTeamID(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	notes, err := nc.repo.FindNotesByTeamID(teamID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to find notes"})
	}
	return c.Status(fiber.StatusOK).JSON(notes)
}

func (nc *NoteController) UpdateNoteTitle(c *fiber.Ctx) error {
	id := c.Params("id")
	var request struct {
		NewTitle string `json:"new_title"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := nc.repo.UpdateNoteTitle(id, request.NewTitle); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update note title"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func (nc *NoteController) DeleteNoteByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if err := nc.repo.DeleteNoteByID(id); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to delete note"})
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}
