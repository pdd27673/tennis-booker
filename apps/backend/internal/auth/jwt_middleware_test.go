package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTMiddleware_ValidToken(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)
	
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	// Generate a valid token
	userID := "user123"
	username := "testuser"
	token, err := jwtService.GenerateToken(userID, username, time.Hour)
	require.NoError(t, err)

	// Create a test handler that checks if claims are in context
	var capturedClaims *AppClaims
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, err := GetUserClaimsFromContext(r.Context())
		require.NoError(t, err)
		capturedClaims = claims
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler with middleware
	handler := middleware(testHandler)

	// Create request with valid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, capturedClaims)
	assert.Equal(t, userID, capturedClaims.UserID)
	assert.Equal(t, username, capturedClaims.Username)

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTMiddleware_NoAuthorizationHeader(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware(testHandler)

	// Create request without Authorization header
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is required")
}

func TestJWTMiddleware_InvalidAuthorizationFormat(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware(testHandler)

	// Create request with invalid Authorization header format
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Basic sometoken")
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header must start with 'Bearer '")
}

func TestJWTMiddleware_EmptyToken(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware(testHandler)

	// Create request with empty token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer ")
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Token is required")
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)
	
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware(testHandler)

	// Create request with invalid token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTMiddleware_ExpiredToken(t *testing.T) {
	// Setup
	mockSecretsProvider := &MockJWTSecretsProvider{}
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)
	
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")
	middleware := JWTMiddleware(jwtService)

	// Generate an expired token
	token, err := jwtService.GenerateToken("user123", "testuser", time.Nanosecond)
	require.NoError(t, err)
	
	// Wait for token to expire
	time.Sleep(time.Millisecond)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Handler should not be called")
	})

	handler := middleware(testHandler)

	// Create request with expired token
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	// Execute request
	handler.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid token")

	mockSecretsProvider.AssertExpectations(t)
}

func TestGetUserClaimsFromContext(t *testing.T) {
	// Create test claims
	claims := &AppClaims{
		UserID:   "user123",
		Username: "testuser",
	}

	// Create context with claims
	ctx := context.WithValue(context.Background(), UserClaimsKey, claims)

	// Test successful retrieval
	retrievedClaims, err := GetUserClaimsFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, claims.UserID, retrievedClaims.UserID)
	assert.Equal(t, claims.Username, retrievedClaims.Username)
}

func TestGetUserClaimsFromContext_NotFound(t *testing.T) {
	// Create context without claims
	ctx := context.Background()

	// Test error case
	claims, err := GetUserClaimsFromContext(ctx)
	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "user claims not found in context")
}

func TestGetUserIDFromContext(t *testing.T) {
	// Create test claims
	expectedUserID := "user123"
	claims := &AppClaims{
		UserID:   expectedUserID,
		Username: "testuser",
	}

	// Create context with claims
	ctx := context.WithValue(context.Background(), UserClaimsKey, claims)

	// Test successful retrieval
	userID, err := GetUserIDFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedUserID, userID)
}

func TestGetUserIDFromContext_NotFound(t *testing.T) {
	// Create context without claims
	ctx := context.Background()

	// Test error case
	userID, err := GetUserIDFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, userID)
}

func TestGetUsernameFromContext(t *testing.T) {
	// Create test claims
	expectedUsername := "testuser"
	claims := &AppClaims{
		UserID:   "user123",
		Username: expectedUsername,
	}

	// Create context with claims
	ctx := context.WithValue(context.Background(), UserClaimsKey, claims)

	// Test successful retrieval
	username, err := GetUsernameFromContext(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedUsername, username)
}

func TestGetUsernameFromContext_NotFound(t *testing.T) {
	// Create context without claims
	ctx := context.Background()

	// Test error case
	username, err := GetUsernameFromContext(ctx)
	assert.Error(t, err)
	assert.Empty(t, username)
} 