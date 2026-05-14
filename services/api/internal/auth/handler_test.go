package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupTestHandler() *AuthHandler {
	repo := newMockUserRepo()
	jwtSvc := NewJWTService("test-secret")
	authSvc := NewAuthService(repo, jwtSvc)
	return NewAuthHandler(authSvc)
}

func TestHandlerRegisterSuccess(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
	if resp["user"] == nil {
		t.Error("expected user in response")
	}
}

func TestHandlerRegisterMissingEmail(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegisterMissingPassword(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"email":        "test@example.com",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerRegisterDuplicateEmail(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Second registration with same email
	jsonBody, _ = json.Marshal(body)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected status 409, got %d", w.Code)
	}
}

func TestHandlerLoginSuccess(t *testing.T) {
	handler := setupTestHandler()

	// Register first
	regBody := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Login
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
}

func TestHandlerLoginMissingFields(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"email": "test@example.com",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestHandlerLoginWrongPassword(t *testing.T) {
	handler := setupTestHandler()

	// Register
	regBody := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	// Login with wrong password
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ = json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerRefreshSuccess(t *testing.T) {
	handler := setupTestHandler()

	// Register to get a refresh token
	regBody := map[string]string{
		"email":        "test@example.com",
		"password":     "password123",
		"display_name": "Test User",
	}
	jsonBody, _ := json.Marshal(regBody)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Register(w, req)

	var regResp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&regResp)
	refreshToken := regResp["refresh_token"].(string)

	// Refresh
	refreshBody := map[string]string{
		"refresh_token": refreshToken,
	}
	jsonBody, _ = json.Marshal(refreshBody)
	req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	handler.Refresh(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp["access_token"] == nil || resp["access_token"] == "" {
		t.Error("expected access_token in response")
	}
	if resp["refresh_token"] == nil || resp["refresh_token"] == "" {
		t.Error("expected refresh_token in response")
	}
}

func TestHandlerRefreshInvalidToken(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{
		"refresh_token": "invalid-token",
	}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Refresh(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestHandlerRefreshMissingToken(t *testing.T) {
	handler := setupTestHandler()

	body := map[string]string{}
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler.Refresh(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}
