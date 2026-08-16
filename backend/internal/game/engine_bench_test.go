package game

import (
	"testing"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// setupLobby creates a ready-to-deal lobby with 4 players and a deterministic deck.
func setupLobby() *models.Lobby {
	gen := &Generator{jikanClient: NewJikanClient("https://api.jikan.moe/v4", 0, "")}
	deck := gen.generateDeterministic("benchmark")

	lobby := &models.Lobby{
		ID:    "BENCH",
		Theme: "benchmark",
		State: models.LobbyStateWaiting,
		Players: []models.Player{
			{ID: "p1", Name: "Player1"},
			{ID: "p2", Name: "Player2", IsBot: true},
			{ID: "p3", Name: "Player3", IsBot: true},
			{ID: "p4", Name: "Player4", IsBot: true},
		},
		Deck:       deck,
		MaxPlayers: 4,
		Round:      1,
		History:    []models.RoundResult{},
	}
	return lobby
}

// BenchmarkDeal measures the time to shuffle and deal 24 cards across 4 players.
func BenchmarkDeal(b *testing.B) {
	engine := NewEngine()
	gen := &Generator{jikanClient: NewJikanClient("https://api.jikan.moe/v4", 0, "")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lobby := setupLobby()
		lobby.Deck = gen.generateDeterministic("benchmark")
		_ = engine.Deal(lobby)
	}
}

// BenchmarkResolveRound measures a single round resolution (compare + score + advance).
func BenchmarkResolveRound(b *testing.B) {
	engine := NewEngine()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		lobby := setupLobby()
		_ = engine.Deal(lobby)
		b.StartTimer()

		_, _ = engine.ResolveRound(lobby, models.AttrStrength)
	}
}

// BenchmarkFullGame measures a complete 6-round match from deal to winner.
func BenchmarkFullGame(b *testing.B) {
	engine := NewEngine()
	bot := NewMaxStatBot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lobby := setupLobby()
		_ = engine.Deal(lobby)

		for lobby.State == models.LobbyStatePlaying {
			currentPlayer := lobby.Players[lobby.CurrentPlayerIndex]
			if !currentPlayer.HasCards() {
				break
			}
			attr := bot.ChooseAttribute(currentPlayer.Hand[0])
			_, _ = engine.ResolveRound(lobby, attr)
		}
	}
}

// BenchmarkDetermineWinner measures winner selection from scored players.
func BenchmarkDetermineWinner(b *testing.B) {
	engine := NewEngine()
	lobby := setupLobby()
	lobby.Players[0].Score = 3
	lobby.Players[1].Score = 1
	lobby.Players[2].Score = 5
	lobby.Players[3].Score = 2

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = engine.DetermineWinner(lobby)
	}
}
