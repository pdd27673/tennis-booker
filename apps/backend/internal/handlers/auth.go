package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"tennis-booker/internal/auth"
	"tennis-booker/internal/database"
	"tennis-booker/internal/models"
	"tennis-booker/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	jwtService *auth.JWTService
	db         database.Database
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(jwtService *auth.JWTService, db database.Database) *AuthHandler {
	return &AuthHandler{
		jwtService: jwtService,
		db:         db,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         models.User `json:"user"`
}

// RefreshRequest represents a token refresh request
type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Find user by email
	ctx, cancel := utils.WithDBTimeout()
	defer cancel()

	collection := h.db.Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(req.Password)); err != nil {
		utils.WriteError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Update last login
	update := bson.M{
		"$set": bson.M{
			"lastLogin": time.Now(),
			"updatedAt": time.Now(),
		},
	}
	_, err = collection.UpdateOne(ctx, bson.M{"_id": user.ID}, update)
	if err != nil {
		// Log error but don't fail the login since token generation succeeded
		// In production, you might want to log this error
	}

	// Don't return password in response
	user.HashedPassword = ""

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}

	utils.WriteSuccess(w, response)
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx, cancel := utils.WithDBTimeout()
	defer cancel()

	collection := h.db.Collection("users")

	// Check if user already exists
	var existingUser models.User
	err := collection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&existingUser)
	if err == nil {
		http.Error(w, "User with this email already exists", http.StatusConflict)
		return
	} else if err != mongo.ErrNoDocuments {
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Create new user
	user := models.User{
		ID:             primitive.NewObjectID(),
		Email:          req.Email,
		Username:       req.Email, // Use email as username for now
		HashedPassword: string(hashedPassword),
		Name:           req.FirstName + " " + req.LastName,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert user into database
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Generate tokens
	accessToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	refreshToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Don't return password in response
	user.HashedPassword = ""

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         user,
	}

	utils.WriteCreated(w, response)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate refresh token
	claims, err := h.jwtService.ValidateToken(req.RefreshToken)
	if err != nil {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	// Find user to ensure they still exist
	ctx, cancel := utils.WithDBTimeout()
	defer cancel()

	userID, err := primitive.ObjectIDFromHex(claims.UserID)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	collection := h.db.Collection("users")
	var user models.User
	err = collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Generate new access token
	accessToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate access token", http.StatusInternalServerError)
		return
	}

	// Generate new refresh token
	newRefreshToken, err := h.jwtService.GenerateToken(user.ID.Hex(), user.Email, 7*24*time.Hour)
	if err != nil {
		http.Error(w, "Failed to generate refresh token", http.StatusInternalServerError)
		return
	}

	// Don't return password in response
	user.HashedPassword = ""

	response := AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         user,
	}

	utils.WriteSuccess(w, response)
}

// GetCurrentUser returns the current authenticated user
func (h *AuthHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context using utility function
	userID, ok := utils.RequireAuth(w, r)
	if !ok {
		return
	}

	ctx, cancel := utils.WithDBTimeout()
	defer cancel()

	collection := h.db.Collection("users")
	var user models.User
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			utils.WriteError(w, "User not found", http.StatusNotFound)
			return
		}
		utils.WriteError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Don't return password
	user.HashedPassword = ""

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// Logout handles user logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// For JWT, logout is typically handled client-side by discarding the token
	// For more secure implementations, you'd maintain a blacklist of tokens
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
