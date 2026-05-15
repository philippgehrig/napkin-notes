package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrTokenExpired = errors.New("token expired")
	ErrWrongType    = errors.New("wrong token type")
)

// JWTService handles generation and validation of JWT tokens.
type JWTService struct {
	secret           []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewJWTService creates a new JWTService with the given signing secret.
func NewJWTService(secret string) *JWTService {
	return &JWTService{
		secret:           []byte(secret),
		accessTokenTTL:  15 * time.Minute,
		refreshTokenTTL: 7 * 24 * time.Hour,
	}
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
}

// GenerateAccessToken creates a short-lived access token for the given userID.
func (s *JWTService) GenerateAccessToken(userID string) (string, error) {
	return s.generateToken(userID, "access", s.accessTokenTTL)
}

// GenerateRefreshToken creates a long-lived refresh token for the given userID.
func (s *JWTService) GenerateRefreshToken(userID string) (string, error) {
	return s.generateToken(userID, "refresh", s.refreshTokenTTL)
}

// ValidateAccessToken validates a token and ensures it is an access token.
func (s *JWTService) ValidateAccessToken(tokenString string) (string, error) {
	return s.validateToken(tokenString, "access")
}

// ValidateRefreshToken validates a token and ensures it is a refresh token.
func (s *JWTService) ValidateRefreshToken(tokenString string) (string, error) {
	return s.validateToken(tokenString, "refresh")
}

func (s *JWTService) generateToken(userID, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
		UserID:    userID,
		TokenType: tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

func (s *JWTService) validateToken(tokenString, expectedType string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})
	if err != nil {
		return "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*tokenClaims)
	if !ok || !token.Valid {
		return "", ErrInvalidToken
	}

	if claims.TokenType != expectedType {
		return "", ErrWrongType
	}

	return claims.UserID, nil
}
