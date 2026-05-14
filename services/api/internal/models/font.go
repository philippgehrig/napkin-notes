package models

import "time"

// FontStatus represents the processing state of a font.
type FontStatus string

const (
	FontStatusPending    FontStatus = "pending"
	FontStatusProcessing FontStatus = "processing"
	FontStatusReady      FontStatus = "ready"
	FontStatusFailed     FontStatus = "failed"
)

// Font represents a user's handwriting font.
type Font struct {
	ID               string     `json:"id"`
	UserID           string     `json:"user_id"`
	Name             string     `json:"name"`
	FilePath         string     `json:"file_path,omitempty"`
	Status           FontStatus `json:"status"`
	TemplateScanPath string     `json:"template_scan_path,omitempty"`
	IsDefault        bool       `json:"is_default"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
