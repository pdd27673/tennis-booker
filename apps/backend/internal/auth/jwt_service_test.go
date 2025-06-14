package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockJWTSecretsProvider is a mock implementation of the JWTSecretsProvider interface
type MockJWTSecretsProvider struct {
	mock.Mock
}

func (m *MockJWTSecretsProvider) GetJWTSecret() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func TestNewJWTService(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	issuer := "tennis-booker"

	jwtService := NewJWTService(mockSecretsProvider, issuer)

	assert.NotNil(t, jwtService)
	assert.Equal(t, issuer, jwtService.issuer)
	assert.Equal(t, mockSecretsProvider, jwtService.secretsProvider)
}

func TestJWTService_GenerateToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	userID := "user123"
	username := "testuser"
	duration := time.Hour

	token, err := jwtService.GenerateToken(userID, username, duration)

	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Verify the token can be parsed and contains correct claims
	parsedToken, err := jwt.ParseWithClaims(token, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	claims, ok := parsedToken.Claims.(*AppClaims)
	require.True(t, ok)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "tennis-booker", claims.Issuer)
	assert.Equal(t, userID, claims.Subject)

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_GenerateToken_VaultError(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call to return an error
	mockSecretsProvider.On("GetJWTSecret").Return("", assert.AnError)

	token, err := jwtService.GenerateToken("user123", "testuser", time.Hour)

	assert.Error(t, err)
	assert.Empty(t, token)
	assert.Contains(t, err.Error(), "failed to fetch JWT secret from Vault")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_ValidateToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call for both generation and validation
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	userID := "user123"
	username := "testuser"
	duration := time.Hour

	// Generate a token first
	token, err := jwtService.GenerateToken(userID, username, duration)
	require.NoError(t, err)

	// Now validate the token
	claims, err := jwtService.ValidateToken(token)

	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)
	assert.Equal(t, "tennis-booker", claims.Issuer)

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	// Try to validate an invalid token
	claims, err := jwtService.ValidateToken("invalid.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse JWT token")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_ValidateToken_ExpiredToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	// Generate a token with very short expiration
	token, err := jwtService.GenerateToken("user123", "testuser", time.Nanosecond)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(time.Millisecond)

	// Try to validate the expired token
	claims, err := jwtService.ValidateToken(token)

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to parse JWT token")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_ValidateToken_VaultError(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call to return an error
	mockSecretsProvider.On("GetJWTSecret").Return("", assert.AnError)

	claims, err := jwtService.ValidateToken("some.token.here")

	assert.Error(t, err)
	assert.Nil(t, claims)
	assert.Contains(t, err.Error(), "failed to fetch JWT secret from Vault")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_GenerateRefreshToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	userID := "user123"
	username := "testuser"

	refreshToken, err := jwtService.GenerateRefreshToken(userID, username)

	require.NoError(t, err)
	assert.NotEmpty(t, refreshToken)

	// Verify the refresh token has longer expiration (7 days)
	parsedToken, err := jwt.ParseWithClaims(refreshToken, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	claims, ok := parsedToken.Claims.(*AppClaims)
	require.True(t, ok)

	// Check that expiration is approximately 7 days from now
	expectedExpiry := time.Now().Add(7 * 24 * time.Hour)
	actualExpiry := claims.ExpiresAt.Time
	timeDiff := actualExpiry.Sub(expectedExpiry)
	assert.True(t, timeDiff < time.Minute && timeDiff > -time.Minute, "Expiry should be approximately 7 days")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_RefreshAccessToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call for both generation and validation
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	userID := "user123"
	username := "testuser"

	// Generate a refresh token first
	refreshToken, err := jwtService.GenerateRefreshToken(userID, username)
	require.NoError(t, err)

	// Generate new access token from refresh token
	accessTokenDuration := time.Hour
	newAccessToken, err := jwtService.RefreshAccessToken(refreshToken, accessTokenDuration)

	require.NoError(t, err)
	assert.NotEmpty(t, newAccessToken)

	// Verify the new access token contains correct claims
	claims, err := jwtService.ValidateToken(newAccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, username, claims.Username)

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_RefreshAccessToken_InvalidRefreshToken(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	// Try to refresh with invalid token
	newAccessToken, err := jwtService.RefreshAccessToken("invalid.refresh.token", time.Hour)

	assert.Error(t, err)
	assert.Empty(t, newAccessToken)
	assert.Contains(t, err.Error(), "invalid refresh token")

	mockSecretsProvider.AssertExpectations(t)
}

func TestJWTService_TokenSigningMethod(t *testing.T) {
	mockSecretsProvider := &MockJWTSecretsProvider{}
	jwtService := NewJWTService(mockSecretsProvider, "tennis-booker")

	// Mock the GetJWTSecret call
	mockSecretsProvider.On("GetJWTSecret").Return("test-secret-key", nil)

	token, err := jwtService.GenerateToken("user123", "testuser", time.Hour)
	require.NoError(t, err)

	// Parse token to check signing method
	parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Verify it's using HMAC-SHA256
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			t.Errorf("Expected HMAC signing method, got %v", token.Header["alg"])
		}
		assert.Equal(t, "HS256", token.Header["alg"])
		return []byte("test-secret-key"), nil
	})

	require.NoError(t, err)
	assert.True(t, parsedToken.Valid)

	mockSecretsProvider.AssertExpectations(t)
}
