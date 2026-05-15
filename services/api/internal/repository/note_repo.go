package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/notes"
)

// PostgresNoteRepo implements notes.NoteRepository using PostgreSQL.
type PostgresNoteRepo struct {
	db *sql.DB
}

// NewPostgresNoteRepo creates a new PostgresNoteRepo.
func NewPostgresNoteRepo(db *sql.DB) *PostgresNoteRepo {
	return &PostgresNoteRepo{db: db}
}

// Create inserts a new note into the database.
func (r *PostgresNoteRepo) Create(ctx context.Context, note *models.Note) error {
	note.ID = uuid.New().String()
	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now

	query := `
		INSERT INTO notes (id, user_id, content, font_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		note.ID, note.UserID, note.Content, note.FontID, note.CreatedAt, note.UpdatedAt,
	)
	return err
}

// GetByID retrieves a note by its ID.
func (r *PostgresNoteRepo) GetByID(ctx context.Context, id string) (*models.Note, error) {
	query := `
		SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		FROM notes WHERE id = $1`

	note := &models.Note{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&note.ID, &note.UserID, &note.Content, &note.FontID,
		&note.DeletedAt, &note.CreatedAt, &note.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, notes.ErrNoteNotFound
	}
	if err != nil {
		return nil, err
	}
	return note, nil
}

// List returns active (non-deleted) notes for a user with pagination.
func (r *PostgresNoteRepo) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	query := `
		SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		FROM notes
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotes(rows)
}

// Update updates an existing note's content and font.
func (r *PostgresNoteRepo) Update(ctx context.Context, note *models.Note) error {
	note.UpdatedAt = time.Now()

	query := `
		UPDATE notes SET content = $1, font_id = $2, updated_at = $3
		WHERE id = $4`

	result, err := r.db.ExecContext(ctx, query,
		note.Content, note.FontID, note.UpdatedAt, note.ID,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notes.ErrNoteNotFound
	}
	return nil
}

// SoftDelete marks a note as deleted by setting deleted_at.
func (r *PostgresNoteRepo) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE notes SET deleted_at = $1, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notes.ErrNoteNotFound
	}
	return nil
}

// Restore un-deletes a note by clearing deleted_at.
func (r *PostgresNoteRepo) Restore(ctx context.Context, id string) error {
	query := `UPDATE notes SET deleted_at = NULL, updated_at = $1 WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notes.ErrNoteNotFound
	}
	return nil
}

// PermanentDelete removes a note from the database permanently.
func (r *PostgresNoteRepo) PermanentDelete(ctx context.Context, id string) error {
	query := `DELETE FROM notes WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notes.ErrNoteNotFound
	}
	return nil
}

// ListTrashed returns soft-deleted notes for a user with pagination.
func (r *PostgresNoteRepo) ListTrashed(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	query := `
		SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		FROM notes
		WHERE user_id = $1 AND deleted_at IS NOT NULL
		ORDER BY deleted_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanNotes(rows)
}

// scanNotes reads note rows from a query result.
func scanNotes(rows *sql.Rows) ([]*models.Note, error) {
	var result []*models.Note
	for rows.Next() {
		note := &models.Note{}
		err := rows.Scan(
			&note.ID, &note.UserID, &note.Content, &note.FontID,
			&note.DeletedAt, &note.CreatedAt, &note.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		result = append(result, note)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		result = []*models.Note{}
	}
	return result, nil
}
