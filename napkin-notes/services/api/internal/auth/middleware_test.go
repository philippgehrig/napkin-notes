package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMiddlewareValidToken(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	token, err := jwtSvc.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var gotUserID string
	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if gotUserID != "user-123" {
		t.Errorf("expected userID 'user-123', got %q", gotUserID)
	}
}

func TestMiddlewareMissingToken(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddlewareInvalidToken(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestMiddlewareMalformedHeader(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "NotBearer some-token")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	jwtSvc := NewJWTService("test-secret")
	mw := NewAuthMiddleware(jwtSvc)

	token, _ := jwtSvc.GenerateAccessToken("user-xyz")

	var gotUserID string
	handler := mw.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if gotUserID != "user-xyz" {
		t.Errorf("expected 'user-xyz', got %q", gotUserID)
	}
}
