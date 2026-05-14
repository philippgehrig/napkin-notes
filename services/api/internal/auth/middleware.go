package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const userIDKey contextKey = "userID"

// AuthMiddleware provides HTTP middleware for JWT authentication.
type AuthMiddleware struct {
	jwtSvc *JWTService
}

// NewAuthMiddleware creates a new AuthMiddleware.
func NewAuthMiddleware(jwtSvc *JWTService) *AuthMiddleware {
	return &AuthMiddleware{jwtSvc: jwtSvc}
}

// Authenticate is middleware that validates the Bearer token and injects userID into context.
func (m *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, `{"error":"invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		userID, err := m.jwtSvc.ValidateAccessToken(parts[1])
		if err != nil {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserID retrieves the authenticated user ID from the request context.
func GetUserID(ctx context.Context) string {
	userID, _ := ctx.Value(userIDKey).(string)
	return userID
}

// UserIDKeyForTest returns the context key used for user ID storage.
// This is intended for use in tests that need to inject a user ID into context.
func UserIDKeyForTest() contextKey {
	return userIDKey
}
