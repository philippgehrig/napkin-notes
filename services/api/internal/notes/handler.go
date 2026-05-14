package notes

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
)

// Handler provides HTTP handlers for note endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new notes Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type createRequest struct {
	Content string  `json:"content"`
	FontID  *string `json:"font_id,omitempty"`
}

type updateRequest struct {
	Content string  `json:"content"`
	FontID  *string `json:"font_id,omitempty"`
}

// Create handles POST /api/notes.
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	note, err := h.svc.Create(r.Context(), userID, req.Content, req.FontID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusCreated, note)
}

// GetByID handles GET /api/notes/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.svc.GetByID(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, note)
}

// List handles GET /api/notes.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	notes, err := h.svc.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

// Update handles PUT /api/notes/{id}.
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}

	note, err := h.svc.Update(r.Context(), noteID, userID, req.Content, req.FontID)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, note)
}

// Delete handles DELETE /api/notes/{id} (soft delete).
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	err := h.svc.SoftDelete(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListTrashed handles GET /api/notes/trash.
func (h *Handler) ListTrashed(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	limit, offset := parsePagination(r)

	notes, err := h.svc.ListTrashed(r.Context(), userID, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, notes)
}

// Restore handles POST /api/notes/{id}/restore.
func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.svc.Restore(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, ErrNoteNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "note not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, note)
}

// parsePagination extracts limit and offset query params.
func parsePagination(r *http.Request) (int, int) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
