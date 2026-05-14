package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	svc := NewJWTService("test-secret-key")
	userID := "user-123"

	token, err := svc.GenerateAccessToken(userID)
	if err != nil {
		t.Fatalf("unexpected error generating access token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	gotUserID, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("unexpected error validating access token: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected userID %q, got %q", userID, gotUserID)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	svc := NewJWTService("test-secret-key")
	userID := "user-456"

	token, err := svc.GenerateRefreshToken(userID)
	if err != nil {
		t.Fatalf("unexpected error generating refresh token: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	gotUserID, err := svc.ValidateRefreshToken(token)
	if err != nil {
		t.Fatalf("unexpected error validating refresh token: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("expected userID %q, got %q", userID, gotUserID)
	}
}

func TestAccessTokenCannotBeValidatedAsRefresh(t *testing.T) {
	svc := NewJWTService("test-secret-key")

	token, err := svc.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ValidateRefreshToken(token)
	if err == nil {
		t.Fatal("expected error validating access token as refresh token")
	}
}

func TestRefreshTokenCannotBeValidatedAsAccess(t *testing.T) {
	svc := NewJWTService("test-secret-key")

	token, err := svc.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error validating refresh token as access token")
	}
}

func TestExpiredAccessToken(t *testing.T) {
	svc := &JWTService{
		secret:           []byte("test-secret-key"),
		accessTokenTTL:  -1 * time.Minute, // already expired
		refreshTokenTTL: 7 * 24 * time.Hour,
	}

	token, err := svc.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestExpiredRefreshToken(t *testing.T) {
	svc := &JWTService{
		secret:           []byte("test-secret-key"),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: -1 * time.Minute, // already expired
	}

	token, err := svc.GenerateRefreshToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.ValidateRefreshToken(token)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestWrongSecretKey(t *testing.T) {
	svc1 := NewJWTService("secret-one")
	svc2 := NewJWTService("secret-two")

	token, err := svc1.GenerateAccessToken("user-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error validating token with wrong secret")
	}
}
