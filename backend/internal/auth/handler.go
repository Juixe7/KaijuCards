package auth

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandler provides HTTP handlers for user registration and login.
type AuthHandler struct {
	repo   models.UserRepository
	secret string
	expiry time.Duration
}

// NewAuthHandler creates an AuthHandler with the given dependencies.
func NewAuthHandler(repo models.UserRepository, secret string, expiry time.Duration) *AuthHandler {
	return &AuthHandler{
		repo:   repo,
		secret: secret,
		expiry: expiry,
	}
}

// registerRequest is the expected JSON body for POST /api/auth/register.
type registerRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// loginRequest is the expected JSON body for POST /api/auth/login.
type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// authResponse is the JSON response for both register and login.
type authResponse struct {
	Token string      `json:"token,omitempty"`
	User  models.User `json:"user"`
}

// HandleRegister handles POST /api/auth/register.
// Creates a new user account with bcrypt-hashed password.
func (h *AuthHandler) HandleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	// Validate input
	if len(req.Username) < 3 || len(req.Username) > 30 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username must be 3-30 characters"})
		return
	}
	if len(req.Password) < 6 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Password must be at least 6 characters"})
		return
	}

	// Check if username is taken
	existing, err := h.repo.GetByUsername(req.Username)
	if err != nil {
		slog.Error("register: failed to check username", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "Username already taken"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error("register: failed to hash password", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	user := &models.User{
		Username:     req.Username,
		PasswordHash: string(hash),
	}

	if err := h.repo.CreateUser(user); err != nil {
		slog.Error("register: failed to create user", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Failed to create account"})
		return
	}

	slog.Info("user registered", "username", user.Username, "id", user.ID)
	writeJSON(w, http.StatusCreated, authResponse{User: *user})
}

// HandleLogin handles POST /api/auth/login.
// Validates credentials and returns a signed JWT on success.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "Username and password required"})
		return
	}

	// Look up user
	user, err := h.repo.GetByUsername(req.Username)
	if err != nil {
		slog.Error("login: failed to look up user", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		return
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "Invalid username or password"})
		return
	}

	// Generate JWT
	token, err := GenerateToken(user.ID, user.Username, h.secret, h.expiry)
	if err != nil {
		slog.Error("login: failed to generate token", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "Internal server error"})
		return
	}

	slog.Info("user logged in", "username", user.Username, "id", user.ID)
	writeJSON(w, http.StatusOK, authResponse{Token: token, User: *user})
}

// writeJSON is a helper that encodes data as JSON and writes it to the response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
