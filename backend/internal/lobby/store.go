// Package lobby implements lobby lifecycle management with thread-safe
// in-memory storage.
package lobby

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// MemoryStore implements models.LobbyStore with a sync.RWMutex-protected map.
// This is the default store; the interface allows swapping to Redis later.
type MemoryStore struct {
	mu      sync.RWMutex
	lobbies map[string]*models.Lobby
}

// NewMemoryStore creates an empty in-memory lobby store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		lobbies: make(map[string]*models.Lobby),
	}
}

// Create persists a new lobby and assigns it a unique 5-character code.
func (s *MemoryStore) Create(lobby *models.Lobby) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Generate unique lobby code
	var code string
	for {
		code = generateCode(5)
		if _, exists := s.lobbies[code]; !exists {
			break
		}
	}

	lobby.ID = code
	s.lobbies[code] = lobby
	return code, nil
}

// Get retrieves a lobby by its code. Returns nil if not found.
func (s *MemoryStore) Get(code string) (*models.Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	lobby, ok := s.lobbies[code]
	if !ok {
		return nil, nil
	}
	return lobby, nil
}

// Update replaces the lobby state atomically.
func (s *MemoryStore) Update(lobby *models.Lobby) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.lobbies[lobby.ID]; !exists {
		return fmt.Errorf("lobby %q not found", lobby.ID)
	}
	s.lobbies[lobby.ID] = lobby
	return nil
}

// Delete removes a lobby by its code.
func (s *MemoryStore) Delete(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.lobbies, code)
	return nil
}

// List returns all active (non-finished) lobbies.
func (s *MemoryStore) List() ([]*models.Lobby, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*models.Lobby
	for _, l := range s.lobbies {
		if l.State != models.LobbyStateFinished {
			result = append(result, l)
		}
	}
	return result, nil
}

// Compile-time interface check.
var _ models.LobbyStore = (*MemoryStore)(nil)

// generateCode produces a random alphanumeric code of the given length.
// Uses crypto/rand for unpredictability.
func generateCode(length int) string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // no 0/O/1/I confusion
	code := make([]byte, length)
	for i := range code {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		code[i] = charset[n.Int64()]
	}
	return string(code)
}
