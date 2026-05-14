package models

import "time"

// Note represents a user's napkin note.
type Note struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	FontID    *string    `json:"font_id,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
