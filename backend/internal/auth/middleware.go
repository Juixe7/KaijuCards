package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// contextKey is an unexported type for context keys to avoid collisions.
type contextKey string

const (
	ctxUserID   contextKey = "userID"
	ctxUsername contextKey = "username"
)

// RequireAuth returns middleware that rejects requests without a valid JWT.
// The token must be provided in the Authorization header as "Bearer <token>".
// On success, the user's ID and username are injected into the request context.
func RequireAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := extractClaims(r, secret)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"` + err.Error() + `"}`))
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, claims.Subject)
			ctx = context.WithValue(ctx, ctxUsername, claims.Username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth returns middleware that extracts JWT claims if present,
// but does NOT reject unauthenticated requests. This allows guest mode:
// context values will be empty strings for guests.
func OptionalAuth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := extractClaims(r, secret)
			if err == nil && claims != nil {
				ctx := context.WithValue(r.Context(), ctxUserID, claims.Subject)
				ctx = context.WithValue(ctx, ctxUsername, claims.Username)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// UserIDFromContext extracts the authenticated user's ID from the request context.
// Returns an empty string for guest/unauthenticated users.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

// UsernameFromContext extracts the authenticated user's username from the request context.
// Returns an empty string for guest/unauthenticated users.
func UsernameFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ctxUsername).(string)
	return v
}

// extractClaims parses the Authorization header and validates the JWT.
func extractClaims(r *http.Request, secret string) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		// Also check query parameter for WebSocket connections
		// (browsers can't set headers on WebSocket upgrade)
		tokenStr := r.URL.Query().Get("token")
		if tokenStr == "" {
			return nil, fmt.Errorf("missing authorization")
		}
		return ValidateToken(tokenStr, secret)
	}

	// Expect "Bearer <token>"
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return nil, fmt.Errorf("invalid authorization format")
	}

	return ValidateToken(parts[1], secret)
}
