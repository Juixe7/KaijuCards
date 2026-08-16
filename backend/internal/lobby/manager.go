package lobby

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/game"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// Manager implements models.LobbyManager by orchestrating lobby lifecycle
// operations. Dependencies are injected via the constructor.
type Manager struct {
	store     models.LobbyStore
	generator models.CardGenerator
	engine    *game.Engine
	userRepo  models.UserRepository
}

// NewManager creates a LobbyManager with injected dependencies.
func NewManager(store models.LobbyStore, generator models.CardGenerator, engine *game.Engine, userRepo models.UserRepository) *Manager {
	return &Manager{
		store:     store,
		generator: generator,
		engine:    engine,
		userRepo:  userRepo,
	}
}


// CreateLobby initializes a new lobby with the given theme, generates cards,
// and adds the host as the first player.
func (m *Manager) CreateLobby(theme string, host models.Player) (*models.Lobby, error) {
	deck, _, err := m.generator.Generate(theme)
	if err != nil {
		return nil, fmt.Errorf("failed to generate cards for theme %q: %w", theme, err)
	}

	lobby := &models.Lobby{
		Theme:              theme,
		State:              models.LobbyStateWaiting,
		Players:            []models.Player{host},
		Deck:               deck,
		MaxPlayers:         models.MaxPlayers,
		CurrentPlayerIndex: 0,
		Round:              1,
		History:            []models.RoundResult{},
		CreatedAt:          time.Now(),
	}

	code, err := m.store.Create(lobby)
	if err != nil {
		return nil, fmt.Errorf("failed to create lobby: %w", err)
	}

	lobby.ID = code
	return lobby, nil
}

// JoinLobby adds a player to an existing lobby. If a player with the same name
// already exists (reconnection), their socket ID is updated instead.
func (m *Manager) JoinLobby(code string, player models.Player) (*models.Lobby, error) {
	lobby, err := m.store.Get(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby %q not found", code)
	}

	// Check for reconnection — same name, update socket ID
	for i, p := range lobby.Players {
		if p.Name == player.Name {
			lobby.Players[i].SocketID = player.SocketID
			if err := m.store.Update(lobby); err != nil {
				return nil, fmt.Errorf("failed to update lobby: %w", err)
			}
			return lobby, nil
		}
	}

	if lobby.State != models.LobbyStateWaiting {
		return nil, fmt.Errorf("cannot join: game already in progress")
	}
	if len(lobby.Players) >= lobby.MaxPlayers {
		return nil, fmt.Errorf("cannot join: lobby is full")
	}

	lobby.Players = append(lobby.Players, player)
	if err := m.store.Update(lobby); err != nil {
		return nil, fmt.Errorf("failed to update lobby: %w", err)
	}
	return lobby, nil
}

// AddBot creates a bot player and adds it to the lobby.
func (m *Manager) AddBot(code string) (*models.Lobby, error) {
	lobby, err := m.store.Get(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby %q not found", code)
	}
	if len(lobby.Players) >= lobby.MaxPlayers {
		return nil, fmt.Errorf("cannot add bot: lobby is full")
	}

	bot := models.Player{
		ID:    models.PlayerID(randomShortID()),
		Name:  "BOT-" + randomShortID()[:3],
		IsBot: true,
		Hand:  nil,
		Score: 0,
	}

	lobby.Players = append(lobby.Players, bot)
	if err := m.store.Update(lobby); err != nil {
		return nil, fmt.Errorf("failed to update lobby: %w", err)
	}
	return lobby, nil
}

// StartGame transitions the lobby from waiting → playing.
// Auto-fills empty slots with bots, shuffles deck, and deals cards.
func (m *Manager) StartGame(code string) (*models.Lobby, error) {
	lobby, err := m.store.Get(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby %q not found", code)
	}
	if lobby.State != models.LobbyStateWaiting {
		return nil, fmt.Errorf("cannot start: lobby state is %q", lobby.State)
	}
	if len(lobby.Players) < 2 {
		return nil, fmt.Errorf("cannot start: need at least 2 players")
	}

	// Removed auto-fill with bots to allow dynamic lobby sizes

	// Deal cards (also transitions state to playing)
	if err := m.engine.Deal(lobby); err != nil {
		return nil, fmt.Errorf("failed to deal cards: %w", err)
	}

	if err := m.store.Update(lobby); err != nil {
		return nil, fmt.Errorf("failed to update lobby: %w", err)
	}
	return lobby, nil
}

