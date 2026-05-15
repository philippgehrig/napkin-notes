# Phase 5: Fonts API

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Implement font upload, metadata, file serving, and status tracking.

**Branch:** `feat/phase-05-fonts-api`

---

## File Structure

```
services/api/internal/
├── models/
│   └── font.go
├── fonts/
│   ├── handler.go
│   ├── handler_test.go
│   ├── service.go
│   └── service_test.go
├── repository/
│   └── font_repo.go
└── storage/
    ├── storage.go        (interface + local implementation)
    └── storage_test.go
```

---

### Task 1: Font model and storage interface

**Files:**
- Create: `services/api/internal/models/font.go`
- Create: `services/api/internal/storage/storage.go`
- Create: `services/api/internal/storage/storage_test.go`

- [ ] **Step 1: Create font model**

Create `services/api/internal/models/font.go`:
```go
package models

import "time"

type FontStatus string

const (
	FontStatusPending    FontStatus = "pending"
	FontStatusProcessing FontStatus = "processing"
	FontStatusReady      FontStatus = "ready"
	FontStatusFailed     FontStatus = "failed"
)

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
```

- [ ] **Step 2: Create storage interface and local implementation**

Create `services/api/internal/storage/storage.go`:
```go
package storage

import (
	"io"
	"os"
	"path/filepath"
)

type Storage interface {
	Save(path string, reader io.Reader) error
	Get(path string) (io.ReadCloser, error)
	Delete(path string) error
}

type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Save(path string, reader io.Reader) error {
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func (s *LocalStorage) Get(path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	return os.Open(fullPath)
}

func (s *LocalStorage) Delete(path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}
```

- [ ] **Step 3: Write storage test**

Create `services/api/internal/storage/storage_test.go`:
```go
package storage

import (
	"io"
	"os"
	"strings"
	"testing"
)

func TestLocalStorage_SaveAndGet(t *testing.T) {
	dir, _ := os.MkdirTemp("", "storage-test")
	defer os.RemoveAll(dir)

	s := NewLocalStorage(dir)

	err := s.Save("fonts/test.woff2", strings.NewReader("fake font data"))
	if err != nil {
		t.Fatalf("save failed: %v", err)
	}

	reader, err := s.Get("fonts/test.woff2")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	defer reader.Close()

	data, _ := io.ReadAll(reader)
	if string(data) != "fake font data" {
		t.Errorf("expected 'fake font data', got %q", string(data))
	}
}

func TestLocalStorage_Delete(t *testing.T) {
	dir, _ := os.MkdirTemp("", "storage-test")
	defer os.RemoveAll(dir)

	s := NewLocalStorage(dir)
	_ = s.Save("fonts/test.woff2", strings.NewReader("data"))

	err := s.Delete("fonts/test.woff2")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err = s.Get("fonts/test.woff2")
	if err == nil {
		t.Error("expected error after delete")
	}
}
```

- [ ] **Step 4: Run tests**

```bash
cd services/api && go test ./internal/storage/ -v
```
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/models/font.go services/api/internal/storage/
git commit -m "feat: add Font model and local file storage"
```

---

### Task 2: Fonts service

**Files:**
- Create: `services/api/internal/fonts/service.go`
- Create: `services/api/internal/fonts/service_test.go`

- [ ] **Step 1: Write fonts service tests**

Create `services/api/internal/fonts/service_test.go`:
```go
package fonts

