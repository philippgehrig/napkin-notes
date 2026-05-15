package export

import (
	"context"
	"image/png"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// NoteGetter defines the interface for fetching notes.
type NoteGetter interface {
	GetByID(ctx context.Context, id, userID string) (*models.Note, error)
}

// Handler handles image export HTTP requests.
type Handler struct {
	notes NoteGetter
}

// NewHandler creates a new export Handler.
func NewHandler(notes NoteGetter) *Handler {
	return &Handler{notes: notes}
}

// Export renders a note as a PNG image and returns it as an attachment.
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.notes.GetByID(r.Context(), noteID, userID)
	if err != nil {
		http.Error(w, `{"error":"note not found"}`, http.StatusNotFound)
		return
	}

	opts := RenderOptions{
		Content:  note.Content,
		Width:    800,
		Height:   600,
		FontSize: 28,
	}

	img, err := RenderNapkin(opts)
	if err != nil {
		http.Error(w, `{"error":"failed to render image"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Content-Disposition", `attachment; filename="napkin.png"`)

	if err := png.Encode(w, img); err != nil {
		http.Error(w, `{"error":"failed to encode image"}`, http.StatusInternalServerError)
		return
	}
}