// PlayRound processes a player's attribute choice, resolves the round, and recursively handles bots.
func (m *Manager) PlayRound(code string, playerID models.PlayerID, attr string) (*models.Lobby, *models.RoundResult, error) {
	lobby, err := m.store.Get(code)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, nil, fmt.Errorf("lobby %q not found", code)
	}

	if lobby.State != models.LobbyStatePlaying {
		return nil, nil, fmt.Errorf("cannot play round: game is not playing")
	}

	currentPlayer := lobby.Players[lobby.CurrentPlayerIndex]
	if currentPlayer.ID != playerID {
		return nil, nil, fmt.Errorf("not your turn")
	}

	// Validate attribute
	validAttr := false
	for _, a := range models.AllAttributes() {
		if string(a) == attr {
			validAttr = true
			break
		}
	}
	if !validAttr {
		return nil, nil, fmt.Errorf("invalid attribute: %s", attr)
	}

	result, err := m.engine.ResolveRound(lobby, models.Attribute(attr))
	if err != nil {
		return nil, nil, err
	}

	// Attach the mutated lobby to the result purely for convenience if needed
	result.LobbyObj = lobby

	if lobby.State == models.LobbyStateFinished && m.userRepo != nil {
		winnerID := ""
		if lobby.Winner != nil && !lobby.Winner.IsBot && lobby.Winner.UserID != "" {
			winnerID = lobby.Winner.UserID
		}
		if !isValidUUID(winnerID) {
			winnerID = ""
		}
		
		// Use lobby.ID (the code) combined with timestamp to ensure unique match ID
		// or just a random UUID. We'll use randomShortID() + lobby.ID
		matchID := randomShortID() + "-" + lobby.ID

		match := &models.Match{
			ID:         matchID,
			LobbyCode:  lobby.ID,
			Theme:      lobby.Theme,
			WinnerID:   winnerID,
			StartedAt:  lobby.CreatedAt, // Assuming lobby creation time is roughly start time
			FinishedAt: time.Now(),
		}

		var mPlayers []models.MatchPlayer
		for _, p := range lobby.Players {
			userID := p.UserID
			if p.IsBot || userID == "" || !isValidUUID(userID) {
				userID = ""
			}
			mPlayers = append(mPlayers, models.MatchPlayer{
				MatchID: matchID,
				UserID:  userID,
				IsBot:   p.IsBot,
				Score:   p.Score,
			})
		}

		// Fire and forget, or handle error. It's safe to just log it, but here we return it if it fails
		if err := m.userRepo.SaveMatch(match, mPlayers); err != nil {
			// In production, we might just log this rather than returning a 500 to the websocket,
			// Returning the error gracefully ensures the game state is not corrupted.
			slog.Error("Failed to save match history", "err", err, "lobbyId", lobby.ID)
		}
	}

	if err := m.store.Update(lobby); err != nil {
		return nil, nil, fmt.Errorf("failed to update lobby: %w", err)
	}
	
	return lobby, result, nil
}

// LeaveLobby removes a player by socket ID. If the lobby empties while still
// in waiting state, it is deleted entirely.
func (m *Manager) LeaveLobby(code string, socketID string) (*models.Lobby, error) {
	lobby, err := m.store.Get(code)
	if err != nil {
		return nil, fmt.Errorf("failed to get lobby: %w", err)
	}
	if lobby == nil {
		return nil, fmt.Errorf("lobby %q not found", code)
	}

	if lobby.State == models.LobbyStateWaiting {
		// Remove the player
		filtered := lobby.Players[:0]
		for _, p := range lobby.Players {
			if p.SocketID != socketID {
				filtered = append(filtered, p)
			}
		}
		lobby.Players = filtered
	}

	// Clean up empty lobbies
	if len(lobby.Players) == 0 && lobby.State == models.LobbyStateWaiting {
		if err := m.store.Delete(code); err != nil {
			return nil, fmt.Errorf("failed to delete empty lobby: %w", err)
		}
		return lobby, nil
	}

	if err := m.store.Update(lobby); err != nil {
		return nil, fmt.Errorf("failed to update lobby: %w", err)
	}
	return lobby, nil
}

// Compile-time interface check.
var _ models.LobbyManager = (*Manager)(nil)

// randomShortID generates a short random hex string for player/bot IDs.
func randomShortID() string {
	b := make([]byte, 3)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// isValidUUID validates whether a string matches standard UUID format (36 chars, hex + hyphens).
func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}