import (
	"context"
	"testing"
	"time"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type mockFontRepo struct {
	fonts map[string]*models.Font
}

func newMockFontRepo() *mockFontRepo {
	return &mockFontRepo{fonts: make(map[string]*models.Font)}
}

func (m *mockFontRepo) Create(ctx context.Context, userID, name, templatePath string) (*models.Font, error) {
	font := &models.Font{
		ID:               "font-1",
		UserID:           userID,
		Name:             name,
		Status:           models.FontStatusPending,
		TemplateScanPath: templatePath,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	m.fonts[font.ID] = font
	return font, nil
}

func (m *mockFontRepo) GetByID(ctx context.Context, id string) (*models.Font, error) {
	font, ok := m.fonts[id]
	if !ok {
		return nil, ErrFontNotFound
	}
	return font, nil
}

func (m *mockFontRepo) ListByUser(ctx context.Context, userID string) ([]*models.Font, error) {
	var result []*models.Font
	for _, f := range m.fonts {
		if f.UserID == userID || f.IsDefault {
			result = append(result, f)
		}
	}
	return result, nil
}

func (m *mockFontRepo) UpdateStatus(ctx context.Context, id string, status models.FontStatus, filePath string) error {
	font, ok := m.fonts[id]
	if !ok {
		return ErrFontNotFound
	}
	font.Status = status
	font.FilePath = filePath
	return nil
}

func (m *mockFontRepo) Delete(ctx context.Context, id, userID string) error {
	font, ok := m.fonts[id]
	if !ok || font.UserID != userID {
		return ErrFontNotFound
	}
	delete(m.fonts, id)
	return nil
}

func TestCreateFont(t *testing.T) {
	repo := newMockFontRepo()
	svc := NewService(repo, nil)

	font, err := svc.Create(context.Background(), "user-1", "My Handwriting", "scans/template.png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if font.Name != "My Handwriting" {
		t.Errorf("expected 'My Handwriting', got %q", font.Name)
	}
	if font.Status != models.FontStatusPending {
		t.Errorf("expected pending status, got %s", font.Status)
	}
}

func TestListFonts(t *testing.T) {
	repo := newMockFontRepo()
	svc := NewService(repo, nil)

	_, _ = svc.Create(context.Background(), "user-1", "Font 1", "scan1.png")

	fonts, err := svc.ListByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fonts) != 1 {
		t.Errorf("expected 1 font, got %d", len(fonts))
	}
}

func TestDeleteFont(t *testing.T) {
	repo := newMockFontRepo()
	svc := NewService(repo, nil)

	_, _ = svc.Create(context.Background(), "user-1", "Font 1", "scan1.png")

	err := svc.Delete(context.Background(), "font-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	fonts, _ := svc.ListByUser(context.Background(), "user-1")
	if len(fonts) != 0 {
		t.Errorf("expected 0 fonts, got %d", len(fonts))
	}
}
```

- [ ] **Step 2: Implement fonts service**

Create `services/api/internal/fonts/service.go`:
```go
package fonts

import (
	"context"
	"errors"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/storage"
)

var ErrFontNotFound = errors.New("font not found")

type FontRepository interface {
	Create(ctx context.Context, userID, name, templatePath string) (*models.Font, error)
	GetByID(ctx context.Context, id string) (*models.Font, error)
	ListByUser(ctx context.Context, userID string) ([]*models.Font, error)
	UpdateStatus(ctx context.Context, id string, status models.FontStatus, filePath string) error
	Delete(ctx context.Context, id, userID string) error
}

type Service struct {
	repo    FontRepository
	storage storage.Storage
}

func NewService(repo FontRepository, storage storage.Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) Create(ctx context.Context, userID, name, templatePath string) (*models.Font, error) {
	return s.repo.Create(ctx, userID, name, templatePath)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Font, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByUser(ctx context.Context, userID string) ([]*models.Font, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *Service) Delete(ctx context.Context, id, userID string) error {
	return s.repo.Delete(ctx, id, userID)
}

func (s *Service) GetStorage() storage.Storage {
	return s.storage
}
```

- [ ] **Step 3: Run tests**

```bash
cd services/api && go test ./internal/fonts/ -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/fonts/
git commit -m "feat: add fonts service with CRUD"
```

---

### Task 3: Fonts HTTP handlers

**Files:**
- Create: `services/api/internal/fonts/handler.go`
- Create: `services/api/internal/fonts/handler_test.go`

- [ ] **Step 1: Write handler tests**

Create `services/api/internal/fonts/handler_test.go`:
```go
package fonts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
)

func setupTestRouter() *chi.Mux {
	repo := newMockFontRepo()
	svc := NewService(repo, nil)
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, "user-1")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Get("/api/fonts", h.List)
	r.Get("/api/fonts/{id}", h.GetByID)
	r.Delete("/api/fonts/{id}", h.Delete)

	return r
}

func TestHandleListFonts(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/fonts", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var fonts []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&fonts)
	// Empty list is fine for initial state
}

func TestHandleGetFont_NotFound(t *testing.T) {
	r := setupTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/fonts/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Implement handler**

Create `services/api/internal/fonts/handler.go`:
```go
package fonts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	fonts, err := h.service.ListByUser(r.Context(), userID)
	if err != nil {
		writeError(w, "failed to list fonts", http.StatusInternalServerError)
		return
	}
	if fonts == nil {
		fonts = []*models.Font{}
	}
	writeJSON(w, fonts, http.StatusOK)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	fontID := chi.URLParam(r, "id")

	font, err := h.service.GetByID(r.Context(), fontID)
	if err != nil {
		writeError(w, "font not found", http.StatusNotFound)
		return
	}

	writeJSON(w, font, http.StatusOK)
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, "file too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("template")
	if err != nil {
		writeError(w, "template file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	name := r.FormValue("name")
	if name == "" {
		name = "My Handwriting"
	}

	ext := filepath.Ext(header.Filename)
	scanPath := fmt.Sprintf("scans/%s/%s%s", userID, uuid.New().String(), ext)

	if h.service.GetStorage() != nil {
		if err := h.service.GetStorage().Save(scanPath, file); err != nil {
			writeError(w, "failed to save file", http.StatusInternalServerError)
			return
		}
	}

	font, err := h.service.Create(r.Context(), userID, name, scanPath)
	if err != nil {
		writeError(w, "failed to create font", http.StatusInternalServerError)
		return
	}

	writeJSON(w, font, http.StatusCreated)
}

func (h *Handler) ServeFile(w http.ResponseWriter, r *http.Request) {
	fontID := chi.URLParam(r, "id")

	font, err := h.service.GetByID(r.Context(), fontID)
	if err != nil || font.FilePath == "" {
		writeError(w, "font file not available", http.StatusNotFound)
		return
	}

	if h.service.GetStorage() == nil {
		writeError(w, "storage not configured", http.StatusInternalServerError)
		return
	}

	reader, err := h.service.GetStorage().Get(font.FilePath)
	if err != nil {
		writeError(w, "font file not found", http.StatusNotFound)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", "font/woff2")
	w.Header().Set("Cache-Control", "public, max-age=31536000")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	http.ServeContent(w, r, font.FilePath, font.UpdatedAt, reader.(readSeeker))
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	fontID := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), fontID, userID); err != nil {
		writeError(w, "font not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type readSeeker interface {
	Read(p []byte) (n int, err error)
	Seek(offset int64, whence int) (int64, error)
}

func writeJSON(w http.ResponseWriter, data interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
```

- [ ] **Step 3: Run tests**

```bash
cd services/api && go test ./internal/fonts/ -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/fonts/
git commit -m "feat: add fonts HTTP handlers with upload and file serving"
```

---

### Task 4: PostgreSQL font repository and wire routes

**Files:**
- Create: `services/api/internal/repository/font_repo.go`
- Modify: `services/api/main.go`

- [ ] **Step 1: Implement font repository**

Create `services/api/internal/repository/font_repo.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/fonts"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type PostgresFontRepo struct {
	db *sql.DB
}

func NewPostgresFontRepo(db *sql.DB) *PostgresFontRepo {
	return &PostgresFontRepo{db: db}
}

func (r *PostgresFontRepo) Create(ctx context.Context, userID, name, templatePath string) (*models.Font, error) {
	font := &models.Font{
		ID:               uuid.New().String(),
		UserID:           userID,
		Name:             name,
		Status:           models.FontStatusPending,
		TemplateScanPath: templatePath,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO fonts (id, user_id, name, status, template_scan_path, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		font.ID, font.UserID, font.Name, font.Status, font.TemplateScanPath, font.CreatedAt, font.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return font, nil
}

func (r *PostgresFontRepo) GetByID(ctx context.Context, id string) (*models.Font, error) {
	font := &models.Font{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, name, file_path, status, template_scan_path, is_default, created_at, updated_at
		 FROM fonts WHERE id = $1`, id,
	).Scan(&font.ID, &font.UserID, &font.Name, &font.FilePath, &font.Status, &font.TemplateScanPath, &font.IsDefault, &font.CreatedAt, &font.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, fonts.ErrFontNotFound
	}
	return font, err
}

