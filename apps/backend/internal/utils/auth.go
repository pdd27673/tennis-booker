package utils

import (
	"net/http"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"tennis-booker/internal/auth"
)

// GetUserFromContext extracts user ID from request context
func GetUserFromContext(r *http.Request) (primitive.ObjectID, error) {
	userIDStr, err := auth.GetUserIDFromContext(r.Context())
	if err != nil {
		return primitive.ObjectID{}, err
	}
	return primitive.ObjectIDFromHex(userIDStr)
}

// RequireAuth validates authentication and returns user ID
func RequireAuth(w http.ResponseWriter, r *http.Request) (primitive.ObjectID, bool) {
	userID, err := GetUserFromContext(r)
	if err != nil {
		WriteError(w, "Authentication required", http.StatusUnauthorized)
		return primitive.ObjectID{}, false
	}
	return userID, true
}
