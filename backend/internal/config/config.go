// Package config provides 12-factor environment configuration for KaijuCards backend.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all application settings loaded from environment variables.
type Config struct {
	// Server
	Port          int
	AllowedOrigin string

	// Jikan API
	JikanBaseURL string
	JikanTimeout time.Duration

	// Auth
	JWTSecret string
	JWTExpiry time.Duration

	// Database
	DatabaseURL string

	// AI
	GeminiAPIKey string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	loadDotEnv()

	return &Config{
		Port:          envInt("PORT", 8080),
		AllowedOrigin: envStr("ALLOWED_ORIGIN", "http://localhost:3000"),
		JikanBaseURL:  envStr("JIKAN_BASE_URL", "https://api.jikan.moe/v4"),
		JikanTimeout:  envDuration("JIKAN_TIMEOUT", 3*time.Second),
		JWTSecret:     envStr("JWT_SECRET", "dev-secret-change-in-production"),
		JWTExpiry:     envDuration("JWT_EXPIRY", 24*time.Hour),
		DatabaseURL:   envStr("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/meta_clash?sslmode=disable"),
		GeminiAPIKey:  envStr("GEMINI_API_KEY", ""),
	}
}

// loadDotEnv loads key-values from .env if present into os.Setenv without overwriting existing env vars.
func loadDotEnv() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err == nil {
			for _, line := range strings.Split(string(data), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					key := strings.TrimSpace(parts[0])
					val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
					if os.Getenv(key) == "" {
						os.Setenv(key, val)
					}
				}
			}
			return
		}
	}
}

func envStr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func envDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return fallback
}
