package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockSecretsProvider for testing
type MockSecretsProvider struct {
	secret string
}

func (m *MockSecretsProvider) GetJWTSecret() (string, error) {
	return m.secret, nil
}

// MockRefreshTokenService for testing
type MockRefreshTokenService struct {
	tokens map[string]*models.RefreshToken
}

func NewMockRefreshTokenService() *MockRefreshTokenService {
	return &MockRefreshTokenService{
		tokens: make(map[string]*models.RefreshToken),
	}
}

func (m *MockRefreshTokenService) CreateRefreshToken(ctx context.Context, userID primitive.ObjectID, token string, expiresAt time.Time) (*models.RefreshToken, error) {
	refreshToken := &models.RefreshToken{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: token, // For testing, we'll store the token directly
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		Revoked:   false,
	}
	m.tokens[token] = refreshToken
	return refreshToken, nil
}

func (m *MockRefreshTokenService) ValidateRefreshToken(ctx context.Context, token string) (*models.RefreshToken, error) {
	refreshToken, exists := m.tokens[token]
	if !exists || refreshToken.Revoked || refreshToken.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}
	return refreshToken, nil
}

func (m *MockRefreshTokenService) RevokeRefreshToken(ctx context.Context, token string) error {
	refreshToken, exists := m.tokens[token]
	if !exists {
		return fmt.Errorf("refresh token not found")
	}
	refreshToken.Revoked = true
	now := time.Now()
	refreshToken.RevokedAt = &now
	return nil
}

func (m *MockRefreshTokenService) RevokeAllUserTokens(ctx context.Context, userID primitive.ObjectID) error {
	now := time.Now()
	for _, token := range m.tokens {
		if token.UserID == userID && !token.Revoked {
			token.Revoked = true
			token.RevokedAt = &now
		}
	}
	return nil
}

func (m *MockRefreshTokenService) CleanupExpiredTokens(ctx context.Context) error {
	for token, refreshToken := range m.tokens {
		if refreshToken.ExpiresAt.Before(time.Now()) || 
		   (refreshToken.Revoked && refreshToken.RevokedAt != nil && refreshToken.RevokedAt.Before(time.Now().AddDate(0, 0, -30))) {
			delete(m.tokens, token)
		}
	}
	return nil
}

func setupTestAuthHandler() (*AuthHandler, *auth.JWTService, *MockRefreshTokenService) {
	// Create user service with fast password hashing for tests
	userService := models.NewInMemoryUserServiceWithPasswordService(
		auth.NewBcryptPasswordServiceWithCost(4),
	)

	// Create JWT service with mock secrets provider
	secretsProvider := &MockSecretsProvider{secret: "test-secret-key-for-testing"}
	jwtService := auth.NewJWTService(secretsProvider, "test-issuer")

	// Create mock refresh token service
	refreshTokenService := NewMockRefreshTokenService()

	// Create auth handler
	authHandler := NewAuthHandler(userService, jwtService, refreshTokenService)

	return authHandler, jwtService, refreshTokenService
}

