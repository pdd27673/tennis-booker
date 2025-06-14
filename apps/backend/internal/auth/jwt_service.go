package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTSecretsProvider defines the interface for fetching JWT secrets
type JWTSecretsProvider interface {
	GetJWTSecret() (string, error)
}

// JWTService handles JWT token generation and validation using secrets from Vault
type JWTService struct {
	secretsProvider JWTSecretsProvider
	issuer          string
}

// AppClaims represents the custom claims for our application
type AppClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service with the provided secrets provider
func NewJWTService(secretsProvider JWTSecretsProvider, issuer string) *JWTService {
	return &JWTService{
		secretsProvider: secretsProvider,
		issuer:          issuer,
	}
}

// GenerateToken generates a new JWT token for the given user
func (js *JWTService) GenerateToken(userID, username string, expirationDuration time.Duration) (string, error) {
	// Fetch JWT secret from Vault
	jwtSecret, err := js.secretsProvider.GetJWTSecret()
	if err != nil {
		return "", fmt.Errorf("failed to fetch JWT secret from Vault: %w", err)
	}

	// Create claims
	claims := AppClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    js.issuer,
			Subject:   userID,
		},
	}

	// Create token with claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign token with secret from Vault
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT token: %w", err)
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func (js *JWTService) ValidateToken(tokenString string) (*AppClaims, error) {
	// Fetch JWT secret from Vault
	jwtSecret, err := js.secretsProvider.GetJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWT secret from Vault: %w", err)
	}

	// Parse token with claims
	token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Extract and validate claims
	if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid JWT token or claims")
}

// GenerateRefreshToken generates a refresh token with longer expiration
func (js *JWTService) GenerateRefreshToken(userID, username string) (string, error) {
	// Refresh tokens typically have longer expiration (e.g., 7 days)
	return js.GenerateToken(userID, username, 7*24*time.Hour)
}

// RefreshAccessToken generates a new access token from a valid refresh token
func (js *JWTService) RefreshAccessToken(refreshToken string, accessTokenDuration time.Duration) (string, error) {
	// Validate the refresh token
	claims, err := js.ValidateToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Generate new access token with the same user info
	return js.GenerateToken(claims.UserID, claims.Username, accessTokenDuration)
}
