package auth

// This file demonstrates how to integrate the JWT service with the SecretsManager
// The SecretsManager from the secrets package implements the JWTSecretsProvider interface

// Example usage (this would typically be in a main function or service initialization):
//
// func ExampleJWTIntegration() {
//     // Create secrets manager
//     secretsManager, err := secrets.NewSecretsManagerFromEnv()
//     if err != nil {
//         log.Fatal("Failed to create secrets manager:", err)
//     }
//     defer secretsManager.Close()
//
//     // Create JWT service using the secrets manager as the provider
//     // The SecretsManager implements JWTSecretsProvider interface
//     jwtService := NewJWTService(secretsManager, "tennis-booker")
//
//     // Generate a token
//     token, err := jwtService.GenerateToken("user123", "testuser", time.Hour)
//     if err != nil {
//         log.Fatal("Failed to generate token:", err)
//     }
//
//     // Validate the token
//     claims, err := jwtService.ValidateToken(token)
//     if err != nil {
//         log.Fatal("Failed to validate token:", err)
//     }
//
//     fmt.Printf("Token validated for user: %s\n", claims.UserID)
// }

// Note: The SecretsManager from the secrets package automatically implements
// the JWTSecretsProvider interface because it has the GetJWTSecret() method.
// This allows us to avoid import cycles while maintaining clean separation of concerns.
