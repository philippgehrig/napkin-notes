package fonts

import (
	"context"
	"errors"
	"io"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/storage"
)

// ErrFontNotFound is returned when a font cannot be found.
var ErrFontNotFound = errors.New("font not found")

// ErrCannotDeleteDefault is returned when attempting to delete a default font.
var ErrCannotDeleteDefault = errors.New("cannot delete default font")

// FontRepository defines the data access interface for fonts.
type FontRepository interface {
	Create(ctx context.Context, font *models.Font) error
	GetByID(ctx context.Context, id string) (*models.Font, error)
	ListByUser(ctx context.Context, userID string) ([]*models.Font, error)
	UpdateStatus(ctx context.Context, id string, status models.FontStatus) error
	Delete(ctx context.Context, id string) error
}

// Service provides business logic for font operations.
type Service struct {
	repo    FontRepository
	storage storage.Storage
}

// NewService creates a new fonts Service.
func NewService(repo FontRepository, storage storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

// Create creates a new font record and saves the template scan to storage.
func (s *Service) Create(ctx context.Context, userID, name, scanPath string, reader io.Reader) (*models.Font, error) {
	// Save the template scan to storage
	if err := s.storage.Save(scanPath, reader); err != nil {
		return nil, err
	}

	font := &models.Font{
		UserID:           userID,
		Name:             name,
		Status:           models.FontStatusPending,
		TemplateScanPath: scanPath,
	}

	if err := s.repo.Create(ctx, font); err != nil {
		return nil, err
	}

	return font, nil
}

// GetByID retrieves a font by ID. Returns ErrFontNotFound if the font
// does not exist or does not belong to the user (unless it's a default font).
func (s *Service) GetByID(ctx context.Context, id, userID string) (*models.Font, error) {
	font, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if font.UserID != userID && !font.IsDefault {
		return nil, ErrFontNotFound
	}
	return font, nil
}

// List returns all fonts available to a user (user's own + defaults).
func (s *Service) List(ctx context.Context, userID string) ([]*models.Font, error) {
	return s.repo.ListByUser(ctx, userID)
}

// Delete removes a font. Cannot delete default fonts.
func (s *Service) Delete(ctx context.Context, id, userID string) error {
	font, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if font.UserID != userID {
		return ErrFontNotFound
	}
	if font.IsDefault {
		return ErrCannotDeleteDefault
	}

	// Delete the file from storage if it exists
	if font.TemplateScanPath != "" {
		_ = s.storage.Delete(font.TemplateScanPath)
	}
	if font.FilePath != "" {
		_ = s.storage.Delete(font.FilePath)
	}

	return s.repo.Delete(ctx, id)
}
