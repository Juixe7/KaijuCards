package auth

import (
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	secret := "test-secret-key-12345"
	userID := "550e8400-e29b-41d4-a716-446655440000"
	username := "testplayer"
	expiry := 1 * time.Hour

	// Generate
	token, err := GenerateToken(userID, username, secret, expiry)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned empty token")
	}

	// Validate
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Subject != userID {
		t.Errorf("Subject = %q, want %q", claims.Subject, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %q, want %q", claims.Username, username)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	token, err := GenerateToken("id", "user", "secret-a", time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, "secret-b")
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	// Create token that already expired
	token, err := GenerateToken("id", "user", "secret", -1*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = ValidateToken(token, "secret")
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	_, err := ValidateToken("not.a.real.token", "secret")
	if err == nil {
		t.Fatal("expected error for malformed token, got nil")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	_, err := ValidateToken("", "secret")
	if err == nil {
		t.Fatal("expected error for empty token, got nil")
	}
}
