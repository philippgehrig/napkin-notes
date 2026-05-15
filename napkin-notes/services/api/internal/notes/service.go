package notes

import (
	"context"
	"errors"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

// ErrNoteNotFound is returned when a note cannot be found.
var ErrNoteNotFound = errors.New("note not found")

// NoteRepository defines the data access interface for notes.
type NoteRepository interface {
	Create(ctx context.Context, note *models.Note) error
	GetByID(ctx context.Context, id string) (*models.Note, error)
	List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error)
	Update(ctx context.Context, note *models.Note) error
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
	ListTrashed(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error)
	PermanentDelete(ctx context.Context, id string) error
}

// Service provides business logic for note operations.
type Service struct {
	repo NoteRepository
}

// NewService creates a new notes Service.
func NewService(repo NoteRepository) *Service {
	return &Service{repo: repo}
}

// Create creates a new note for the given user.
func (s *Service) Create(ctx context.Context, userID, content string, fontID *string, textureVariant int) (*models.Note, error) {
	if textureVariant < 1 || textureVariant > 3 {
		textureVariant = 1
	}
	note := &models.Note{
		UserID:         userID,
		Content:        content,
		FontID:         fontID,
		TextureVariant: textureVariant,
	}
	if err := s.repo.Create(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

// GetByID retrieves a note by ID, scoped to the given user.
func (s *Service) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// List returns active (non-deleted) notes for a user with pagination.
func (s *Service) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	limit = capLimit(limit)
	return s.repo.List(ctx, userID, limit, offset)
}

// Update updates a note's content and/or font, scoped to the given user.
func (s *Service) Update(ctx context.Context, id, userID, content string, fontID *string, textureVariant int) (*models.Note, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	note.Content = content
	note.FontID = fontID
	if textureVariant >= 1 && textureVariant <= 3 {
		note.TextureVariant = textureVariant
	}
	if err := s.repo.Update(ctx, note); err != nil {
		return nil, err
	}
	return note, nil
}

// SoftDelete marks a note as deleted, scoped to the given user.
func (s *Service) SoftDelete(ctx context.Context, id, userID string) error {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if note.UserID != userID {
		return ErrNoteNotFound
	}
	return s.repo.SoftDelete(ctx, id)
}

// Restore un-deletes a note, scoped to the given user.
func (s *Service) Restore(ctx context.Context, id, userID string) (*models.Note, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	if err := s.repo.Restore(ctx, id); err != nil {
		return nil, err
	}
	// Re-fetch to get updated timestamps
	return s.repo.GetByID(ctx, id)
}

// ListTrashed returns soft-deleted notes for a user with pagination.
func (s *Service) ListTrashed(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	limit = capLimit(limit)
	return s.repo.ListTrashed(ctx, userID, limit, offset)
}

// PermanentDelete permanently removes a note from the database, scoped to the given user.
func (s *Service) PermanentDelete(ctx context.Context, id, userID string) error {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if note.UserID != userID {
		return ErrNoteNotFound
	}
	return s.repo.PermanentDelete(ctx, id)
}

// capLimit enforces default and maximum limits.
func capLimit(limit int) int {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
