package controllers

import (
	"fmt"

	"encoding/base64"
	"go-server/models"
	"go-server/repository"

	"github.com/gofiber/fiber/v2"
)

type NoteController struct {
	repo *repository.NoteRepository
}

func NewNoteController(repo *repository.NoteRepository) *NoteController {
	return &NoteController{repo: repo}
}

func (nc *NoteController) CreateNote(c *fiber.Ctx) error {
	var note models.Note
	if err := c.BodyParser(&note); err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid JSON"})
	}

	if err := nc.repo.SaveNote(note); err != nil {
		fmt.Print(err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to save note"})
	}

	return c.Status(fiber.StatusCreated).JSON(note)
}

func decodeNoteBase64(encodedNote string) ([]byte, error) {
	decodedData, err := base64.StdEncoding.DecodeString(encodedNote)
	if err != nil {
		return nil, err
	}
	return decodedData, nil
}

func (nc *NoteController) GetNoteByTeamID(c *fiber.Ctx) error {
	teamID := c.Params("teamId")
	note, err := nc.repo.FindNoteByTeamID(teamID)
	if err != nil {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"note": nil})
	}

	return c.Status(fiber.StatusOK).JSON(note)
}
