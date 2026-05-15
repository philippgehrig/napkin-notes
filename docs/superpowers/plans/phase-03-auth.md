# Phase 3: Auth API

> **For agentic workers:** Use superpowers:subagent-driven-development or superpowers:executing-plans.

**Goal:** Implement JWT-based authentication with register, login, refresh, and logout.

**Branch:** `feat/phase-03-auth`

---

## File Structure

```
services/api/
├── internal/
│   ├── auth/
│   │   ├── handler.go       (HTTP handlers)
│   │   ├── handler_test.go
│   │   ├── service.go       (business logic)
│   │   ├── service_test.go
│   │   ├── jwt.go           (token generation/validation)
│   │   ├── jwt_test.go
│   │   └── middleware.go    (auth middleware)
│   ├── models/
│   │   └── user.go
│   └── database/
│       └── (existing)
├── go.mod (add chi, bcrypt, jwt deps)
└── main.go (wire up chi router)
```

---

### Task 1: Add Chi router and auth dependencies

**Files:**
- Modify: `services/api/go.mod`
- Modify: `services/api/main.go`

- [ ] **Step 1: Add dependencies**

```bash
cd services/api && go get github.com/go-chi/chi/v5 github.com/go-chi/chi/v5/middleware github.com/golang-jwt/jwt/v5 golang.org/x/crypto/bcrypt github.com/google/uuid
```

- [ ] **Step 2: Refactor main.go to use Chi**

Replace `services/api/main.go`:
```go
package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/philippgehrig/napkin-notes/services/api/internal/database"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, `{"status":"ok"}`)
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migration driver: %w", err)
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		ex, _ := os.Executable()
		migrationsPath = filepath.Join(filepath.Dir(ex), "migrations")
	}
	m, err := migrate.NewWithDatabaseInstance("file://"+migrationsPath, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migration init: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migration up: %w", err)
	}
	return nil
}

func main() {
	var db *sql.DB
	dsn := database.BuildDSN(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_SSLMODE"),
	)

	if os.Getenv("DB_HOST") != "" {
		var err error
		db, err = database.Connect(dsn)
		if err != nil {
			log.Fatalf("database connection failed: %v", err)
		}
		if os.Getenv("RUN_MIGRATIONS") == "true" {
			if err := runMigrations(db); err != nil {
				log.Fatalf("migrations failed: %v", err)
			}
			log.Println("migrations completed successfully")
		}
	}

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)

	r.Get("/health", healthHandler)

	// Auth routes will be added here
	_ = db

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("API listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
```

- [ ] **Step 3: Run existing test**

```bash
cd services/api && go test ./... -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/
git commit -m "feat: switch to Chi router with middleware"
```

---

### Task 2: User model

**Files:**
- Create: `services/api/internal/models/user.go`

- [ ] **Step 1: Create user model**

```go
package models

import "time"

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	DisplayName  string    `json:"display_name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: Commit**

```bash
git add services/api/internal/models/
git commit -m "feat: add User model"
```

---

### Task 3: JWT token generation and validation

**Files:**
- Create: `services/api/internal/auth/jwt.go`
- Create: `services/api/internal/auth/jwt_test.go`

- [ ] **Step 1: Write JWT tests**

Create `services/api/internal/auth/jwt_test.go`:
```go
package auth

import (
	"testing"
	"time"
)

func TestGenerateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")
	token, err := jwtService.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")
	token, _ := jwtService.GenerateAccessToken("user-123")

	userID, err := jwtService.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("expected user-123, got %s", userID)
	}
}

func TestValidateAccessToken_Expired(t *testing.T) {
	jwtService := &JWTService{
		secret:          []byte("test-secret"),
		accessTokenTTL:  -1 * time.Hour,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
	token, _ := jwtService.GenerateAccessToken("user-123")

	_, err := jwtService.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	jwtService1 := NewJWTService("secret-1")
	jwtService2 := NewJWTService("secret-2")

	token, _ := jwtService1.GenerateAccessToken("user-123")
	_, err := jwtService2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")
	token, err := jwtService.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateRefreshToken(t *testing.T) {
	jwtService := NewJWTService("test-secret")
	token, _ := jwtService.GenerateRefreshToken("user-123")

	userID, err := jwtService.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if userID != "user-123" {
		t.Errorf("expected user-123, got %s", userID)
	}
}
```

- [ ] **Step 2: Run tests — expect fail**

```bash
cd services/api && go test ./internal/auth/ -v
```

- [ ] **Step 3: Implement JWT service**

