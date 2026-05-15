package export

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// mockNoteGetter is a test implementation of NoteGetter.
type mockNoteGetter struct {
	notes map[string]*models.Note
}

func newMockNoteGetter() *mockNoteGetter {
	return &mockNoteGetter{notes: make(map[string]*models.Note)}
}

func (m *mockNoteGetter) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	note, ok := m.notes[id]
	if !ok {
		return nil, errors.New("note not found")
	}
	if note.UserID != userID {
		return nil, errors.New("note not found")
	}
	return note, nil
}

// fakeAuthMiddleware injects a userID into the context for testing.
func fakeAuthMiddleware(userID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKeyForTest(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func setupExportRouter(userID string, getter *mockNoteGetter) *chi.Mux {
	handler := NewHandler(getter)

	r := chi.NewRouter()
	r.Use(fakeAuthMiddleware(userID))
	r.Get("/api/notes/{id}/export", handler.Export)

	return r
}

func TestHandler_Export_Success(t *testing.T) {
	getter := newMockNoteGetter()
	getter.notes["note-1"] = &models.Note{
		ID:      "note-1",
		UserID:  "user-1",
		Content: "Hello export",
	}

	router := setupExportRouter("user-1", getter)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/note-1/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	ct := w.Header().Get("Content-Type")
	if ct != "image/png" {
		t.Errorf("expected Content-Type=image/png, got %s", ct)
	}

	cd := w.Header().Get("Content-Disposition")
	if cd != `attachment; filename="napkin.png"` {
		t.Errorf("expected Content-Disposition attachment, got %s", cd)
	}

	// Verify it's a valid PNG (starts with PNG magic bytes)
	body := w.Body.Bytes()
	if len(body) < 8 {
		t.Fatal("response body too short to be a PNG")
	}
	pngMagic := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	for i, b := range pngMagic {
		if body[i] != b {
			t.Fatalf("invalid PNG magic byte at position %d: expected %02X, got %02X", i, b, body[i])
		}
	}
}

func TestHandler_Export_NotFound(t *testing.T) {
	getter := newMockNoteGetter()

	router := setupExportRouter("user-1", getter)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/nonexistent/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Export_WrongUser(t *testing.T) {
	getter := newMockNoteGetter()
	getter.notes["note-1"] = &models.Note{
		ID:      "note-1",
		UserID:  "user-1",
		Content: "Secret note",
	}

	router := setupExportRouter("user-2", getter)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/note-1/export", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
