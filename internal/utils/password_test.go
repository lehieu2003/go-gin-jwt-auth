package utils

import (
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "SecretPassword@123"

	// 1. Hash password
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty password hash")
	}
	if hash == password {
		t.Fatal("hash should not equal plain password")
	}

	// 2. Check correct password
	if !CheckPasswordHash(password, hash) {
		t.Error("CheckPasswordHash failed for correct password")
	}

	// 3. Check wrong password
	if CheckPasswordHash("WrongPassword@456", hash) {
		t.Error("CheckPasswordHash succeeded for wrong password")
	}

	// 4. Check empty password against hash
	if CheckPasswordHash("", hash) {
		t.Error("CheckPasswordHash succeeded for empty password")
	}

	// 5. Check password with invalid hash string
	if CheckPasswordHash(password, "invalid-hash-format") {
		t.Error("CheckPasswordHash succeeded for invalid hash format")
	}

	// 6. Verify salting: two hashes of the same password should differ
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword second run failed: %v", err)
	}
	if hash == hash2 {
		t.Error("expected different hashes due to bcrypt random salting")
	}
	if !CheckPasswordHash(password, hash2) {
		t.Error("CheckPasswordHash failed for second hash")
	}
}
