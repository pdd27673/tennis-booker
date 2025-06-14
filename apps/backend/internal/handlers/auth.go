package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	userService         models.UserService
	jwtService          *auth.JWTService
	refreshTokenService models.RefreshTokenService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(userService models.UserService, jwtService *auth.JWTService, refreshTokenService models.RefreshTokenService) *AuthHandler {
	return &AuthHandler{
		userService:         userService,
		jwtService:          jwtService,
		refreshTokenService: refreshTokenService,
	}
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RefreshRequest represents the request body for token refresh
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents the request body for logout
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the response for successful authentication
type AuthResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	User         *UserInfo `json:"user"`
}

// UserInfo represents user information returned in responses
type UserInfo struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Phone    string `json:"phone,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := h.validateRegisterRequest(&req); err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Create user
	user, err := h.userService.CreateUser(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		if strings.Contains(err.Error(), "user already exists") {
			h.writeErrorResponse(w, "User already exists", http.StatusConflict)
			return
		}
		h.writeErrorResponse(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokenPair(r.Context(), user)
	if err != nil {
		h.writeErrorResponse(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			Phone:    user.Phone,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	if err := h.validateLoginRequest(&req); err != nil {
		h.writeErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Find user by username
	user, err := h.userService.FindByUsername(r.Context(), req.Username)
	if err != nil {
		h.writeErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := h.userService.VerifyPassword(r.Context(), user, req.Password); err != nil {
		h.writeErrorResponse(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokenPair(r.Context(), user)
	if err != nil {
		h.writeErrorResponse(w, "Failed to generate tokens", http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
		User: &UserInfo{
			ID:       user.ID.Hex(),
			Username: user.Username,
			Email:    user.Email,
			Name:     user.Name,
			Phone:    user.Phone,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Me handles the protected endpoint to get current user information
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user claims from context (set by JWT middleware)
	claims, err := auth.GetUserClaimsFromContext(r.Context())
	if err != nil {
		h.writeErrorResponse(w, "User not authenticated", http.StatusUnauthorized)
		return
	}

	// Get full user information
	user, err := h.userService.FindByID(r.Context(), claims.UserID)
	if err != nil {
		h.writeErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Prepare response
	userInfo := UserInfo{
		ID:       user.ID.Hex(),
		Username: user.Username,
		Email:    user.Email,
		Name:     user.Name,
		Phone:    user.Phone,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userInfo)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		h.writeErrorResponse(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Validate refresh token in database
	storedToken, err := h.refreshTokenService.ValidateRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		h.writeErrorResponse(w, "Invalid or expired refresh token", http.StatusUnauthorized)
		return
	}

	// Get user information
	user, err := h.userService.FindByID(r.Context(), storedToken.UserID.Hex())
	if err != nil {
		h.writeErrorResponse(w, "User not found", http.StatusNotFound)
		return
	}

	// Generate new access token
	newAccessToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Username, 15*time.Minute)
	if err != nil {
		h.writeErrorResponse(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	response := struct {
		Token string `json:"token"`
	}{
		Token: newAccessToken,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout handles user logout by invalidating the refresh token
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.RefreshToken == "" {
		h.writeErrorResponse(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	// Revoke the refresh token
	err := h.refreshTokenService.RevokeRefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		// Even if the token doesn't exist, we return success for security
		// This prevents information leakage about token existence
		if strings.Contains(err.Error(), "refresh token not found") {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.writeErrorResponse(w, "Failed to logout", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// generateTokenPair generates both access and refresh tokens and stores the refresh token
func (h *AuthHandler) generateTokenPair(ctx context.Context, user *models.User) (string, string, error) {
	// Generate access token (15 minutes)
	accessToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Username, 15*time.Minute)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (7 days)
	refreshToken, err := h.jwtService.GenerateRefreshToken(user.ID.Hex(), user.Username)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in database
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	_, err = h.refreshTokenService.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return "", "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// validateRegisterRequest validates the registration request
func (h *AuthHandler) validateRegisterRequest(req *RegisterRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	if len(req.Username) < 3 {
		return fmt.Errorf("username must be at least 3 characters long")
	}
	if len(req.Password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	if !strings.Contains(req.Email, "@") {
		return fmt.Errorf("invalid email format")
	}
	return nil
}

// validateLoginRequest validates the login request
func (h *AuthHandler) validateLoginRequest(req *LoginRequest) error {
	if req.Username == "" {
		return fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

// writeErrorResponse writes an error response in JSON format
func (h *AuthHandler) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	json.NewEncoder(w).Encode(errorResp)
}