Create `services/api/internal/auth/jwt.go`:
```go
package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
}

func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "access",
		"exp":  time.Now().Add(s.accessTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": "refresh",
		"exp":  time.Now().Add(s.refreshTokenTTL).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) ValidateAccessToken(tokenStr string) (string, error) {
	return s.validateToken(tokenStr, "access")
}

func (s *JWTService) ValidateRefreshToken(tokenStr string) (string, error) {
	return s.validateToken(tokenStr, "refresh")
}

func (s *JWTService) validateToken(tokenStr, expectedType string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.secret, nil
	})
	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != expectedType {
		return "", fmt.Errorf("expected %s token, got %s", expectedType, tokenType)
	}

	userID, _ := claims["sub"].(string)
	if userID == "" {
		return "", fmt.Errorf("missing user ID in token")
	}

	return userID, nil
}
```

- [ ] **Step 4: Run tests — expect pass**

```bash
cd services/api && go test ./internal/auth/ -v
```

- [ ] **Step 5: Commit**

```bash
git add services/api/internal/auth/
git commit -m "feat: add JWT service with access and refresh tokens"
```

---

### Task 4: Auth service (business logic)

**Files:**
- Create: `services/api/internal/auth/service.go`
- Create: `services/api/internal/auth/service_test.go`

- [ ] **Step 1: Write auth service tests**

Create `services/api/internal/auth/service_test.go`:
```go
package auth

import (
	"context"
	"testing"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type mockUserRepo struct {
	users map[string]*models.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*models.User)}
}

func (m *mockUserRepo) CreateUser(ctx context.Context, email, passwordHash, displayName string) (*models.User, error) {
	user := &models.User{
		ID:           "generated-uuid",
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
	}
	m.users[email] = user
	return user, nil
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user, ok := m.users[email]
	if !ok {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, ErrUserNotFound
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	jwt := NewJWTService("test-secret")
	svc := NewService(repo, jwt)

	result, err := svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected refresh token")
	}
	if result.User.Email != "test@example.com" {
		t.Errorf("expected test@example.com, got %s", result.User.Email)
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	jwt := NewJWTService("test-secret")
	svc := NewService(repo, jwt)

	_, _ = svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	_, err := svc.Register(context.Background(), "test@example.com", "password456", "Test User 2")
	if err != ErrEmailTaken {
		t.Errorf("expected ErrEmailTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	repo := newMockUserRepo()
	jwt := NewJWTService("test-secret")
	svc := NewService(repo, jwt)

	_, _ = svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	result, err := svc.Login(context.Background(), "test@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected access token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	jwt := NewJWTService("test-secret")
	svc := NewService(repo, jwt)

	_, _ = svc.Register(context.Background(), "test@example.com", "password123", "Test User")
	_, err := svc.Login(context.Background(), "test@example.com", "wrongpassword")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	jwt := NewJWTService("test-secret")
	svc := NewService(repo, jwt)

	_, err := svc.Login(context.Background(), "nobody@example.com", "password123")
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}
```

- [ ] **Step 2: Implement auth service**

Create `services/api/internal/auth/service.go`:
```go
package auth

import (
	"context"
	"errors"

	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrEmailTaken       = errors.New("email already registered")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash, displayName string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

type AuthResult struct {
	User         *models.User `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}

type Service struct {
	users UserRepository
	jwt   *JWTService
}

func NewService(users UserRepository, jwt *JWTService) *Service {
	return &Service{users: users, jwt: jwt}
}

