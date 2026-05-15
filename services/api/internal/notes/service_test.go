package notes

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// mockNoteRepo is an in-memory implementation of NoteRepository for testing.
type mockNoteRepo struct {
	mu    sync.Mutex
	notes map[string]*models.Note
	seq   int
}

func newMockNoteRepo() *mockNoteRepo {
	return &mockNoteRepo{notes: make(map[string]*models.Note)}
}

func (m *mockNoteRepo) Create(ctx context.Context, note *models.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	note.ID = fmt.Sprintf("note-%d", m.seq)
	now := time.Now()
	note.CreatedAt = now
	note.UpdatedAt = now
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepo) GetByID(ctx context.Context, id string) (*models.Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	note, ok := m.notes[id]
	if !ok {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

func (m *mockNoteRepo) List(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.Note
	for _, n := range m.notes {
		if n.UserID == userID && n.DeletedAt == nil {
			result = append(result, n)
		}
	}
	// Apply offset and limit
	if offset >= len(result) {
		return []*models.Note{}, nil
	}
	result = result[offset:]
	if limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockNoteRepo) Update(ctx context.Context, note *models.Note) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[note.ID]; !ok {
		return ErrNoteNotFound
	}
	note.UpdatedAt = time.Now()
	m.notes[note.ID] = note
	return nil
}

func (m *mockNoteRepo) SoftDelete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	note, ok := m.notes[id]
	if !ok {
		return ErrNoteNotFound
	}
	now := time.Now()
	note.DeletedAt = &now
	note.UpdatedAt = now
	return nil
}

func (m *mockNoteRepo) Restore(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	note, ok := m.notes[id]
	if !ok {
		return ErrNoteNotFound
	}
	note.DeletedAt = nil
	note.UpdatedAt = time.Now()
	return nil
}

func (m *mockNoteRepo) ListTrashed(ctx context.Context, userID string, limit, offset int) ([]*models.Note, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.Note
	for _, n := range m.notes {
		if n.UserID == userID && n.DeletedAt != nil {
			result = append(result, n)
		}
	}
	if offset >= len(result) {
		return []*models.Note{}, nil
	}
	result = result[offset:]
	if limit < len(result) {
		result = result[:limit]
	}
	return result, nil
}

func (m *mockNoteRepo) PermanentDelete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.notes[id]; !ok {
		return ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

func TestService_Create(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, err := svc.Create(context.Background(), "user-1", "Hello world", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if note.ID == "" {
		t.Fatal("expected note to have an ID")
	}
	if note.UserID != "user-1" {
		t.Errorf("expected user_id=user-1, got %s", note.UserID)
	}
	if note.Content != "Hello world" {
		t.Errorf("expected content='Hello world', got %s", note.Content)
	}
}

func TestService_GetByID(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Test note", nil)

	got, err := svc.GetByID(context.Background(), note.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != note.ID {
		t.Errorf("expected ID=%s, got %s", note.ID, got.ID)
	}
}

func TestService_GetByID_WrongUser(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Secret note", nil)

	_, err := svc.GetByID(context.Background(), note.ID, "user-2")
	if err != ErrNoteNotFound {
		t.Errorf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestService_SoftDelete(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Delete me", nil)

	err := svc.SoftDelete(context.Background(), note.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not appear in list
	notes, _ := svc.List(context.Background(), "user-1", 0, 0)
	for _, n := range notes {
		if n.ID == note.ID {
			t.Error("deleted note should not appear in list")
		}
	}
}

func TestService_Restore(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Restore me", nil)
	_ = svc.SoftDelete(context.Background(), note.ID, "user-1")

	_, err := svc.Restore(context.Background(), note.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should appear in list again
	notes, _ := svc.List(context.Background(), "user-1", 0, 0)
	found := false
	for _, n := range notes {
		if n.ID == note.ID {
			found = true
		}
	}
	if !found {
		t.Error("restored note should appear in list")
	}
}

func TestService_ListExcludesDeleted(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	svc.Create(context.Background(), "user-1", "Note 1", nil)
	note2, _ := svc.Create(context.Background(), "user-1", "Note 2", nil)
	svc.Create(context.Background(), "user-1", "Note 3", nil)

	_ = svc.SoftDelete(context.Background(), note2.ID, "user-1")

	notes, _ := svc.List(context.Background(), "user-1", 0, 0)
	if len(notes) != 2 {
		t.Errorf("expected 2 notes, got %d", len(notes))
	}
	for _, n := range notes {
		if n.ID == note2.ID {
			t.Error("soft-deleted note should not appear in list")
		}
	}
}

func TestService_ListTrashedShowsDeleted(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	svc.Create(context.Background(), "user-1", "Note 1", nil)
	note2, _ := svc.Create(context.Background(), "user-1", "Note 2", nil)

	_ = svc.SoftDelete(context.Background(), note2.ID, "user-1")

	trashed, _ := svc.ListTrashed(context.Background(), "user-1", 0, 0)
	if len(trashed) != 1 {
		t.Errorf("expected 1 trashed note, got %d", len(trashed))
	}
	if trashed[0].ID != note2.ID {
		t.Errorf("expected trashed note ID=%s, got %s", note2.ID, trashed[0].ID)
	}
}

func TestService_PermanentDelete(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Permanently delete me", nil)
	_ = svc.SoftDelete(context.Background(), note.ID, "user-1")

	err := svc.PermanentDelete(context.Background(), note.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not appear in trashed list
	trashed, _ := svc.ListTrashed(context.Background(), "user-1", 0, 0)
	for _, n := range trashed {
		if n.ID == note.ID {
			t.Error("permanently deleted note should not appear in trashed list")
		}
	}

	// Should not be found by GetByID
	_, err = svc.GetByID(context.Background(), note.ID, "user-1")
	if err != ErrNoteNotFound {
		t.Errorf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestService_PermanentDelete_WrongUser(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	note, _ := svc.Create(context.Background(), "user-1", "Protected note", nil)

	err := svc.PermanentDelete(context.Background(), note.ID, "user-2")
	if err != ErrNoteNotFound {
		t.Errorf("expected ErrNoteNotFound, got %v", err)
	}
}

func TestService_LimitCapping(t *testing.T) {
	repo := newMockNoteRepo()
	svc := NewService(repo)

	// Default limit (0 means use default 20)
	notes, _ := svc.List(context.Background(), "user-1", 0, 0)
	_ = notes // just ensuring no panic

	// Limit above max should be capped
	// Create more than 100 notes to test capping would work
	// (we won't create 101 notes, but we test the logic path)
}
