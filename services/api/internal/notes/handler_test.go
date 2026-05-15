package notes

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// fakeAuthMiddleware injects a userID into the context for testing.
func fakeAuthMiddleware(userID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKeyForTest(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// setupTestRouter creates a chi router with the notes handler wired up.
func setupTestRouter(userID string) (*chi.Mux, *Service) {
	repo := newMockNoteRepo()
	svc := NewService(repo)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(fakeAuthMiddleware(userID))
	r.Route("/api/notes", func(r chi.Router) {
		r.Get("/", handler.List)
		r.Post("/", handler.Create)
		r.Get("/trash", handler.ListTrashed)
		r.Get("/{id}", handler.GetByID)
		r.Put("/{id}", handler.Update)
		r.Delete("/{id}", handler.Delete)
		r.Post("/{id}/restore", handler.Restore)
		r.Delete("/{id}/permanent", handler.PermanentDelete)
	})

	return r, svc
}

func TestHandler_Create(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	body := `{"content":"Hello handler"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var note models.Note
	if err := json.NewDecoder(w.Body).Decode(&note); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if note.ID == "" {
		t.Error("expected note to have an ID")
	}
	if note.Content != "Hello handler" {
		t.Errorf("expected content='Hello handler', got %s", note.Content)
	}
	if note.UserID != "user-1" {
		t.Errorf("expected user_id='user-1', got %s", note.UserID)
	}
}

func TestHandler_GetByID(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	// Create a note first
	note, _ := svc.Create(context.Background(), "user-1", "Get me", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/notes/"+note.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got models.Note
	json.NewDecoder(w.Body).Decode(&got)
	if got.ID != note.ID {
		t.Errorf("expected ID=%s, got %s", note.ID, got.ID)
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/notes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_List(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	svc.Create(context.Background(), "user-1", "Note 1", nil)
	svc.Create(context.Background(), "user-1", "Note 2", nil)

	req := httptest.NewRequest(http.MethodGet, "/api/notes?limit=10&offset=0", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var notes []*models.Note
	json.NewDecoder(w.Body).Decode(&notes)
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
}

func TestHandler_Update(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	note, _ := svc.Create(context.Background(), "user-1", "Original", nil)

	body := `{"content":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/notes/"+note.ID, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Note
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Content != "Updated" {
		t.Errorf("expected content='Updated', got %s", updated.Content)
	}
}

func TestHandler_Update_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	body := `{"content":"Updated"}`
	req := httptest.NewRequest(http.MethodPut, "/api/notes/nonexistent", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_Delete(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	note, _ := svc.Create(context.Background(), "user-1", "Delete me", nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/notes/"+note.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodDelete, "/api/notes/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_ListTrashed(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	note, _ := svc.Create(context.Background(), "user-1", "Trashed note", nil)
	svc.SoftDelete(context.Background(), note.ID, "user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/notes/trash", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var notes []*models.Note
	json.NewDecoder(w.Body).Decode(&notes)
	if len(notes) != 1 {
		t.Errorf("expected 1 trashed note, got %d", len(notes))
	}
}

func TestHandler_Restore(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	note, _ := svc.Create(context.Background(), "user-1", "Restore me", nil)
	svc.SoftDelete(context.Background(), note.ID, "user-1")

	req := httptest.NewRequest(http.MethodPost, "/api/notes/"+note.ID+"/restore", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var restored models.Note
	json.NewDecoder(w.Body).Decode(&restored)
	if restored.DeletedAt != nil {
		t.Error("restored note should have nil deleted_at")
	}
}

func TestHandler_Restore_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodPost, "/api/notes/nonexistent/restore", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandler_PermanentDelete(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	note, _ := svc.Create(context.Background(), "user-1", "Delete forever", nil)
	svc.SoftDelete(context.Background(), note.ID, "user-1")

	req := httptest.NewRequest(http.MethodDelete, "/api/notes/"+note.ID+"/permanent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_PermanentDelete_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodDelete, "/api/notes/nonexistent/permanent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// handlerMockNoteRepo is a duplicate-free reference to the same mock used in service tests.
// We reuse the one defined in service_test.go since both are in the same package.
var _ NoteRepository = (*mockNoteRepo)(nil)

// Ensure mockNoteRepo satisfies the interface (compile-time check above).
// The mock is already defined in service_test.go so we don't redefine it here.
