package fonts

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
)

const maxUploadSize = 10 << 20 // 10MB

// Handler provides HTTP handlers for font endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a new fonts Handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// List handles GET /api/fonts.
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	fonts, err := h.svc.List(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, fonts)
}

// GetByID handles GET /api/fonts/{id}.
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	fontID := chi.URLParam(r, "id")

	font, err := h.svc.GetByID(r.Context(), fontID, userID)
	if err != nil {
		if errors.Is(err, ErrFontNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "font not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, font)
}

// Upload handles POST /api/fonts.
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file too large or invalid form"})
		return
	}

	file, header, err := r.FormFile("template")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing template file"})
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	if name == "" {
		name = header.Filename
	}

	// Generate a storage path for the scan
	ext := filepath.Ext(header.Filename)
	scanPath := fmt.Sprintf("scans/%s/%s%s", userID, name, ext)

	font, err := h.svc.Create(r.Context(), userID, name, scanPath, file)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to upload font"})
		return
	}

	writeJSON(w, http.StatusCreated, font)
}

// ServeFile handles GET /api/fonts/{id}/file.
func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	fontID := chi.URLParam(r, "id")

	font, err := h.svc.GetByID(r.Context(), fontID, userID)
	if err != nil {
		if errors.Is(err, ErrFontNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "font not found"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	if font.FilePath == "" {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "font file not available"})
		return
	}

	reader, err := h.svc.storage.Get(font.FilePath)
	if err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "font file not found"})
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.WriteHeader(http.StatusOK)

	buf := make([]byte, 32*1024)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if readErr != nil {
			break
		}
	}
}

// Delete handles DELETE /api/fonts/{id}.
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	fontID := chi.URLParam(r, "id")

	err := h.svc.Delete(r.Context(), fontID, userID)
	if err != nil {
		if errors.Is(err, ErrFontNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "font not found"})
			return
		}
		if errors.Is(err, ErrCannotDeleteDefault) {
			writeJSON(w, http.StatusForbidden, map[string]string{"error": "cannot delete default font"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// writeJSON writes a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