func TestAuthHandler_Register(t *testing.T) {
	authHandler, _, _ := setupTestAuthHandler()

	t.Run("successful registration", func(t *testing.T) {
		reqBody := RegisterRequest{
			Username: "testuser",
			Email:    "test@example.com",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authHandler.Register(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AuthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotNil(t, response.User)
		assert.Equal(t, "testuser", response.User.Username)
		assert.Equal(t, "test@example.com", response.User.Email)
		assert.NotEmpty(t, response.User.ID)
	})

	t.Run("duplicate username", func(t *testing.T) {
		// First registration
		reqBody := RegisterRequest{
			Username: "duplicate",
			Email:    "first@example.com",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		authHandler.Register(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Second registration with same username
		reqBody.Email = "second@example.com"
		body, _ = json.Marshal(reqBody)

		req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w = httptest.NewRecorder()
		authHandler.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)

		var errorResp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "User already exists", errorResp.Message)
	})

	t.Run("duplicate email", func(t *testing.T) {
		// First registration
		reqBody := RegisterRequest{
			Username: "user1",
			Email:    "duplicate@example.com",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		authHandler.Register(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		// Second registration with same email
		reqBody.Username = "user2"
		body, _ = json.Marshal(reqBody)

		req = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		w = httptest.NewRecorder()
		authHandler.Register(w, req)

		assert.Equal(t, http.StatusConflict, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		authHandler.Register(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation errors", func(t *testing.T) {
		testCases := []struct {
			name     string
			request  RegisterRequest
			expected string
		}{
			{
				name:     "empty username",
				request:  RegisterRequest{Username: "", Email: "test@example.com", Password: "DEMO_PASSWORD"},
				expected: "username is required",
			},
			{
				name:     "empty email",
				request:  RegisterRequest{Username: "testuser", Email: "", Password: "DEMO_PASSWORD"},
				expected: "email is required",
			},
			{
				name:     "empty password",
				request:  RegisterRequest{Username: "testuser", Email: "test@example.com", Password: ""},
				expected: "password is required",
			},
			{
				name:     "short username",
				request:  RegisterRequest{Username: "ab", Email: "test@example.com", Password: "DEMO_PASSWORD"},
				expected: "username must be at least 3 characters long",
			},
			{
				name:     "short password",
				request:  RegisterRequest{Username: "testuser", Email: "test@example.com", Password: "12345"},
				expected: "password must be at least 6 characters long",
			},
			{
				name:     "invalid email",
				request:  RegisterRequest{Username: "testuser", Email: "invalid-email", Password: "DEMO_PASSWORD"},
				expected: "invalid email format",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				body, _ := json.Marshal(tc.request)
				req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
				w := httptest.NewRecorder()

				authHandler.Register(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var errorResp ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, errorResp.Message)
			})
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/register", nil)
		w := httptest.NewRecorder()

		authHandler.Register(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestAuthHandler_Login(t *testing.T) {
	authHandler, _, _ := setupTestAuthHandler()

	// Create a test user first
	ctx := context.Background()
	userService := authHandler.userService
	testUser, err := userService.CreateUser(ctx, "loginuser", "login@example.com", "DEMO_PASSWORD")
	require.NoError(t, err)

	t.Run("successful login", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "loginuser",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authHandler.Login(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response AuthResponse
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)

		assert.NotEmpty(t, response.Token)
		assert.NotEmpty(t, response.RefreshToken)
		assert.NotNil(t, response.User)
		assert.Equal(t, "loginuser", response.User.Username)
		assert.Equal(t, "login@example.com", response.User.Email)
		assert.Equal(t, testUser.ID.Hex(), response.User.ID)
	})

	t.Run("invalid credentials - wrong password", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "loginuser",
			Password: "wrongpassword",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid credentials", errorResp.Message)
	})

	t.Run("invalid credentials - user not found", func(t *testing.T) {
		reqBody := LoginRequest{
			Username: "nonexistent",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.Login(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		authHandler.Login(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("validation errors", func(t *testing.T) {
		testCases := []struct {
			name     string
			request  LoginRequest
			expected string
		}{
			{
				name:     "empty username",
				request:  LoginRequest{Username: "", Password: "DEMO_PASSWORD"},
				expected: "username is required",
			},
			{
				name:     "empty password",
				request:  LoginRequest{Username: "testuser", Password: ""},
				expected: "password is required",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				body, _ := json.Marshal(tc.request)
				req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
				w := httptest.NewRecorder()

				authHandler.Login(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)

				var errorResp ErrorResponse
				err := json.NewDecoder(w.Body).Decode(&errorResp)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, errorResp.Message)
			})
		}
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		w := httptest.NewRecorder()

		authHandler.Login(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestAuthHandler_Me(t *testing.T) {
	authHandler, jwtService, _ := setupTestAuthHandler()

	// Create a test user first
	ctx := context.Background()
	userService := authHandler.userService
	testUser, err := userService.CreateUser(ctx, "meuser", "me@example.com", "DEMO_PASSWORD")
	require.NoError(t, err)

	t.Run("successful me request", func(t *testing.T) {
		// Generate a valid token
		token, err := jwtService.GenerateToken(testUser.ID.Hex(), testUser.Username, 15*time.Minute)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		// Add user claims to context (normally done by JWT middleware)
		claims := &auth.AppClaims{
			UserID:   testUser.ID.Hex(),
			Username: testUser.Username,
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		authHandler.Me(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var userInfo UserInfo
		err = json.NewDecoder(w.Body).Decode(&userInfo)
		require.NoError(t, err)

		assert.Equal(t, testUser.ID.Hex(), userInfo.ID)
		assert.Equal(t, "meuser", userInfo.Username)
		assert.Equal(t, "me@example.com", userInfo.Email)
	})

	t.Run("missing user claims", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
		w := httptest.NewRecorder()

		authHandler.Me(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("user not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

		// Add claims for non-existent user
		claims := &auth.AppClaims{
			UserID:   primitive.NewObjectID().Hex(),
			Username: "nonexistent",
		}
		ctx := auth.SetUserClaimsInContext(req.Context(), claims)
		req = req.WithContext(ctx)

		w := httptest.NewRecorder()
		authHandler.Me(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/me", nil)
		w := httptest.NewRecorder()

		authHandler.Me(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	authHandler, jwtService, refreshTokenService := setupTestAuthHandler()

	// Create a test user first
	ctx := context.Background()
	userService := authHandler.userService
	testUser, err := userService.CreateUser(ctx, "refreshuser", "refresh@example.com", "DEMO_PASSWORD")
	require.NoError(t, err)

	t.Run("successful token refresh", func(t *testing.T) {
		// Generate a refresh token and store it
		refreshToken, err := jwtService.GenerateRefreshToken(testUser.ID.Hex(), testUser.Username)
		require.NoError(t, err)

		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		_, err = refreshTokenService.CreateRefreshToken(ctx, testUser.ID, refreshToken, expiresAt)
		require.NoError(t, err)

		reqBody := RefreshRequest{
			RefreshToken: refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var response struct {
			Token string `json:"token"`
		}
		err = json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.NotEmpty(t, response.Token)

		// Verify the new token is valid
		claims, err := jwtService.ValidateToken(response.Token)
		require.NoError(t, err)
		assert.Equal(t, testUser.ID.Hex(), claims.UserID)
		assert.Equal(t, testUser.Username, claims.Username)
	})

	t.Run("invalid refresh token", func(t *testing.T) {
		reqBody := RefreshRequest{
			RefreshToken: "invalid-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var errorResp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Invalid or expired refresh token", errorResp.Message)
	})

	t.Run("expired refresh token", func(t *testing.T) {
		// Generate a refresh token and store it with past expiration
		refreshToken, err := jwtService.GenerateRefreshToken(testUser.ID.Hex(), testUser.Username)
		require.NoError(t, err)

		expiresAt := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago
		_, err = refreshTokenService.CreateRefreshToken(ctx, testUser.ID, refreshToken, expiresAt)
		require.NoError(t, err)

		reqBody := RefreshRequest{
			RefreshToken: refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("missing refresh token", func(t *testing.T) {
		reqBody := RefreshRequest{
			RefreshToken: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/refresh", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/refresh", nil)
		w := httptest.NewRecorder()

		authHandler.RefreshToken(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestAuthHandler_Logout(t *testing.T) {
	authHandler, jwtService, refreshTokenService := setupTestAuthHandler()

	// Create a test user first
	ctx := context.Background()
	userService := authHandler.userService
	testUser, err := userService.CreateUser(ctx, "logoutuser", "logout@example.com", "DEMO_PASSWORD")
	require.NoError(t, err)

	t.Run("successful logout", func(t *testing.T) {
		// Generate a refresh token and store it
		refreshToken, err := jwtService.GenerateRefreshToken(testUser.ID.Hex(), testUser.Username)
		require.NoError(t, err)

		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		_, err = refreshTokenService.CreateRefreshToken(ctx, testUser.ID, refreshToken, expiresAt)
		require.NoError(t, err)

		// Verify token is valid before logout
		_, err = refreshTokenService.ValidateRefreshToken(ctx, refreshToken)
		require.NoError(t, err)

		reqBody := LogoutRequest{
			RefreshToken: refreshToken,
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authHandler.Logout(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify token is no longer valid after logout
		_, err = refreshTokenService.ValidateRefreshToken(ctx, refreshToken)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid or expired refresh token")
	})

	t.Run("logout with non-existent token", func(t *testing.T) {
		reqBody := LogoutRequest{
			RefreshToken: "non-existent-token",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.Logout(w, req)

		// Should return success even for non-existent tokens (security best practice)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing refresh token", func(t *testing.T) {
		reqBody := LogoutRequest{
			RefreshToken: "",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		authHandler.Logout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var errorResp ErrorResponse
		err := json.NewDecoder(w.Body).Decode(&errorResp)
		require.NoError(t, err)
		assert.Equal(t, "Refresh token is required", errorResp.Message)
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/auth/logout", strings.NewReader("invalid json"))
		w := httptest.NewRecorder()

		authHandler.Logout(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/auth/logout", nil)
		w := httptest.NewRecorder()

		authHandler.Logout(w, req)

		assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	})
}

func TestAuthHandler_Integration(t *testing.T) {
	authHandler, _, refreshTokenService := setupTestAuthHandler()

	t.Run("complete authentication flow", func(t *testing.T) {
		// 1. Register a new user
		registerReq := RegisterRequest{
			Username: "integrationuser",
			Email:    "integration@example.com",
			Password: "DEMO_PASSWORD",
		}
		body, _ := json.Marshal(registerReq)

		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		authHandler.Register(w, req)
		assert.Equal(t, http.StatusCreated, w.Code)

		var registerResp AuthResponse
		err := json.NewDecoder(w.Body).Decode(&registerResp)
		require.NoError(t, err)

		// 2. Login with the same user
		loginReq := LoginRequest{
			Username: "integrationuser",
			Password: "DEMO_PASSWORD",
		}
		body, _ = json.Marshal(loginReq)

		req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		authHandler.Login(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var loginResp AuthResponse
		err = json.NewDecoder(w.Body).Decode(&loginResp)
		require.NoError(t, err)

		// 3. Use refresh token to get new access token
		refreshReq := RefreshRequest{
			RefreshToken: loginResp.RefreshToken,
		}
		body, _ = json.Marshal(refreshReq)

		req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		authHandler.RefreshToken(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var refreshResp struct {
			Token string `json:"token"`
		}
		err = json.NewDecoder(w.Body).Decode(&refreshResp)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.Token)

		// 4. Logout to invalidate refresh token
		logoutReq := LogoutRequest{
			RefreshToken: loginResp.RefreshToken,
		}
		body, _ = json.Marshal(logoutReq)

		req = httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		authHandler.Logout(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		// 5. Try to use refresh token again (should fail)
		req = httptest.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()

		authHandler.RefreshToken(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// 6. Verify refresh token is revoked in storage
		_, err = refreshTokenService.ValidateRefreshToken(context.Background(), loginResp.RefreshToken)
		assert.Error(t, err)
	})
} 