package fonts

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
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

// setupTestRouter creates a chi router with the fonts handler wired up.
func setupTestRouter(userID string) (*chi.Mux, *Service) {
	repo := newMockFontRepo()
	store := newMockStorage()
	svc := NewService(repo, store)
	handler := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(fakeAuthMiddleware(userID))
	r.Route("/api/fonts", func(r chi.Router) {
		r.Get("/", handler.List)
		r.Post("/", handler.Upload)
		r.Get("/{id}", handler.GetByID)
		r.Get("/{id}/file", handler.ServeFile)
		r.Delete("/{id}", handler.Delete)
	})

	return r, svc
}

func TestHandler_List_Empty(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/fonts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var fonts []*models.Font
	if err := json.NewDecoder(w.Body).Decode(&fonts); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(fonts) != 0 {
		t.Errorf("expected 0 fonts, got %d", len(fonts))
	}
}

func TestHandler_List_WithFonts(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	svc.Create(context.Background(), "user-1", "Font A", "scans/user-1/a.png", bytes.NewReader([]byte("data-a")))
	svc.Create(context.Background(), "user-1", "Font B", "scans/user-1/b.png", bytes.NewReader([]byte("data-b")))

	req := httptest.NewRequest(http.MethodGet, "/api/fonts", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var fonts []*models.Font
	json.NewDecoder(w.Body).Decode(&fonts)
	if len(fonts) != 2 {
		t.Errorf("expected 2 fonts, got %d", len(fonts))
	}
}

func TestHandler_GetByID_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodGet, "/api/fonts/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_GetByID_Found(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	font, _ := svc.Create(context.Background(), "user-1", "My Font", "scans/user-1/my.png", bytes.NewReader([]byte("data")))

	req := httptest.NewRequest(http.MethodGet, "/api/fonts/"+font.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got models.Font
	json.NewDecoder(w.Body).Decode(&got)
	if got.ID != font.ID {
		t.Errorf("expected ID=%s, got %s", font.ID, got.ID)
	}
}

func TestHandler_Upload(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add template file
	part, err := writer.CreateFormFile("template", "scan.png")
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	part.Write([]byte("fake-image-data"))

	// Add name field
	writer.WriteField("name", "My Handwriting")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/fonts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var font models.Font
	json.NewDecoder(w.Body).Decode(&font)
	if font.ID == "" {
		t.Error("expected font to have an ID")
	}
	if font.Name != "My Handwriting" {
		t.Errorf("expected name='My Handwriting', got %s", font.Name)
	}
	if font.Status != models.FontStatusPending {
		t.Errorf("expected status=pending, got %s", font.Status)
	}
}

func TestHandler_Upload_MissingFile(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	writer.WriteField("name", "No File")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/fonts", &buf)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete(t *testing.T) {
	router, svc := setupTestRouter("user-1")

	font, _ := svc.Create(context.Background(), "user-1", "Delete Me", "scans/user-1/del.png", bytes.NewReader([]byte("data")))

	req := httptest.NewRequest(http.MethodDelete, "/api/fonts/"+font.ID, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Delete_NotFound(t *testing.T) {
	router, _ := setupTestRouter("user-1")

	req := httptest.NewRequest(http.MethodDelete, "/api/fonts/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}
