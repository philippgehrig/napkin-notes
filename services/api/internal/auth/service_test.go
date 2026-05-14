package auth

import (
	"context"
	"sync"
	"testing"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

// mockUserRepo is an in-memory implementation of UserRepository for testing.
type mockUserRepo struct {
	mu    sync.RWMutex
	users map[string]*models.User // keyed by ID
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*models.User)}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Check for duplicate email
	for _, u := range m.users {
		if u.Email == user.Email {
			return ErrEmailTaken
		}
	}
	m.users[user.ID] = user
	return nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, u := range m.users {
		if u.Email == email {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func TestRegisterSuccess(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	result, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.User == nil {
		t.Fatal("expected non-nil user")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
	if result.User.DisplayName != "Test User" {
		t.Errorf("expected display name 'Test User', got %s", result.User.DisplayName)
	}
	if result.User.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}
	if result.User.PasswordHash == "password123" {
		t.Error("password should be hashed, not stored in plain text")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	_, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error on first register: %v", err)
	}

	_, err = svc.Register(context.Background(), "test@example.com", "otherpass", "Other User")
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLoginSuccess(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	_, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error on login: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", result.User.Email)
	}
}

func TestLoginWrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	_, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	_, err = svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLoginUserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	_, err := svc.Login(context.Background(), "nonexistent@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestRefreshAccessTokenSuccess(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	regResult, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error on register: %v", err)
	}

	result, err := svc.RefreshAccessToken(context.Background(), regResult.RefreshToken)
	if err != nil {
		t.Fatalf("unexpected error on refresh: %v", err)
	}

	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestRefreshAccessTokenInvalid(t *testing.T) {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewAuthService(repo, jwtSvc)

	_, err := svc.RefreshAccessToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}
