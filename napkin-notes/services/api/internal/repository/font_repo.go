package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/fonts"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// PostgresFontRepo implements fonts.FontRepository using PostgreSQL.
type PostgresFontRepo struct {
	db *sql.DB
}

// NewPostgresFontRepo creates a new PostgresFontRepo.
func NewPostgresFontRepo(db *sql.DB) *PostgresFontRepo {
	return &PostgresFontRepo{db: db}
}

// Create inserts a new font into the database.
func (r *PostgresFontRepo) Create(ctx context.Context, font *models.Font) error {
	font.ID = uuid.New().String()
	now := time.Now()
	font.CreatedAt = now
	font.UpdatedAt = now

	query := `
		INSERT INTO fonts (id, user_id, name, file_path, status, template_scan_path, is_default, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		font.ID, font.UserID, font.Name, font.FilePath, font.Status,
		font.TemplateScanPath, font.IsDefault, font.CreatedAt, font.UpdatedAt,
	)
	return err
}

// GetByID retrieves a font by its ID.
func (r *PostgresFontRepo) GetByID(ctx context.Context, id string) (*models.Font, error) {
	query := `
		SELECT id, user_id, name, file_path, status, template_scan_path, is_default, created_at, updated_at
		FROM fonts WHERE id = $1`

	font := &models.Font{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&font.ID, &font.UserID, &font.Name, &font.FilePath, &font.Status,
		&font.TemplateScanPath, &font.IsDefault, &font.CreatedAt, &font.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fonts.ErrFontNotFound
	}
	if err != nil {
		return nil, err
	}
	return font, nil
}

// ListByUser returns all fonts belonging to a user plus all default fonts.
func (r *PostgresFontRepo) ListByUser(ctx context.Context, userID string) ([]*models.Font, error) {
	query := `
		SELECT id, user_id, name, file_path, status, template_scan_path, is_default, created_at, updated_at
		FROM fonts
		WHERE user_id = $1 OR is_default = true
		ORDER BY is_default DESC, created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Font
	for rows.Next() {
		font := &models.Font{}
		err := rows.Scan(
			&font.ID, &font.UserID, &font.Name, &font.FilePath, &font.Status,
			&font.TemplateScanPath, &font.IsDefault, &font.CreatedAt, &font.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, font)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []*models.Font{}
	}
	return result, nil
}

// UpdateStatus updates the status of a font.
func (r *PostgresFontRepo) UpdateStatus(ctx context.Context, id string, status models.FontStatus) error {
	query := `UPDATE fonts SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fonts.ErrFontNotFound
	}
	return nil
}

// Delete removes a font from the database. Prevents deleting default fonts.
func (r *PostgresFontRepo) Delete(ctx context.Context, id string) error {
	// First check if it's a default font
	var isDefault bool
	checkQuery := `SELECT is_default FROM fonts WHERE id = $1`
	err := r.db.QueryRowContext(ctx, checkQuery, id).Scan(&isDefault)
	if err == sql.ErrNoRows {
		return fonts.ErrFontNotFound
	}
	if err != nil {
		return err
	}
	if isDefault {
		return fonts.ErrCannotDeleteDefault
	}

	query := `DELETE FROM fonts WHERE id = $1 AND is_default = false`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fonts.ErrFontNotFound
	}
	return nil
}
