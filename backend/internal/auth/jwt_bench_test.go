package auth

import (
	"testing"
	"time"
)

// BenchmarkGenerateToken measures HS256 JWT creation.
func BenchmarkGenerateToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateToken("user-123", "testuser", "benchmark-secret-key", 24*time.Hour)
	}
}

// BenchmarkValidateToken measures JWT parse + signature verification.
func BenchmarkValidateToken(b *testing.B) {
	secret := "benchmark-secret-key"
	token, _ := GenerateToken("user-123", "testuser", secret, 24*time.Hour)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ValidateToken(token, secret)
	}
}