func (r *PostgresFontRepo) ListByUser(ctx context.Context, userID string) ([]*models.Font, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, name, file_path, status, template_scan_path, is_default, created_at, updated_at
		 FROM fonts WHERE user_id = $1 OR is_default = TRUE
		 ORDER BY is_default DESC, created_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Font
	for rows.Next() {
		font := &models.Font{}
		if err := rows.Scan(&font.ID, &font.UserID, &font.Name, &font.FilePath, &font.Status, &font.TemplateScanPath, &font.IsDefault, &font.CreatedAt, &font.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, font)
	}
	return result, nil
}

func (r *PostgresFontRepo) UpdateStatus(ctx context.Context, id string, status models.FontStatus, filePath string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE fonts SET status = $1, file_path = $2, updated_at = NOW() WHERE id = $3`,
		status, filePath, id,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fonts.ErrFontNotFound
	}
	return nil
}

func (r *PostgresFontRepo) Delete(ctx context.Context, id, userID string) error {
	result, err := r.db.ExecContext(ctx,
		`DELETE FROM fonts WHERE id = $1 AND user_id = $2 AND is_default = FALSE`, id, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fonts.ErrFontNotFound
	}
	return nil
}
```

- [ ] **Step 2: Wire font routes in main.go**

Add after notes routes:
```go
		fontRepo := repository.NewPostgresFontRepo(db)
		fileStorage := storage.NewLocalStorage(os.Getenv("STORAGE_PATH"))
		fontSvc := fonts.NewService(fontRepo, fileStorage)
		fontHandler := fonts.NewHandler(fontSvc)

		r.Route("/api/fonts", func(r chi.Router) {
			r.Use(auth.Middleware(jwtSvc))
			r.Get("/", fontHandler.List)
			r.Post("/", fontHandler.Upload)
			r.Get("/{id}", fontHandler.GetByID)
			r.Get("/{id}/file", fontHandler.ServeFile)
			r.Delete("/{id}", fontHandler.Delete)
		})
```

- [ ] **Step 3: Verify compilation**

```bash
cd services/api && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add services/api/
git commit -m "feat: add font repository and wire font routes"
```

---

### Task 5: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-05-fonts-api
gh pr create --title "feat: fonts API with upload, metadata, and file serving" --body "## Summary
- Add Font model with status enum
- Add local file storage abstraction
- Add fonts service and repository
- Add HTTP handlers for font CRUD and file upload
- Wire routes with auth middleware

## Test plan
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] Font upload accepts multipart form
- [ ] Font list returns user fonts + defaults
- [ ] Font file serving returns correct content type

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
