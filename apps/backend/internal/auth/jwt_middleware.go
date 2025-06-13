package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserClaimsKey is the context key for storing user claims
	UserClaimsKey contextKey = "user_claims"
)

// JWTMiddleware creates HTTP middleware for JWT authentication
func JWTMiddleware(jwtService *JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "Authorization header is required", http.StatusUnauthorized)
				return
			}

			// Check if header starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Authorization header must start with 'Bearer '", http.StatusUnauthorized)
				return
			}

			// Extract token string
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				http.Error(w, "Token is required", http.StatusUnauthorized)
				return
			}

			// Validate token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid token: %v", err), http.StatusUnauthorized)
				return
			}

			// Add claims to request context
			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			r = r.WithContext(ctx)

			// Call next handler
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserClaimsFromContext extracts user claims from the request context
func GetUserClaimsFromContext(ctx context.Context) (*AppClaims, error) {
	claims, ok := ctx.Value(UserClaimsKey).(*AppClaims)
	if !ok {
		return nil, fmt.Errorf("user claims not found in context")
	}
	return claims, nil
}

// GetUserIDFromContext is a convenience function to get user ID from context
func GetUserIDFromContext(ctx context.Context) (string, error) {
	claims, err := GetUserClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	return claims.UserID, nil
}

// GetUsernameFromContext is a convenience function to get username from context
func GetUsernameFromContext(ctx context.Context) (string, error) {
	claims, err := GetUserClaimsFromContext(ctx)
	if err != nil {
		return "", err
	}
	return claims.Username, nil
} 