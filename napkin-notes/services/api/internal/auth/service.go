package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/philippgehrig/napkin-notes/services/api/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailTaken         = errors.New("email already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository defines persistence operations for users.
type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

// AuthResult is the response returned by auth operations.
type AuthResult struct {
	AccessToken  string
	RefreshToken string
	User         *models.User
}

// AuthService handles user registration, login, and token refresh.
type AuthService struct {
	repo   UserRepository
	jwtSvc *JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(repo UserRepository, jwtSvc *JWTService) *AuthService {
	return &AuthService{
		repo:   repo,
		jwtSvc: jwtSvc,
	}
}

// Register creates a new user and returns auth tokens.
func (s *AuthService) Register(ctx context.Context, email, password, displayName string) (*AuthResult, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: string(hash),
		DisplayName:  displayName,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

// Login authenticates a user by email and password and returns tokens.
func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.generateTokens(user)
}

// RefreshAccessToken validates a refresh token and returns new tokens.
func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	userID, err := s.jwtSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.generateTokens(user)
}

func (s *AuthService) generateTokens(user *models.User) (*AuthResult, error) {
	accessToken, err := s.jwtSvc.GenerateAccessToken(user.ID)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtSvc.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, err
	}

	return &AuthResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}, nil
}
