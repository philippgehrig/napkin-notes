# Phase 4: Notes CRUD API

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Implement full notes CRUD with soft-delete, trash, and restore.

**Branch:** `feat/phase-04-notes-api`

---

## File Structure

```
services/api/internal/
├── models/
│   └── note.go
├── notes/
│   ├── handler.go
│   ├── handler_test.go
│   ├── service.go
│   └── service_test.go
└── repository/
    └── note_repo.go
```

---

### Task 1: Note model

**Files:**
- Create: `services/api/internal/models/note.go`

- [ ] **Step 1: Create note model**

```go
package models

import "time"

type Note struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Content   string     `json:"content"`
	FontID    *string    `json:"font_id,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}
```

- [ ] **Step 2: Commit**

```bash
git add services/api/internal/models/note.go
git commit -m "feat: add Note model"
```

---

### Task 2: Notes service with tests

**Files:**
- Create: `services/api/internal/notes/service.go`
- Create: `services/api/internal/notes/service_test.go`

- [ ] **Step 1: Write service tests**

Create `services/api/internal/notes/service_test.go`:
```go
package notes

import (
	"context"
	"testing"
	"time"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type mockNoteRepo struct {
	notes map[string]*models.Note
}

func newMockNoteRepo() *mockNoteRepo {
	return &mockNoteRepo{notes: make(map[string]*models.Note)}
}

func (m *mockNoteRepo) Create(ctx context.Context, userID, content string, fontID *string) (*models.Note, error) {
	note := &models.Note{
		ID:        "note-1",
		UserID:    userID,
		Content:   content,
		FontID:    fontID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.notes[note.ID] = note
	return note, nil
}

func (m *mockNoteRepo) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	note, ok := m.notes[id]
	if !ok || note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

func (m *mockNoteRepo) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	var result []*models.Note
	for _, n := range m.notes {
		if n.UserID == userID && n.DeletedAt == nil {
			result = append(result, n)
		}
	}
	return result, nil
}

func (m *mockNoteRepo) Update(ctx context.Context, id, userID, content string, fontID *string) (*models.Note, error) {
	note, ok := m.notes[id]
	if !ok || note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	note.Content = content
	note.FontID = fontID
	note.UpdatedAt = time.Now()
	return note, nil
}

func (m *mockNoteRepo) SoftDelete(ctx context.Context, id, userID string) error {
	note, ok := m.notes[id]
	if !ok || note.UserID != userID {
		return ErrNoteNotFound
	}
	now := time.Now()
	note.DeletedAt = &now
	return nil
}

func (m *mockNoteRepo) Restore(ctx context.Context, id, userID string) (*models.Note, error) {
	note, ok := m.notes[id]
	if !ok || note.UserID != userID {
		return nil, ErrNoteNotFound
	}
	note.DeletedAt = nil
	return note, nil
}

func (m *mockNoteRepo) ListTrashed(ctx context.Context, userID string) ([]*models.Note, error) {
	var result []*models.Note
	for _, n := range m.notes {
		if n.UserID == userID && n.DeletedAt != nil {
			result = append(result, n)
		}
	}
	return result, nil
}

func TestCreateNote(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, err := svc.Create(context.Background(), "user-1", "Hello world", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.Content != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", note.Content)
	}
	if note.UserID != "user-1" {
		t.Errorf("expected user-1, got %s", note.UserID)
	}
}

func TestGetNote(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	created, _ := svc.Create(context.Background(), "user-1", "Test note", nil)
	note, err := svc.GetByID(context.Background(), created.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID != created.ID {
		t.Errorf("expected %s, got %s", created.ID, note.ID)
	}
}

func TestGetNote_WrongUser(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	created, _ := svc.Create(context.Background(), "user-1", "Test note", nil)
	_, err := svc.GetByID(context.Background(), created.ID, "user-2")
	if err != ErrNoteNotFound {
		t.Errorf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestSoftDelete(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	created, _ := svc.Create(context.Background(), "user-1", "Test note", nil)
	err := svc.SoftDelete(context.Background(), created.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not appear in regular list
	notes, _ := svc.List(context.Background(), "user-1", 10, 0)
	if len(notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(notes))
	}

	// Should appear in trash
	trashed, _ := svc.ListTrashed(context.Background(), "user-1")
	if len(trashed) != 1 {
		t.Errorf("expected 1 trashed note, got %d", len(trashed))
	}
}

func TestRestore(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	created, _ := svc.Create(context.Background(), "user-1", "Test note", nil)
	_ = svc.SoftDelete(context.Background(), created.ID, "user-1")
	_, err := svc.Restore(context.Background(), created.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	notes, _ := svc.List(context.Background(), "user-1", 10, 0)
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}
```

- [ ] **Step 2: Implement notes service**

Create `services/api/internal/notes/service.go`:
```go
package notes

import (
	"context"
	"errors"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

var ErrNoteNotFound = errors.New("note not found")

type NoteRepository interface {
	Create(ctx context.Context, userID, content string, fontID *string) (*models.Note, error)
	GetByID(ctx context.Context, id, userID string) (*models.Note, error)
	List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error)
	Update(ctx context.Context, id, userID, content string, fontID *string) (*models.Note, error)
	SoftDelete(ctx context.Context, id, userID string) error
	Restore(ctx context.Context, id, userID string) (*models.Note, error)
	ListTrashed(ctx context.Context, userID string) ([]*models.Note, error)
}

type Service struct {
	repo NoteRepository
}

func NewService(repo NoteRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID, content string, fontID *string) (*models.Note, error) {
	return s.repo.Create(ctx, userID, content, fontID)
}

func (s *Service) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	return s.repo.GetByID(ctx, id, userID)
}

func (s *Service) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return s.repo.List(ctx, userID, limit, offset)
}

func (s *Service) Update(ctx context.Context, id, userID, content string, fontID *string) (*models.Note, error) {
	return s.repo.Update(ctx, id, userID, content, fontID)
}

func (s *Service) SoftDelete(ctx context.Context, id, userID string) error {
	return s.repo.SoftDelete(ctx, id, userID)
}

func (s *Service) Restore(ctx context.Context, id, userID string) (*models.Note, error) {
	return s.repo.Restore(ctx, id, userID)
}

func (s *Service) ListTrashed(ctx context.Context, userID string) ([]*models.Note, error) {
	return s.repo.ListTrashed(ctx, userID)
}
```

- [ ] **Step 3: Run tests**

```bash
cd services/api && go test ./internal/notes/ -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/notes/ services/api/internal/models/note.go
git commit -m "feat: add notes service with CRUD and soft-delete"
```

---

### Task 3: Notes HTTP handlers

**Files:**
- Create: `services/api/internal/notes/handler.go`
- Create: `services/api/internal/notes/handler_test.go`

- [ ] **Step 1: Write handler tests**

Create `services/api/internal/notes/handler_test.go`:
```go
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
)

func setupTestRouter() (*chi.Mux, *Handler) {
	repo := newMockNoteRepo()
	svc := NewService(repo)
	h := NewHandler(svc)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), auth.UserIDKey, "user-1")
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Post("/api/notes", h.Create)
	r.Get("/api/notes", h.List)
	r.Get("/api/notes/{id}", h.GetByID)
	r.Put("/api/notes/{id}", h.Update)
	r.Delete("/api/notes/{id}", h.Delete)
	r.Get("/api/notes/trash", h.ListTrashed)
	r.Post("/api/notes/{id}/restore", h.Restore)

	return r, h
}

