package models

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"tennis-booker/internal/auth"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a user in the system
type User struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Username        string             `bson:"username" json:"username"`
	Email           string             `bson:"email" json:"email"`
	HashedPassword  string             `bson:"hashed_password" json:"-"` // Never expose in JSON
	Name            string             `bson:"name" json:"name"`
	Phone           string             `bson:"phone,omitempty" json:"phone,omitempty"`
	PreferredCourts []string           `bson:"preferred_courts,omitempty" json:"preferred_courts,omitempty"`
	PreferredDays   []string           `bson:"preferred_days,omitempty" json:"preferred_days,omitempty"`
	PreferredTimes  []TimeRange        `bson:"preferred_times,omitempty" json:"preferred_times,omitempty"`
	NotifyBy        []string           `bson:"notify_by,omitempty" json:"notify_by,omitempty"` // "email", "sms"
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at" json:"updated_at"`
}

// TimeRange represents a time range for preferred booking times
type TimeRange struct {
	Start string `bson:"start" json:"start"` // Format: "HH:MM" in 24-hour format
	End   string `bson:"end" json:"end"`     // Format: "HH:MM" in 24-hour format
}

// UserService defines the interface for user operations
type UserService interface {
	CreateUser(ctx context.Context, username, email, password string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	VerifyPassword(ctx context.Context, user *User, password string) error
}

// Common errors
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrInvalidInput      = errors.New("invalid input")
)

// InMemoryUserService provides an in-memory implementation of UserService for development/testing
type InMemoryUserService struct {
	users           map[string]*User // key: username
	usersByEmail    map[string]*User // key: email
	usersById       map[string]*User // key: id
	passwordService auth.PasswordService
	mu              sync.RWMutex
}

// NewInMemoryUserService creates a new in-memory user service
func NewInMemoryUserService() *InMemoryUserService {
	return &InMemoryUserService{
		users:           make(map[string]*User),
		usersByEmail:    make(map[string]*User),
		usersById:       make(map[string]*User),
		passwordService: auth.NewBcryptPasswordService(),
	}
}

// NewInMemoryUserServiceWithPasswordService creates a new in-memory user service with custom password service
func NewInMemoryUserServiceWithPasswordService(passwordService auth.PasswordService) *InMemoryUserService {
	return &InMemoryUserService{
		users:           make(map[string]*User),
		usersByEmail:    make(map[string]*User),
		usersById:       make(map[string]*User),
		passwordService: passwordService,
	}
}

// CreateUser creates a new user with the given credentials
func (s *InMemoryUserService) CreateUser(ctx context.Context, username, email, password string) (*User, error) {
	if username == "" || email == "" || password == "" {
		return nil, ErrInvalidInput
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists by username
	if _, exists := s.users[username]; exists {
		return nil, fmt.Errorf("username '%s': %w", username, ErrUserAlreadyExists)
	}

	// Check if user already exists by email
	if _, exists := s.usersByEmail[email]; exists {
		return nil, fmt.Errorf("email '%s': %w", email, ErrUserAlreadyExists)
	}

	// Hash the password
	hashedPassword, err := s.passwordService.HashPassword(ctx, password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create new user
	now := time.Now()
	user := &User{
		ID:             primitive.NewObjectID(),
		Username:       username,
		Email:          email,
		HashedPassword: hashedPassword,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	// Store user in all maps
	userID := user.ID.Hex()
	s.users[username] = user
	s.usersByEmail[email] = user
	s.usersById[userID] = user

	return user, nil
}

// FindByUsername retrieves a user by username
func (s *InMemoryUserService) FindByUsername(ctx context.Context, username string) (*User, error) {
	if username == "" {
		return nil, ErrInvalidInput
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, fmt.Errorf("username '%s': %w", username, ErrUserNotFound)
	}

	return user, nil
}

// FindByEmail retrieves a user by email
func (s *InMemoryUserService) FindByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, ErrInvalidInput
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmail[email]
	if !exists {
		return nil, fmt.Errorf("email '%s': %w", email, ErrUserNotFound)
	}

	return user, nil
}

// FindByID retrieves a user by ID
func (s *InMemoryUserService) FindByID(ctx context.Context, id string) (*User, error) {
	if id == "" {
		return nil, ErrInvalidInput
	}

	// Validate ObjectID format
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return nil, fmt.Errorf("id '%s': %w", id, ErrInvalidUserID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersById[id]
	if !exists {
		return nil, fmt.Errorf("id '%s': %w", id, ErrUserNotFound)
	}

	return user, nil
}

// UpdateUser updates an existing user
func (s *InMemoryUserService) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return ErrInvalidInput
	}

	userID := user.ID.Hex()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user exists
	existingUser, exists := s.usersById[userID]
	if !exists {
		return fmt.Errorf("id '%s': %w", userID, ErrUserNotFound)
	}

	// Find the old username and email by searching the maps
	var oldUsername, oldEmail string
	for username, u := range s.users {
		if u == existingUser {
			oldUsername = username
			break
		}
	}
	for email, u := range s.usersByEmail {
		if u == existingUser {
			oldEmail = email
			break
		}
	}

	// Check for username conflicts (only if username is changing)
	if oldUsername != user.Username {
		if _, exists := s.users[user.Username]; exists {
			return fmt.Errorf("username '%s': %w", user.Username, ErrUserAlreadyExists)
		}
	}

	// Check for email conflicts (only if email is changing)
	if oldEmail != user.Email {
		if _, exists := s.usersByEmail[user.Email]; exists {
			return fmt.Errorf("email '%s': %w", user.Email, ErrUserAlreadyExists)
		}
	}

	// Update timestamp
	user.UpdatedAt = time.Now()

	// Remove old entries from maps using found old values
	delete(s.users, oldUsername)
	delete(s.usersByEmail, oldEmail)

	// Add updated user to all maps with new values
	s.users[user.Username] = user
	s.usersByEmail[user.Email] = user
	s.usersById[userID] = user

	return nil
}

// DeleteUser deletes a user by ID
func (s *InMemoryUserService) DeleteUser(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidInput
	}

	// Validate ObjectID format
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		return fmt.Errorf("id '%s': %w", id, ErrInvalidUserID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.usersById[id]
	if !exists {
		return fmt.Errorf("id '%s': %w", id, ErrUserNotFound)
	}

	// Remove from all maps
	delete(s.users, user.Username)
	delete(s.usersByEmail, user.Email)
	delete(s.usersById, id)

	return nil
}

// VerifyPassword verifies the password for a user
func (s *InMemoryUserService) VerifyPassword(ctx context.Context, user *User, password string) error {
	if user == nil || password == "" {
		return ErrInvalidInput
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	// Find the user by ID to get the latest hashed password
	userID := user.ID.Hex()
	existingUser, exists := s.usersById[userID]
	if !exists {
		return fmt.Errorf("user ID '%s': %w", userID, ErrUserNotFound)
	}

	return s.passwordService.VerifyPassword(ctx, existingUser.HashedPassword, password)
}

// Collection returns the name of the MongoDB collection for users
func (s *InMemoryUserService) Collection() string {
	return "users"
} 