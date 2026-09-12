package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGenerateAndValidateAccessToken(t *testing.T) {
	secret := "super-secret-key-for-test-32chars!"
	userID := uuid.New()
	email := "test@example.com"
	duration := 15 * time.Minute

	// Test case: Happy path - generate and validate
	token, exp, err := GenerateAccessToken(userID, email, secret, duration)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token string")
	}
	if exp.Before(time.Now()) {
		t.Fatalf("expected expiration time in the future, got %v", exp)
	}

	claims, err := ValidateAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken failed for valid token: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected UserID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %v, got %v", email, claims.Email)
	}
	if claims.Issuer != "go-gin-auth-service" {
		t.Errorf("expected Issuer 'go-gin-auth-service', got %v", claims.Issuer)
	}

	// Test case: Validate with wrong secret
	wrongSecret := "wrong-secret-key-for-testing-1234!"
	_, err = ValidateAccessToken(token, wrongSecret)
	if err == nil {
		t.Error("expected error when validating with wrong secret, got nil")
	}

	// Test case: Expired token
	expiredToken, _, err := GenerateAccessToken(userID, email, secret, -10*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken failed for expired token test: %v", err)
	}
	_, err = ValidateAccessToken(expiredToken, secret)
	if err == nil {
		t.Error("expected error for expired token, got nil")
	}

	// Test case: Malformed token
	_, err = ValidateAccessToken("not.a.valid.jwt.token", secret)
	if err == nil {
		t.Error("expected error for malformed token string, got nil")
	}
}

func TestGenerateSecureRandomString(t *testing.T) {
	lengths := []int{16, 32, 64}

	for _, length := range lengths {
		str1, err := GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("GenerateSecureRandomString(%d) error: %v", length, err)
		}
		// hex encoding doubles byte length
		expectedLen := length * 2
		if len(str1) != expectedLen {
			t.Errorf("expected string length %d, got %d", expectedLen, len(str1))
		}

		str2, err := GenerateSecureRandomString(length)
		if err != nil {
			t.Fatalf("GenerateSecureRandomString(%d) error: %v", length, err)
		}
		if str1 == str2 {
			t.Errorf("expected random strings to be unique, got identical %q", str1)
		}
	}
}

func TestHashToken(t *testing.T) {
	input := "my-refresh-token-12345"
	expectedHash := sha256.Sum256([]byte(input))
	expectedHex := hex.EncodeToString(expectedHash[:])

	gotHex := HashToken(input)
	if gotHex != expectedHex {
		t.Errorf("HashToken(%q) = %q; want %q", input, gotHex, expectedHex)
	}

	// Ensure hashing is deterministic
	if gotHex != HashToken(input) {
		t.Errorf("HashToken is not deterministic for same input")
	}
}
