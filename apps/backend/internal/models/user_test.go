package models

import (
	"context"
	"fmt"
	"testing"
	"time"

	"tennis-booker/internal/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestNewInMemoryUserService(t *testing.T) {
	service := NewInMemoryUserService()

	assert.NotNil(t, service)
	assert.NotNil(t, service.users)
	assert.NotNil(t, service.usersByEmail)
	assert.NotNil(t, service.usersById)
	assert.NotNil(t, service.passwordService)
	assert.Equal(t, 0, len(service.users))
}

func TestNewInMemoryUserServiceWithPasswordService(t *testing.T) {
	customPasswordService := auth.NewBcryptPasswordServiceWithCost(4) // Lower cost for faster tests
	service := NewInMemoryUserServiceWithPasswordService(customPasswordService)

	assert.NotNil(t, service)
	assert.NotNil(t, service.users)
	assert.NotNil(t, service.usersByEmail)
	assert.NotNil(t, service.usersById)
	assert.Equal(t, customPasswordService, service.passwordService)
	assert.Equal(t, 0, len(service.users))
}

func TestInMemoryUserService_CreateUser(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4)) // Faster for tests
	ctx := context.Background()

	t.Run("successful user creation", func(t *testing.T) {
		username := "testuser"
		email := "test@example.com"
		password := "DEMO_PASSWORD"

		user, err := service.CreateUser(ctx, username, email, password)
		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, username, user.Username)
		assert.Equal(t, email, user.Email)
		assert.NotEmpty(t, user.HashedPassword)
		assert.NotEqual(t, password, user.HashedPassword) // Password should be hashed
		assert.NotZero(t, user.ID)
		assert.False(t, user.CreatedAt.IsZero())
		assert.False(t, user.UpdatedAt.IsZero())

		// Verify password hashing worked
		err = service.VerifyPassword(ctx, user, password)
		assert.NoError(t, err)

		// Verify wrong password fails
		err = service.VerifyPassword(ctx, user, "wrongpassword")
		assert.Error(t, err)
	})

	t.Run("duplicate username", func(t *testing.T) {
		username := "duplicateuser"
		email1 := "user1@example.com"
		email2 := "user2@example.com"
		password := "DEMO_PASSWORD"

		// Create first user
		_, err := service.CreateUser(ctx, username, email1, password)
		require.NoError(t, err)

		// Try to create second user with same username
		_, err = service.CreateUser(ctx, username, email2, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
		assert.Contains(t, err.Error(), username)
	})

	t.Run("duplicate email", func(t *testing.T) {
		username1 := "user1"
		username2 := "user2"
		email := "duplicate@example.com"
		password := "DEMO_PASSWORD"

		// Create first user
		_, err := service.CreateUser(ctx, username1, email, password)
		require.NoError(t, err)

		// Try to create second user with same email
		_, err = service.CreateUser(ctx, username2, email, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
		assert.Contains(t, err.Error(), email)
	})

	t.Run("invalid input - empty username", func(t *testing.T) {
		_, err := service.CreateUser(ctx, "", "test@example.com", "DEMO_PASSWORD")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("invalid input - empty email", func(t *testing.T) {
		_, err := service.CreateUser(ctx, "testuser", "", "DEMO_PASSWORD")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("invalid input - empty password", func(t *testing.T) {
		_, err := service.CreateUser(ctx, "testuser", "test@example.com", "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})
}

func TestInMemoryUserService_FindByUsername(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Create a test user
	username := "finduser"
	email := "finduser@example.com"
	password := "DEMO_PASSWORD"
	createdUser, err := service.CreateUser(ctx, username, email, password)
	require.NoError(t, err)

	t.Run("successful find by username", func(t *testing.T) {
		foundUser, err := service.FindByUsername(ctx, username)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, username, foundUser.Username)
		assert.Equal(t, email, foundUser.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := service.FindByUsername(ctx, "nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), "nonexistent")
	})

	t.Run("invalid input - empty username", func(t *testing.T) {
		_, err := service.FindByUsername(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})
}

func TestInMemoryUserService_FindByEmail(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Create a test user
	username := "emailuser"
	email := "emailuser@example.com"
	password := "DEMO_PASSWORD"
	createdUser, err := service.CreateUser(ctx, username, email, password)
	require.NoError(t, err)

	t.Run("successful find by email", func(t *testing.T) {
		foundUser, err := service.FindByEmail(ctx, email)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, username, foundUser.Username)
		assert.Equal(t, email, foundUser.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		_, err := service.FindByEmail(ctx, "nonexistent@example.com")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), "nonexistent@example.com")
	})

	t.Run("invalid input - empty email", func(t *testing.T) {
		_, err := service.FindByEmail(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})
}

func TestInMemoryUserService_FindByID(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Create a test user
	username := "iduser"
	email := "iduser@example.com"
	password := "DEMO_PASSWORD"
	createdUser, err := service.CreateUser(ctx, username, email, password)
	require.NoError(t, err)

	userID := createdUser.ID.Hex()

	t.Run("successful find by ID", func(t *testing.T) {
		foundUser, err := service.FindByID(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, createdUser.ID, foundUser.ID)
		assert.Equal(t, username, foundUser.Username)
		assert.Equal(t, email, foundUser.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		_, err := service.FindByID(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), nonExistentID)
	})

	t.Run("invalid input - empty ID", func(t *testing.T) {
		_, err := service.FindByID(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("invalid input - malformed ID", func(t *testing.T) {
		_, err := service.FindByID(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
		assert.Contains(t, err.Error(), "invalid-id")
	})
}

func TestInMemoryUserService_VerifyPassword(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Create a test user
	username := "verifyuser"
	email := "verifyuser@example.com"
	password := "correctpassword"
	user, err := service.CreateUser(ctx, username, email, password)
	require.NoError(t, err)

	t.Run("successful password verification", func(t *testing.T) {
		err := service.VerifyPassword(ctx, user, password)
		assert.NoError(t, err)
	})

	t.Run("incorrect password", func(t *testing.T) {
		err := service.VerifyPassword(ctx, user, "wrongpassword")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid password")
	})

	t.Run("invalid input - nil user", func(t *testing.T) {
		err := service.VerifyPassword(ctx, nil, password)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("invalid input - empty password", func(t *testing.T) {
		err := service.VerifyPassword(ctx, user, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("user not found", func(t *testing.T) {
		nonExistentUser := &User{
			ID:       primitive.NewObjectID(),
			Username: "nonexistent",
			Email:    "nonexistent@example.com",
		}
		err := service.VerifyPassword(ctx, nonExistentUser, password)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})
}

func TestInMemoryUserService_UpdateUser(t *testing.T) {
	ctx := context.Background()

	t.Run("successful user update", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		// Create a test user
		username := "updateuser"
		email := "updateuser@example.com"
		password := "DEMO_PASSWORD"
		createdUser, err := service.CreateUser(ctx, username, email, password)
		require.NoError(t, err)

		// Modify user data
		createdUser.Name = "Updated Name"
		createdUser.Phone = "+1234567890"
		originalUpdatedAt := createdUser.UpdatedAt

		// Wait a bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		err = service.UpdateUser(ctx, createdUser)
		require.NoError(t, err)

		// Verify update
		updatedUser, err := service.FindByID(ctx, createdUser.ID.Hex())
		require.NoError(t, err)
		assert.Equal(t, "Updated Name", updatedUser.Name)
		assert.Equal(t, "+1234567890", updatedUser.Phone)
		assert.True(t, updatedUser.UpdatedAt.After(originalUpdatedAt))
	})

	t.Run("update username", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		// Create another user for this test
		user, err := service.CreateUser(ctx, "usernametest", "usernametest@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)

		// Update username
		oldUsername := user.Username
		user.Username = "newusername"

		err = service.UpdateUser(ctx, user)
		require.NoError(t, err)

		// Verify old username is gone and new one works
		_, err = service.FindByUsername(ctx, oldUsername)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		foundUser, err := service.FindByUsername(ctx, "newusername")
		require.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
	})

	t.Run("update email", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		// Create another user for this test
		user, err := service.CreateUser(ctx, "emailtest", "emailtest@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)

		// Update email
		oldEmail := user.Email
		user.Email = "newemail@example.com"

		err = service.UpdateUser(ctx, user)
		require.NoError(t, err)

		// Verify old email is gone and new one works
		_, err = service.FindByEmail(ctx, oldEmail)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		foundUser, err := service.FindByEmail(ctx, "newemail@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, foundUser.ID)
	})

	t.Run("update username conflict", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		// Create two users
		user1, err := service.CreateUser(ctx, "conflict1", "conflict1@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)
		user2, err := service.CreateUser(ctx, "conflict2", "conflict2@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)

		// Try to update user2's username to user1's username
		user2.Username = user1.Username
		err = service.UpdateUser(ctx, user2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
		assert.Contains(t, err.Error(), user1.Username)
	})

	t.Run("update email conflict", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		// Create two users
		user1, err := service.CreateUser(ctx, "emailconflict1", "emailconflict1@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)
		user2, err := service.CreateUser(ctx, "emailconflict2", "emailconflict2@example.com", "DEMO_PASSWORD")
		require.NoError(t, err)

		// Try to update user2's email to user1's email
		user2.Email = user1.Email
		err = service.UpdateUser(ctx, user2)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user already exists")
		assert.Contains(t, err.Error(), user1.Email)
	})

	t.Run("user not found", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		nonExistentUser := &User{
			ID:       primitive.NewObjectID(),
			Username: "nonexistent",
			Email:    "nonexistent@example.com",
		}

		err := service.UpdateUser(ctx, nonExistentUser)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("invalid input - nil user", func(t *testing.T) {
		service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))

		err := service.UpdateUser(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})
}

func TestInMemoryUserService_DeleteUser(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Create a test user
	username := "deleteuser"
	email := "deleteuser@example.com"
	password := "DEMO_PASSWORD"
	createdUser, err := service.CreateUser(ctx, username, email, password)
	require.NoError(t, err)

	userID := createdUser.ID.Hex()

	t.Run("successful user deletion", func(t *testing.T) {
		err := service.DeleteUser(ctx, userID)
		require.NoError(t, err)

		// Verify user is deleted from all maps
		_, err = service.FindByID(ctx, userID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		_, err = service.FindByUsername(ctx, username)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")

		_, err = service.FindByEmail(ctx, email)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("user not found", func(t *testing.T) {
		nonExistentID := primitive.NewObjectID().Hex()
		err := service.DeleteUser(ctx, nonExistentID)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "user not found")
		assert.Contains(t, err.Error(), nonExistentID)
	})

	t.Run("invalid input - empty ID", func(t *testing.T) {
		err := service.DeleteUser(ctx, "")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidInput, err)
	})

	t.Run("invalid input - malformed ID", func(t *testing.T) {
		err := service.DeleteUser(ctx, "invalid-id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid user ID")
		assert.Contains(t, err.Error(), "invalid-id")
	})
}

func TestInMemoryUserService_Collection(t *testing.T) {
	service := NewInMemoryUserService()
	assert.Equal(t, "users", service.Collection())
}

func TestInMemoryUserService_ConcurrentAccess(t *testing.T) {
	service := NewInMemoryUserServiceWithPasswordService(auth.NewBcryptPasswordServiceWithCost(4))
	ctx := context.Background()

	// Test concurrent user creation
	t.Run("concurrent user creation", func(t *testing.T) {
		const numGoroutines = 10
		results := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				username := fmt.Sprintf("user%d", id)
				email := fmt.Sprintf("user%d@example.com", id)
				_, err := service.CreateUser(ctx, username, email, "DEMO_PASSWORD")
				results <- err
			}(i)
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-results
			assert.NoError(t, err)
		}

		// Verify all users were created
		for i := 0; i < numGoroutines; i++ {
			username := fmt.Sprintf("user%d", i)
			user, err := service.FindByUsername(ctx, username)
			assert.NoError(t, err)
			assert.Equal(t, username, user.Username)
		}
	})
}
