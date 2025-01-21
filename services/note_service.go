package service

import (
	"go-server/repository"
)

// NoteService: CRDT + MongoDB 연동
type NoteService struct {
	noteRepo *repository.NoteRepository
}

func NewNoteService(noteRepo *repository.NoteRepository) *NoteService {
	return &NoteService{noteRepo: noteRepo}
}
