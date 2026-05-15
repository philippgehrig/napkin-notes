package models

import "time"

// User represents a registered user of the application.
type User struct {
	ID           string
	Email        string
	PasswordHash string
	DisplayName  string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
