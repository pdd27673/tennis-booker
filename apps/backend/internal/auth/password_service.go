package auth

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService defines the interface for password operations
type PasswordService interface {
	HashPassword(ctx context.Context, password string) (string, error)
	VerifyPassword(ctx context.Context, hashedPassword, password string) error
}

// BcryptPasswordService implements PasswordService using bcrypt
type BcryptPasswordService struct {
	cost int // bcrypt cost factor
}

// NewBcryptPasswordService creates a new bcrypt password service
func NewBcryptPasswordService() *BcryptPasswordService {
	return &BcryptPasswordService{
		cost: bcrypt.DefaultCost, // Cost of 10 (good balance of security and performance)
	}
}

// NewBcryptPasswordServiceWithCost creates a new bcrypt password service with custom cost
func NewBcryptPasswordServiceWithCost(cost int) *BcryptPasswordService {
	return &BcryptPasswordService{
		cost: cost,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (s *BcryptPasswordService) HashPassword(ctx context.Context, password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	// Check context for cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedBytes), nil
}

// VerifyPassword verifies a plain text password against a bcrypt hash
func (s *BcryptPasswordService) VerifyPassword(ctx context.Context, hashedPassword, password string) error {
	if hashedPassword == "" {
		return errors.New("hashed password cannot be empty")
	}
	if password == "" {
		return errors.New("password cannot be empty")
	}

	// Check context for cancellation
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return errors.New("invalid password")
		}
		return fmt.Errorf("failed to verify password: %w", err)
	}

	return nil
}

// GetCost returns the current bcrypt cost factor
func (s *BcryptPasswordService) GetCost() int {
	return s.cost
}

// SetCost updates the bcrypt cost factor (for testing or configuration changes)
func (s *BcryptPasswordService) SetCost(cost int) error {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return fmt.Errorf("invalid bcrypt cost: %d (must be between %d and %d)", 
			cost, bcrypt.MinCost, bcrypt.MaxCost)
	}
	s.cost = cost
	return nil
} 