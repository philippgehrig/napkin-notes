package auth

import (
	"encoding/json"
	"errors"
	"net/http"
)

// AuthHandler handles HTTP requests for authentication endpoints.
type AuthHandler struct {
	authSvc *AuthService
}

// NewAuthHandler creates a new AuthHandler.
func NewAuthHandler(authSvc *AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
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

type authResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         userResponse `json:"user"`
}

type userResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			writeError(w, "email already taken", http.StatusConflict)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeAuthResponse(w, result, http.StatusCreated)
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, "email and password are required", http.StatusBadRequest)
		return
	}

	result, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			writeError(w, "invalid credentials", http.StatusUnauthorized)
			return
		}
		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeAuthResponse(w, result, http.StatusOK)
}

// Refresh handles POST /api/auth/refresh.
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		writeError(w, "refresh_token is required", http.StatusBadRequest)
		return
	}

	result, err := h.authSvc.RefreshAccessToken(r.Context(), req.RefreshToken)
	if err != nil {
		writeError(w, "invalid refresh token", http.StatusUnauthorized)
		return
	}

	writeAuthResponse(w, result, http.StatusOK)
}

func writeAuthResponse(w http.ResponseWriter, result *AuthResult, status int) {
	resp := authResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
		User: userResponse{
			ID:          result.User.ID,
			Email:       result.User.Email,
			DisplayName: result.User.DisplayName,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