func (s *Service) Register(ctx context.Context, email, password, displayName string) (*AuthResult, error) {
	existing, _ := s.users.GetUserByEmail(ctx, email)
	if existing != nil {
		return nil, ErrEmailTaken
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user, err := s.users.CreateUser(ctx, email, string(hash), displayName)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

func (s *Service) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

func (s *Service) RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	userID, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

func (s *Service) generateTokens(user *models.User) (*AuthResult, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
```

- [ ] **Step 3: Run tests — expect pass**

```bash
cd services/api && go test ./internal/auth/ -v
```

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/
git commit -m "feat: add auth service with register, login, refresh"
```

---

### Task 5: Auth middleware

**Files:**
- Create: `services/api/internal/auth/middleware.go`
- Add tests to: `services/api/internal/auth/jwt_test.go`

- [ ] **Step 1: Implement auth middleware**

Create `services/api/internal/auth/middleware.go`:
```go
package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Middleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			userID, err := jwtService.ValidateAccessToken(parts[1])
			if err != nil {
				http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(UserIDKey).(string)
	return userID
}
```

- [ ] **Step 2: Commit**

```bash
git add services/api/internal/auth/middleware.go
git commit -m "feat: add JWT auth middleware"
```

---

### Task 6: Auth HTTP handlers

**Files:**
- Create: `services/api/internal/auth/handler.go`
- Create: `services/api/internal/auth/handler_test.go`

- [ ] **Step 1: Write handler tests**

Create `services/api/internal/auth/handler_test.go`:
```go
package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestHandler() *Handler {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	svc := NewService(repo, jwtSvc)
	return NewHandler(svc)
}

func TestHandleRegister_Success(t *testing.T) {
	h := setupTestHandler()
	body := `{"email":"test@example.com","password":"password123","display_name":"Test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var result AuthResult
	json.NewDecoder(w.Body).Decode(&result)
	if result.AccessToken == "" {
		t.Error("expected access token in response")
	}
}

func TestHandleRegister_MissingFields(t *testing.T) {
	h := setupTestHandler()
	body := `{"email":"test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestHandleLogin_Success(t *testing.T) {
	h := setupTestHandler()

	// Register first
	regBody := `{"email":"test@example.com","password":"password123","display_name":"Test"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	h.Register(regW, regReq)

	// Login
	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	h := setupTestHandler()

	regBody := `{"email":"test@example.com","password":"password123","display_name":"Test"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	h.Register(regW, regReq)

	body := `{"email":"test@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
```

- [ ] **Step 2: Implement handler**

Create `services/api/internal/auth/handler.go`:
```go
package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

type registerRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.service.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, "email already registered", http.StatusConflict)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, result, http.StatusCreated)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Email == "" || req.Password == "" {
		writeError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, result, http.StatusOK)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.RefreshToken == "" {
		writeError(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	result, err := h.service.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, "invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	writeJSON(w, result, http.StatusOK)
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
cd services/api && go test ./internal/auth/ -v
```
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add services/api/internal/auth/
git commit -m "feat: add auth HTTP handlers (register, login, refresh)"
```

---

### Task 7: Wire auth routes into main

**Files:**
- Modify: `services/api/main.go`
- Create: `services/api/internal/repository/user_repo.go`

- [ ] **Step 1: Create PostgreSQL user repository**

Create `services/api/internal/repository/user_repo.go`:
```go
package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/auth"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
)

type PostgresUserRepo struct {
	db *sql.DB
}

func NewPostgresUserRepo(db *sql.DB) *PostgresUserRepo {
	return &PostgresUserRepo{db: db}
}

func (r *PostgresUserRepo) CreateUser(ctx context.Context, email, passwordHash, displayName string) (*models.User, error) {
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		DisplayName:  displayName,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO users (id, email, password_hash, display_name, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		user.ID, user.Email, user.PasswordHash, user.DisplayName, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *PostgresUserRepo) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, display_name, created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	return user, err
}

func (r *PostgresUserRepo) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	user := &models.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, email, password_hash, display_name, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	return user, err
}
```

- [ ] **Step 2: Wire routes in main.go**

Add to `main.go` after the health route:
```go
	if db != nil {
		jwtSecret := os.Getenv("JWT_SECRET")
		if jwtSecret == "" {
			jwtSecret = "dev-secret-change-me"
		}

		userRepo := repository.NewPostgresUserRepo(db)
		jwtSvc := auth.NewJWTService(jwtSecret)
		authSvc := auth.NewService(userRepo, jwtSvc)
		authHandler := auth.NewHandler(authSvc)

		r.Route("/api/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
		})
	}
```

Add imports for `repository` package.

- [ ] **Step 3: Verify compilation**

```bash
cd services/api && go build ./...
```

- [ ] **Step 4: Commit**

```bash
git add services/api/
git commit -m "feat: wire auth routes and add PostgreSQL user repository"
```

---

### Task 8: Create PR

- [ ] **Step 1: Push and create PR**

```bash
git push -u origin feat/phase-03-auth
gh pr create --title "feat: JWT authentication with register, login, refresh" --body "## Summary
- Add JWT service for access/refresh token generation and validation
- Add auth service with register, login, refresh business logic
- Add auth HTTP handlers with validation
- Add auth middleware for protected routes
- Add PostgreSQL user repository
- Wire routes into Chi router

## Test plan
- [ ] \`cd services/api && go test ./... -v\` passes
- [ ] Register endpoint returns tokens on valid input
- [ ] Login rejects wrong password with 401
- [ ] Refresh generates new token pair

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
