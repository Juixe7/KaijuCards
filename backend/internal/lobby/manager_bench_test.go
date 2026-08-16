package lobby

import (
	"testing"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/game"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// BenchmarkCreateLobby measures lobby creation with deterministic card generation.
func BenchmarkCreateLobby(b *testing.B) {
	store := NewMemoryStore()
	gen := game.NewGenerator("https://api.jikan.moe/v4", 0, "", nil)
	engine := game.NewEngine()
	mgr := NewManager(store, gen, engine, nil)

	host := models.Player{ID: "host-1", Name: "BenchHost"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = mgr.CreateLobby("benchmark-theme", host)
	}
}

// BenchmarkLobbyLifecycle measures a full create → join bots → start → play 6 rounds lifecycle.
func BenchmarkLobbyLifecycle(b *testing.B) {
	gen := game.NewGenerator("https://api.jikan.moe/v4", 0, "", nil)
	engine := game.NewEngine()
	bot := game.NewMaxStatBot()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store := NewMemoryStore()
		mgr := NewManager(store, gen, engine, nil)

		host := models.Player{ID: "host-1", Name: "BenchHost"}
		lobby, _ := mgr.CreateLobby("benchmark", host)

		// Add 3 bots
		for j := 0; j < 3; j++ {
			_, _ = mgr.AddBot(lobby.ID)
		}

		// Start
		lobby, _ = mgr.StartGame(lobby.ID)

		// Play all rounds
		for lobby.State == models.LobbyStatePlaying {
			cp := lobby.Players[lobby.CurrentPlayerIndex]
			if !cp.HasCards() {
				break
			}
			attr := bot.ChooseAttribute(cp.Hand[0])
			lobby, _, _ = mgr.PlayRound(lobby.ID, cp.ID, string(attr))
		}
	}
}

// BenchmarkConcurrentLobbies measures parallel lobby creation under contention.
func BenchmarkConcurrentLobbies(b *testing.B) {
	store := NewMemoryStore()
	gen := game.NewGenerator("https://api.jikan.moe/v4", 0, "", nil)
	engine := game.NewEngine()
	mgr := NewManager(store, gen, engine, nil)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			host := models.Player{ID: models.PlayerID("host-" + time.Now().String()), Name: "P"}
			_, _ = mgr.CreateLobby("parallel-bench", host)
			i++
		}
	})
}

// BenchmarkMemoryStoreGet measures read performance on the in-memory store.
func BenchmarkMemoryStoreGet(b *testing.B) {
	store := NewMemoryStore()
	// Pre-populate
	lobby := &models.Lobby{
		Theme:   "bench",
		State:   models.LobbyStateWaiting,
		Players: []models.Player{{ID: "p1", Name: "Player1"}},
	}
	code, _ := store.Create(lobby)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = store.Get(code)
	}
}
