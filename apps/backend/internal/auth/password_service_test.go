package auth

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestNewBcryptPasswordService(t *testing.T) {
	service := NewBcryptPasswordService()
	
	assert.NotNil(t, service)
	assert.Equal(t, bcrypt.DefaultCost, service.GetCost())
}

func TestNewBcryptPasswordServiceWithCost(t *testing.T) {
	customCost := 12
	service := NewBcryptPasswordServiceWithCost(customCost)
	
	assert.NotNil(t, service)
	assert.Equal(t, customCost, service.GetCost())
}

func TestBcryptPasswordService_HashPassword(t *testing.T) {
	service := NewBcryptPasswordService()
	ctx := context.Background()

	t.Run("successful password hashing", func(t *testing.T) {
		password := "testDEMO_PASSWORD"
		
		hashedPassword, err := service.HashPassword(ctx, password)
		require.NoError(t, err)
		assert.NotEmpty(t, hashedPassword)
		assert.NotEqual(t, password, hashedPassword)
		
		// Verify the hash starts with bcrypt prefix
		assert.True(t, strings.HasPrefix(hashedPassword, "$2a$") || 
					strings.HasPrefix(hashedPassword, "$2b$") || 
					strings.HasPrefix(hashedPassword, "$2y$"))
		
		// Verify we can verify the password with the hash
		err = service.VerifyPassword(ctx, hashedPassword, password)
		assert.NoError(t, err)
	})

	t.Run("different passwords produce different hashes", func(t *testing.T) {
		password1 := "password1"
		password2 := "password2"
		
		hash1, err := service.HashPassword(ctx, password1)
		require.NoError(t, err)
		
		hash2, err := service.HashPassword(ctx, password2)
		require.NoError(t, err)
		
		assert.NotEqual(t, hash1, hash2)
	})

	t.Run("same password produces different hashes (salt)", func(t *testing.T) {
		password := "samepassword"
		
		hash1, err := service.HashPassword(ctx, password)
		require.NoError(t, err)
		
		hash2, err := service.HashPassword(ctx, password)
		require.NoError(t, err)
		
		// Different hashes due to random salt
		assert.NotEqual(t, hash1, hash2)
		
		// But both should verify correctly
		err = service.VerifyPassword(ctx, hash1, password)
		assert.NoError(t, err)
		
		err = service.VerifyPassword(ctx, hash2, password)
		assert.NoError(t, err)
	})

	t.Run("empty password", func(t *testing.T) {
		_, err := service.HashPassword(ctx, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password cannot be empty")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		_, err := service.HashPassword(ctx, "password")
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()
		
		time.Sleep(1 * time.Millisecond) // Ensure timeout
		
		_, err := service.HashPassword(ctx, "password")
		assert.Error(t, err)
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}

func TestBcryptPasswordService_VerifyPassword(t *testing.T) {
	service := NewBcryptPasswordService()
	ctx := context.Background()

	t.Run("successful password verification", func(t *testing.T) {
		password := "correctpassword"
		hashedPassword, err := service.HashPassword(ctx, password)
		require.NoError(t, err)
		
		err = service.VerifyPassword(ctx, hashedPassword, password)
		assert.NoError(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		password := "correctpassword"
		wrongPassword := "wrongpassword"
		hashedPassword, err := service.HashPassword(ctx, password)
		require.NoError(t, err)
		
		err = service.VerifyPassword(ctx, hashedPassword, wrongPassword)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password")
	})

	t.Run("empty hashed password", func(t *testing.T) {
		err := service.VerifyPassword(ctx, "", "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "hashed password cannot be empty")
	})

	t.Run("empty password", func(t *testing.T) {
		hashedPassword := "$2a$10$validhashformat"
		err := service.VerifyPassword(ctx, hashedPassword, "")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "password cannot be empty")
	})

	t.Run("invalid hash format", func(t *testing.T) {
		invalidHash := "not-a-valid-bcrypt-hash"
		err := service.VerifyPassword(ctx, invalidHash, "password")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to verify password")
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		
		hashedPassword := "$2a$10$validhashformat"
		err := service.VerifyPassword(ctx, hashedPassword, "password")
		assert.Error(t, err)
		assert.Equal(t, context.Canceled, err)
	})
}

func TestBcryptPasswordService_SetCost(t *testing.T) {
	service := NewBcryptPasswordService()

	t.Run("valid cost", func(t *testing.T) {
		newCost := 12
		err := service.SetCost(newCost)
		assert.NoError(t, err)
		assert.Equal(t, newCost, service.GetCost())
	})

	t.Run("cost too low", func(t *testing.T) {
		invalidCost := bcrypt.MinCost - 1
		err := service.SetCost(invalidCost)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bcrypt cost")
	})

	t.Run("cost too high", func(t *testing.T) {
		invalidCost := bcrypt.MaxCost + 1
		err := service.SetCost(invalidCost)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid bcrypt cost")
	})

	t.Run("minimum valid cost", func(t *testing.T) {
		err := service.SetCost(bcrypt.MinCost)
		assert.NoError(t, err)
		assert.Equal(t, bcrypt.MinCost, service.GetCost())
	})

	t.Run("maximum valid cost", func(t *testing.T) {
		err := service.SetCost(bcrypt.MaxCost)
		assert.NoError(t, err)
		assert.Equal(t, bcrypt.MaxCost, service.GetCost())
	})
}

func TestBcryptPasswordService_DifferentCosts(t *testing.T) {
	ctx := context.Background()
	password := "testpassword"

	t.Run("different costs produce different hashes", func(t *testing.T) {
		service1 := NewBcryptPasswordServiceWithCost(bcrypt.MinCost)
		service2 := NewBcryptPasswordServiceWithCost(bcrypt.MinCost + 1)
		
		hash1, err := service1.HashPassword(ctx, password)
		require.NoError(t, err)
		
		hash2, err := service2.HashPassword(ctx, password)
		require.NoError(t, err)
		
		// Hashes should be different due to different costs
		assert.NotEqual(t, hash1, hash2)
		
		// But each service should verify its own hash
		err = service1.VerifyPassword(ctx, hash1, password)
		assert.NoError(t, err)
		
		err = service2.VerifyPassword(ctx, hash2, password)
		assert.NoError(t, err)
		
		// And cross-verification should also work (bcrypt is backward compatible)
		err = service1.VerifyPassword(ctx, hash2, password)
		assert.NoError(t, err)
		
		err = service2.VerifyPassword(ctx, hash1, password)
		assert.NoError(t, err)
	})
}

func TestBcryptPasswordService_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	password := "performancetest"

	t.Run("hashing performance with different costs", func(t *testing.T) {
		costs := []int{bcrypt.MinCost, bcrypt.DefaultCost, bcrypt.DefaultCost + 2}
		
		for _, cost := range costs {
			t.Run(fmt.Sprintf("cost_%d", cost), func(t *testing.T) {
				service := NewBcryptPasswordServiceWithCost(cost)
				
				start := time.Now()
				_, err := service.HashPassword(ctx, password)
				duration := time.Since(start)
				
				require.NoError(t, err)
				t.Logf("Cost %d took %v", cost, duration)
				
				// Sanity check: higher cost should generally take longer
				// (though this isn't a strict requirement for the test)
				if cost == bcrypt.MinCost {
					assert.Less(t, duration, 100*time.Millisecond, "Min cost should be fast")
				}
			})
		}
	})
} 