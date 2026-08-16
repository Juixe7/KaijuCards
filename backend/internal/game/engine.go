package game

import (
	"fmt"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// Engine implements the core game logic: dealing cards, resolving rounds,
// and determining the winner.
type Engine struct{}

// NewEngine creates a game engine.
func NewEngine() *Engine {
	return &Engine{}
}

// Deal shuffles the lobby's deck and distributes CardsPerPlayer cards
// to each player in round-robin order. Any leftover cards go to Kitty.
// Mutates the lobby's Players[].Hand and Kitty fields.
func (e *Engine) Deal(lobby *models.Lobby) error {
	if lobby.State != models.LobbyStateWaiting {
		return fmt.Errorf("cannot deal: lobby state is %q, expected %q", lobby.State, models.LobbyStateWaiting)
	}
	if len(lobby.Players) < 2 {
		return fmt.Errorf("cannot deal: need at least 2 players, have %d", len(lobby.Players))
	}

	shuffled := ShuffleDeck(lobby.Deck)

	// Reset hands
	for i := range lobby.Players {
		lobby.Players[i].Hand = nil
	}

	// Round-robin deal
	playerCount := len(lobby.Players)
	totalToDeal := models.CardsPerPlayer * playerCount
	if totalToDeal > len(shuffled) {
		totalToDeal = len(shuffled)
	}

	for i := 0; i < totalToDeal; i++ {
		pIdx := i % playerCount
		lobby.Players[pIdx].Hand = append(lobby.Players[pIdx].Hand, shuffled[i])
	}

	// Remainder goes to kitty
	if totalToDeal < len(shuffled) {
		lobby.Kitty = shuffled[totalToDeal:]
	} else {
		lobby.Kitty = nil
	}

	// Transition state
	lobby.State = models.LobbyStatePlaying
	lobby.Round = 1
	lobby.CurrentPlayerIndex = 0

	return nil
}

// ResolveRound compares the chosen attribute across all players' top cards.
// Returns the round result and mutates lobby state (advances round, updates scores).
func (e *Engine) ResolveRound(lobby *models.Lobby, attr models.Attribute) (*models.RoundResult, error) {
	if lobby.State != models.LobbyStatePlaying {
		return nil, fmt.Errorf("cannot resolve: lobby state is %q, expected %q", lobby.State, models.LobbyStatePlaying)
	}

	// Collect top cards from all players
	reveals := make([]models.Card, len(lobby.Players))
	for i, p := range lobby.Players {
		if !p.HasCards() {
			return nil, fmt.Errorf("player %q has no cards remaining", p.Name)
		}
		reveals[i] = lobby.Players[i].Hand[0]
		// Remove the top card
		lobby.Players[i].Hand = lobby.Players[i].Hand[1:]
	}

	// Find the winner by comparing the chosen attribute
	bestVal := -1
	var winnerIdxs []int
	for i, card := range reveals {
		val := card.Stats.Get(attr)
		if val > bestVal {
			bestVal = val
			winnerIdxs = []int{i}
		} else if val == bestVal {
			winnerIdxs = append(winnerIdxs, i)
		}
	}

	// Update scores
	var winnerIDs []models.PlayerID
	for _, idx := range winnerIdxs {
		lobby.Players[idx].Score++
		winnerIDs = append(winnerIDs, lobby.Players[idx].ID)
	}

	result := &models.RoundResult{
		Round:     lobby.Round,
		Attr:      attr,
		Reveals:   reveals,
		WinnerIDs: winnerIDs,
	}
	lobby.History = append(lobby.History, *result)

	// Winner goes next (if tied, first tied player goes next)
	lobby.CurrentPlayerIndex = winnerIdxs[0]

	// Check if game is over (no cards remaining)
	cardsRemaining := 0
	for _, p := range lobby.Players {
		cardsRemaining += len(p.Hand)
	}

	if cardsRemaining == 0 {
		lobby.State = models.LobbyStateFinished
		lobby.Winner = e.DetermineWinner(lobby)
	} else {
		lobby.Round++
	}

	return result, nil
}

// DetermineWinner returns a pointer to the player with the highest score.
// In case of a tie, the first player with the highest score wins.
func (e *Engine) DetermineWinner(lobby *models.Lobby) *models.Player {
	if len(lobby.Players) == 0 {
		return nil
	}

	best := &lobby.Players[0]
	for i := 1; i < len(lobby.Players); i++ {
		if lobby.Players[i].Score > best.Score {
			best = &lobby.Players[i]
		}
	}
	return best
}
