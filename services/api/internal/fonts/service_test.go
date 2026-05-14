package fonts

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"github.com/philippgehrig/napkin-notes/services/api/internal/storage"
)

// mockFontRepo is an in-memory implementation of FontRepository for testing.
type mockFontRepo struct {
	mu    sync.Mutex
	fonts map[string]*models.Font
	seq   int
}

func newMockFontRepo() *mockFontRepo {
	return &mockFontRepo{fonts: make(map[string]*models.Font)}
}

func (m *mockFontRepo) Create(ctx context.Context, font *models.Font) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seq++
	font.ID = fmt.Sprintf("font-%d", m.seq)
	now := time.Now()
	font.CreatedAt = now
	font.UpdatedAt = now
	m.fonts[font.ID] = font
	return nil
}

func (m *mockFontRepo) GetByID(ctx context.Context, id string) (*models.Font, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	font, ok := m.fonts[id]
	if !ok {
		return nil, ErrFontNotFound
	}
	return font, nil
}

func (m *mockFontRepo) ListByUser(ctx context.Context, userID string) ([]*models.Font, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.Font
	for _, f := range m.fonts {
		if f.UserID == userID || f.IsDefault {
			result = append(result, f)
		}
	}
	if result == nil {
		result = []*models.Font{}
	}
	return result, nil
}

func (m *mockFontRepo) UpdateStatus(ctx context.Context, id string, status models.FontStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	font, ok := m.fonts[id]
	if !ok {
		return ErrFontNotFound
	}
	font.Status = status
	font.UpdatedAt = time.Now()
	return nil
}

func (m *mockFontRepo) Delete(ctx context.Context, id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	font, ok := m.fonts[id]
	if !ok {
		return ErrFontNotFound
	}
	if font.IsDefault {
		return ErrCannotDeleteDefault
	}
	delete(m.fonts, id)
	return nil
}

// mockStorage is an in-memory implementation of storage.Storage for testing.
type mockStorage struct {
	mu    sync.Mutex
	files map[string][]byte
}

func newMockStorage() *mockStorage {
	return &mockStorage{files: make(map[string][]byte)}
}

func (s *mockStorage) Save(path string, reader io.Reader) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	s.files[path] = data
	return nil
}

func (s *mockStorage) Get(path string) (io.ReadCloser, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, ok := s.files[path]
	if !ok {
		return nil, fmt.Errorf("storage: file not found: %s", path)
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (s *mockStorage) Delete(path string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.files, path)
	return nil
}

// Verify mockStorage satisfies storage.Storage at compile time.
var _ storage.Storage = (*mockStorage)(nil)

func newTestService() (*Service, *mockFontRepo, *mockStorage) {
	repo := newMockFontRepo()
	store := newMockStorage()
	svc := NewService(repo, store)
	return svc, repo, store
}

func TestService_Create(t *testing.T) {
	svc, _, store := newTestService()

	font, err := svc.Create(context.Background(), "user-1", "My Font", "scans/user-1/scan.png", &fakeReader{data: []byte("scan-data")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if font.ID == "" {
		t.Fatal("expected font to have an ID")
	}
	if font.UserID != "user-1" {
		t.Errorf("expected user_id=user-1, got %s", font.UserID)
	}
	if font.Name != "My Font" {
		t.Errorf("expected name='My Font', got %s", font.Name)
	}
	if font.Status != models.FontStatusPending {
		t.Errorf("expected status=pending, got %s", font.Status)
	}
	if font.TemplateScanPath != "scans/user-1/scan.png" {
		t.Errorf("expected template_scan_path='scans/user-1/scan.png', got %s", font.TemplateScanPath)
	}

	// Verify file was saved to storage
	store.mu.Lock()
	_, exists := store.files["scans/user-1/scan.png"]
	store.mu.Unlock()
	if !exists {
		t.Error("expected scan file to be saved in storage")
	}
}

func TestService_List(t *testing.T) {
	svc, repo, _ := newTestService()

	// Add a default font directly to the repo
	repo.mu.Lock()
	repo.seq++
	repo.fonts["font-default"] = &models.Font{
		ID:        "font-default",
		UserID:    "system",
		Name:      "Default Font",
		Status:    models.FontStatusReady,
		IsDefault: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo.mu.Unlock()

	// Create a user font
	svc.Create(context.Background(), "user-1", "User Font", "scans/user-1/scan.png", &fakeReader{data: []byte("data")})

	fonts, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fonts) != 2 {
		t.Errorf("expected 2 fonts (1 user + 1 default), got %d", len(fonts))
	}
}

func TestService_Delete(t *testing.T) {
	svc, _, _ := newTestService()

	font, _ := svc.Create(context.Background(), "user-1", "Delete Me", "scans/user-1/del.png", &fakeReader{data: []byte("data")})

	err := svc.Delete(context.Background(), font.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not be found anymore
	_, err = svc.GetByID(context.Background(), font.ID, "user-1")
	if err != ErrFontNotFound {
		t.Errorf("expected ErrFontNotFound, got %v", err)
	}
}

func TestService_Delete_WrongUser(t *testing.T) {
	svc, _, _ := newTestService()

	font, _ := svc.Create(context.Background(), "user-1", "Not Yours", "scans/user-1/nope.png", &fakeReader{data: []byte("data")})

	err := svc.Delete(context.Background(), font.ID, "user-2")
	if err != ErrFontNotFound {
		t.Errorf("expected ErrFontNotFound, got %v", err)
	}
}

func TestService_GetByID(t *testing.T) {
	svc, _, _ := newTestService()

	font, _ := svc.Create(context.Background(), "user-1", "Get Me", "scans/user-1/get.png", &fakeReader{data: []byte("data")})

	got, err := svc.GetByID(context.Background(), font.ID, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != font.ID {
		t.Errorf("expected ID=%s, got %s", font.ID, got.ID)
	}
}

func TestService_GetByID_NotFound(t *testing.T) {
	svc, _, _ := newTestService()

	_, err := svc.GetByID(context.Background(), "nonexistent", "user-1")
	if err != ErrFontNotFound {
		t.Errorf("expected ErrFontNotFound, got %v", err)
	}
}

// fakeReader is a simple io.Reader for testing.
type fakeReader struct {
	data []byte
	pos  int
}

func (r *fakeReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