func TestHandleCreateNote(t *testing.T) {
	r, _ := setupTestRouter()

	body := `{"content":"My first napkin note"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var note map[string]interface{}
	json.NewDecoder(w.Body).Decode(&note)
	if note["content"] != "My first napkin note" {
		t.Errorf("unexpected content: %v", note["content"])
	}
}

func TestHandleListNotes(t *testing.T) {
	r, _ := setupTestRouter()

	// Create a note first
	body := `{"content":"Note 1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// List notes
	req = httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var notes []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&notes)
	if len(notes) != 1 {
		t.Errorf("expected 1 note, got %d", len(notes))
	}
}

func TestHandleDeleteNote(t *testing.T) {
	r, _ := setupTestRouter()

	// Create
	body := `{"content":"To be deleted"}`
	req := httptest.NewRequest(http.MethodPost, "/api/notes", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var created map[string]interface{}
	json.NewDecoder(w.Body).Decode(&created)

	// Delete
	req = httptest.NewRequest(http.MethodDelete, "/api/notes/"+created["id"].(string), nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d: %s", w.Code, w.Body.String())
	}

	// Verify not in list
	req = httptest.NewRequest(http.MethodGet, "/api/notes", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	var notes []map[string]interface{}
	json.NewDecoder(w.Body).Decode(&notes)
	if len(notes) != 0 {
		t.Errorf("expected 0 notes, got %d", len(notes))
	}
}
```

- [ ] **Step 2: Implement handler**

Create `services/api/internal/notes/handler.go`:
```go
package notes

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type createRequest struct {
	Content string  `json:"content"`
	FontID  *string `json:"font_id,omitempty"`
}

type updateRequest struct {
	Content string  `json:"content"`
	FontID  *string `json:"font_id,omitempty"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	note, err := h.service.Create(r.Context(), userID, req.Content, req.FontID)
	if err != nil {
		writeError(w, "failed to create note", http.StatusInternalServerError)
		return
	}

	writeJSON(w, note, http.StatusCreated)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.service.GetByID(r.Context(), noteID, userID)
	if err != nil {
		writeError(w, "note not found", http.StatusNotFound)
		return
	}

	writeJSON(w, note, http.StatusOK)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	notes, err := h.service.List(r.Context(), userID, limit, offset)
	if err != nil {
		writeError(w, "failed to list notes", http.StatusInternalServerError)
		return
	}

	writeJSON(w, notes, http.StatusOK)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	note, err := h.service.Update(r.Context(), noteID, userID, req.Content, req.FontID)
	if err != nil {
		writeError(w, "note not found", http.StatusNotFound)
		return
	}

	writeJSON(w, note, http.StatusOK)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	err := h.service.SoftDelete(r.Context(), noteID, userID)
	if err != nil {
		writeError(w, "note not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListTrashed(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())

	notes, err := h.service.ListTrashed(r.Context(), userID)
	if err != nil {
		writeError(w, "failed to list trashed notes", http.StatusInternalServerError)
		return
	}

	writeJSON(w, notes, http.StatusOK)
}

func (h *Handler) Restore(w http.ResponseWriter, r *http.Request) {
	userID := auth.GetUserID(r.Context())
	noteID := chi.URLParam(r, "id")

	note, err := h.service.Restore(r.Context(), noteID, userID)
	if err != nil {
		writeError(w, "note not found", http.StatusNotFound)
		return
	}

	writeJSON(w, note, http.StatusOK)
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
cd services/api && go test ./internal/notes/ -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/notes/
git commit -m "feat: add notes HTTP handlers with CRUD and trash"
```

---

### Task 4: Notes PostgreSQL repository

**Files:**
- Create: `services/api/internal/repository/note_repo.go`

- [ ] **Step 1: Implement note repository**

Create `services/api/internal/repository/note_repo.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/notes"
)

type PostgresNoteRepo struct {
	db *sql.DB
}

func NewPostgresNoteRepo(db *sql.DB) *PostgresNoteRepo {
	return &PostgresNoteRepo{db: db}
}

func (r *PostgresNoteRepo) Create(ctx context.Context, userID, content string, fontID *string) (*models.Note, error) {
	note := &models.Note{
		ID:        uuid.New().String(),
		UserID:    userID,
		Content:   content,
		FontID:    fontID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO notes (id, user_id, content, font_id, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		note.ID, note.UserID, note.Content, note.FontID, note.CreatedAt, note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

func (r *PostgresNoteRepo) GetByID(ctx context.Context, id, userID string) (*models.Note, error) {
	note := &models.Note{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		 FROM notes WHERE id = $1 AND user_id = $2`, id, userID,
	).Scan(&note.ID, &note.UserID, &note.Content, &note.FontID, &note.DeletedAt, &note.CreatedAt, &note.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, notes.ErrNoteNotFound
	}
	return note, err
}

func (r *PostgresNoteRepo) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		 FROM notes WHERE user_id = $1 AND deleted_at IS NULL
		 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, userID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Note
	for rows.Next() {
		note := &models.Note{}
		if err := rows.Scan(&note.ID, &note.UserID, &note.Content, &note.FontID, &note.DeletedAt, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, note)
	}
	return result, nil
}

