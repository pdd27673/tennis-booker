package utils

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse represents a standard API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	})
}

// WriteJSON writes a JSON response with proper headers
func WriteJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(data)
}

// WriteSuccess writes a standardized success response
func WriteSuccess(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, data, http.StatusOK)
}

// WriteCreated writes a standardized created response
func WriteCreated(w http.ResponseWriter, data interface{}) error {
	return WriteJSON(w, data, http.StatusCreated)
}
