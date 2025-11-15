package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestInitJWT(t *testing.T) {
	secret := "test-secret-key"
	InitJWT(secret)

	if len(jwtSecret) == 0 {
		t.Error("JWT secret was not initialized")
	}

	if string(jwtSecret) != secret {
		t.Error("JWT secret does not match expected value")
	}
}

func TestGenerateToken(t *testing.T) {
	// Initialize JWT
	InitJWT("test-secret-key")

	userID := 1
	email := "test@example.com"
	role := "professor"

	token, err := GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Error("Generated token is empty")
	}

	// Token should be reasonably long (JWT tokens are base64 encoded)
	if len(token) < 50 {
		t.Error("Generated token seems too short")
	}
}

func TestValidateToken(t *testing.T) {
	// Initialize JWT
	InitJWT("test-secret-key")

	userID := 1
	email := "test@example.com"
	role := "professor"

	// Generate token
	token, err := GenerateToken(userID, email, role)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Validate token
	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	// Check claims
	if claims.UserID != userID {
		t.Errorf("Expected UserID %d, got %d", userID, claims.UserID)
	}

	if claims.Email != email {
		t.Errorf("Expected Email %s, got %s", email, claims.Email)
	}

	if claims.Role != role {
		t.Errorf("Expected Role %s, got %s", role, claims.Role)
	}

	// Check expiration is set
	if claims.ExpiresAt == nil {
		t.Error("Token expiration is not set")
	}

	// Check expiration is in the future
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token expiration is in the past")
	}

	// Check expiration is approximately 24 hours from now
	expectedExpiry := time.Now().Add(24 * time.Hour)
	diff := expectedExpiry.Sub(claims.ExpiresAt.Time)
	if diff > 5*time.Minute || diff < -5*time.Minute {
		t.Errorf("Token expiration is not approximately 24 hours: diff = %v", diff)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	InitJWT("test-secret-key")

	// Test with invalid token
	_, err := ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token, got nil")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	// Generate token with one secret
	InitJWT("secret-1")
	token, err := GenerateToken(1, "test@example.com", "professor")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Try to validate with different secret
	InitJWT("secret-2")
	_, err = ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating token with wrong secret, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	InitJWT("test-secret-key")

	// Create a token with expiration in the past
	claims := Claims{
		UserID: 1,
		Email:  "test@example.com",
		Role:   "professor",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}

	// Validate should fail for expired token
	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Error("Expected error for expired token, got nil")
	}
}

func TestGenerateToken_WithoutInit(t *testing.T) {
	// Reset JWT secret
	jwtSecret = nil

	_, err := GenerateToken(1, "test@example.com", "professor")
	if err == nil {
		t.Error("Expected error when JWT secret is not initialized, got nil")
	}
}

func TestValidateToken_WithoutInit(t *testing.T) {
	// Reset JWT secret
	jwtSecret = nil

	_, err := ValidateToken("some.token.here")
	if err == nil {
		t.Error("Expected error when JWT secret is not initialized, got nil")
	}
}

