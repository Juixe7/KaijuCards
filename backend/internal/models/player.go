package models

// PlayerID is an opaque identifier for a player within a game session.
type PlayerID string

// Player represents a participant (human or bot) in a lobby/game.
type Player struct {
	ID       PlayerID `json:"id"`
	Name     string   `json:"name"`
	IsBot    bool     `json:"isBot"`
	Hand     Deck     `json:"hand"`
	Score    int      `json:"score"` // Round wins accumulated during the game

	// SocketID tracks the WebSocket connection for human players.
	// Empty for bots.
	SocketID string `json:"socketId,omitempty"`

	// UserID links to the authenticated user account, if any.
	// Empty for guest players and bots.
	UserID string `json:"userId,omitempty"`
}

// TopCard returns the first card in the player's hand without removing it.
// Returns nil if the hand is empty.
func (p *Player) TopCard() *Card {
	if len(p.Hand) == 0 {
		return nil
	}
	return &p.Hand[0]
}

// HasCards reports whether the player has cards remaining.
func (p *Player) HasCards() bool {
	return len(p.Hand) > 0
}

// RoundResult captures the outcome of a single round for history tracking.
type RoundResult struct {
	Round    int       `json:"round"`
	Attr     Attribute `json:"attr"`
	Reveals  []Card    `json:"reveals"`   // Each player's played card
	WinnerIDs []PlayerID  `json:"winnerIds"`

	// LobbyObj is used to pass the mutated lobby back to handlers (omitted from JSON)
	LobbyObj *Lobby `json:"-"`
}

// BotStrategy defines how a bot selects its attribute each turn.
// Pluggable so we can swap between simple max-stat and smarter heuristics.
type BotStrategy interface {
	// ChooseAttribute picks the best attribute to play given the bot's top card.
	ChooseAttribute(card Card) Attribute
}
