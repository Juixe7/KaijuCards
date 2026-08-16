package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// mockRepo is a simple in-memory implementation of models.UserRepository for testing.
type mockRepo struct {
	users map[string]*models.User // keyed by username
}

func newMockRepo() *mockRepo {
	return &mockRepo{users: make(map[string]*models.User)}
}

func (m *mockRepo) CreateUser(user *models.User) error {
	user.ID = "test-uuid-" + user.Username
	user.CreatedAt = time.Now()
	m.users[user.Username] = user
	return nil
}

func (m *mockRepo) GetByID(id string) (*models.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockRepo) GetByUsername(username string) (*models.User, error) {
	u, ok := m.users[username]
	if !ok {
		return nil, nil
	}
	return u, nil
}

func (m *mockRepo) SaveMatch(match *models.Match, players []models.MatchPlayer) error {
	return nil
}

func (m *mockRepo) GetMatchHistory(userID string, limit int) ([]models.MatchSummary, error) {
	return nil, nil
}

func (m *mockRepo) GetWinLoss(userID string) (int, int, error) {
	return 0, 0, nil
}

func TestHandleRegister_Success(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	body := `{"username":"newplayer","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleRegister(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	var resp authResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.User.Username != "newplayer" {
		t.Errorf("username = %q, want %q", resp.User.Username, "newplayer")
	}
	if resp.User.ID == "" {
		t.Error("expected non-empty user ID")
	}
}

func TestHandleRegister_DuplicateUsername(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	// Pre-seed a user
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	repo.users["existing"] = &models.User{ID: "x", Username: "existing", PasswordHash: string(hash)}

	body := `{"username":"existing","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleRegister(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestHandleRegister_ShortPassword(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	body := `{"username":"player","password":"12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleRegister_ShortUsername(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	body := `{"username":"ab","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleRegister(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleLogin_Success(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	// Seed user
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass123"), bcrypt.MinCost)
	repo.users["testuser"] = &models.User{
		ID:           "uuid-test",
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	body := `{"username":"testuser","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleLogin(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d. Body: %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp authResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.User.Username != "testuser" {
		t.Errorf("username = %q, want %q", resp.User.Username, "testuser")
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)
	repo.users["testuser"] = &models.User{
		ID:           "uuid-test",
		Username:     "testuser",
		PasswordHash: string(hash),
	}

	body := `{"username":"testuser","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestHandleLogin_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	h := NewAuthHandler(repo, "test-secret", time.Hour)

	body := `{"username":"nobody","password":"pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBufferString(body))
	w := httptest.NewRecorder()

	h.HandleLogin(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}