func (r *PostgresNoteRepo) Update(ctx context.Context, id, userID, content string, fontID *string) (*models.Note, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notes SET content = $1, font_id = $2, updated_at = NOW()
		 WHERE id = $3 AND user_id = $4 AND deleted_at IS NULL`, content, fontID, id, userID,
	)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, notes.ErrNoteNotFound
	}
	return r.GetByID(ctx, id, userID)
}

func (r *PostgresNoteRepo) SoftDelete(ctx context.Context, id, userID string) error {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notes SET deleted_at = NOW(), updated_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, id, userID,
	)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return notes.ErrNoteNotFound
	}
	return nil
}

func (r *PostgresNoteRepo) Restore(ctx context.Context, id, userID string) (*models.Note, error) {
	result, err := r.db.ExecContext(ctx,
		`UPDATE notes SET deleted_at = NULL, updated_at = NOW()
		 WHERE id = $1 AND user_id = $2 AND deleted_at IS NOT NULL`, id, userID,
	)
	if err != nil {
		return nil, err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, notes.ErrNoteNotFound
	}
	return r.GetByID(ctx, id, userID)
}

func (r *PostgresNoteRepo) ListTrashed(ctx context.Context, userID string) ([]*models.Note, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, content, font_id, deleted_at, created_at, updated_at
		 FROM notes WHERE user_id = $1 AND deleted_at IS NOT NULL
		 ORDER BY deleted_at DESC`, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*models.Note
	for rows.Next() {
		note := &models.Note{}
		if err := rows.Scan(&note.ID, &note.UserID, &note.Content, &note.FontID, &note.DeletedAt, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		result = append(result, note)
	}
	return result, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add services/api/internal/repository/note_repo.go
git commit -m "feat: add PostgreSQL note repository"
```

---

### Task 5: Wire notes routes in main.go

**Files:**
- Modify: `services/api/main.go`

- [ ] **Step 1: Add notes routes to main**

Add to `main.go` after auth routes:
```go
		noteRepo := repository.NewPostgresNoteRepo(db)
		noteSvc := notes.NewService(noteRepo)
		noteHandler := notes.NewHandler(noteSvc)

		r.Route("/api/notes", func(r chi.Router) {
			r.Use(auth.Middleware(jwtSvc))
			r.Post("/", noteHandler.Create)
			r.Get("/", noteHandler.List)
			r.Get("/trash", noteHandler.ListTrashed)
			r.Get("/{id}", noteHandler.GetByID)
			r.Put("/{id}", noteHandler.Update)
			r.Delete("/{id}", noteHandler.Delete)
			r.Post("/{id}/restore", noteHandler.Restore)
		})
```

Add import for `notes` package.

- [ ] **Step 2: Verify compilation**

```bash
cd services/api && go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add services/api/main.go
git commit -m "feat: wire notes routes with auth middleware"
```

---

### Task 6: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-04-notes-api
gh pr create --title "feat: notes CRUD API with soft-delete and trash" --body "## Summary
- Add Note model with soft-delete support
- Add notes service with create, read, update, soft-delete, restore
- Add HTTP handlers for all notes endpoints
- Add PostgreSQL note repository
- Wire routes with auth middleware protection

## Test plan
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] Create note returns 201 with note data
- [ ] Delete soft-deletes (appears in trash, not in list)
- [ ] Restore moves note back from trash

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
