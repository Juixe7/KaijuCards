package models

import "time"

// LobbyState represents the finite-state-machine stages of a lobby.
type LobbyState string

const (
	LobbyStateWaiting  LobbyState = "waiting"  // Accepting joins, not yet started
	LobbyStatePlaying  LobbyState = "playing"  // Game in progress
	LobbyStateFinished LobbyState = "finished" // All rounds complete, winner decided
)

// Lobby holds the complete runtime state for a single game session.
// Persisted in-memory during gameplay; only match results go to the DB.
type Lobby struct {
	ID                string        `json:"id"`
	Theme             string        `json:"theme"`
	State             LobbyState    `json:"state"`
	Players           []Player      `json:"players"`
	Deck              Deck          `json:"deck"`              // Full generated deck (pre-deal)
	Kitty             Deck          `json:"kitty"`             // Leftover cards after dealing
	MaxPlayers        int           `json:"maxPlayers"`
	CurrentPlayerIndex int          `json:"currentPlayerIndex"`
	Round             int           `json:"round"`
	History           []RoundResult `json:"history"`
	Winner            *Player       `json:"winner,omitempty"`  // Set when state → finished
	CreatedAt         time.Time     `json:"createdAt"`
}

// LobbyStore is the interface for lobby CRUD operations.
// Default implementation is in-memory; interface allows swapping to Redis later
// as noted in the implementation plan.
type LobbyStore interface {
	// Create persists a new lobby and returns its assigned code.
	Create(lobby *Lobby) (string, error)

	// Get retrieves a lobby by its code. Returns nil if not found.
	Get(code string) (*Lobby, error)

	// Update replaces the lobby state atomically.
	Update(lobby *Lobby) error

	// Delete removes a lobby (e.g. when all players leave during waiting).
	Delete(code string) error

	// List returns all active (non-finished) lobbies.
	List() ([]*Lobby, error)
}

// LobbyManager orchestrates lobby lifecycle operations:
// creating, joining, starting games, and cleaning up.
type LobbyManager interface {
	// CreateLobby initializes a new lobby with the given theme and host player.
	CreateLobby(theme string, host Player) (*Lobby, error)

	// JoinLobby adds a player to an existing lobby.
	JoinLobby(code string, player Player) (*Lobby, error)

	// AddBot adds a bot to the lobby, respecting MaxPlayers.
	AddBot(code string) (*Lobby, error)

	// StartGame transitions the lobby from waiting → playing,
	// fills empty slots with bots, shuffles, and deals cards.
	StartGame(code string) (*Lobby, error)

	// PlayRound executes a game round for the given lobby, resolving logic,
	// checking win states, advancing turn, and bot execution.
	PlayRound(code string, playerID PlayerID, attr string) (*Lobby, *RoundResult, error)

	// LeaveLobby removes a player. Deletes the lobby if empty and still waiting.
	LeaveLobby(code string, socketID string) (*Lobby, error)
}
