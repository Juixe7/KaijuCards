package lobby

import (
	"errors"
	"testing"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/game"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

type mockUserRepo struct {
	savedMatch  *models.Match
	savedPlayers []models.MatchPlayer
}

func (m *mockUserRepo) CreateUser(user *models.User) error {
	return nil
}

func (m *mockUserRepo) GetByID(id string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) GetByUsername(username string) (*models.User, error) {
	return nil, nil
}

func (m *mockUserRepo) SaveMatch(match *models.Match, players []models.MatchPlayer) error {
	m.savedMatch = match
	m.savedPlayers = players
	return nil
}

func (m *mockUserRepo) GetMatchHistory(userID string, limit int) ([]models.MatchSummary, error) {
	return nil, nil
}

func (m *mockUserRepo) GetWinLoss(userID string) (wins, losses int, err error) {
	return 0, 0, nil
}

type mockGenerator struct {
	shouldErr bool
}

func (g *mockGenerator) Generate(theme string) (models.Deck, models.CardSource, error) {
	if g.shouldErr {
		return nil, models.CardSourceGenerate, errors.New("mock generator error")
	}
	deck := make(models.Deck, models.TotalCards)
	for i := range deck {
		deck[i] = models.Card{
			ID:    string(rune('1' + i)),
			Name:  "Card",
			Stats: models.Stats{Strength: 50, IQ: 50, Speed: 50, Rank: 50},
		}
	}
	return deck, models.CardSourceGenerate, nil
}

func TestManager_CreateLobby(t *testing.T) {
	store := NewMemoryStore()
	gen := &mockGenerator{}
	engine := game.NewEngine()
	repo := &mockUserRepo{}
	mgr := NewManager(store, gen, engine, repo)

	host := models.Player{ID: "p1", Name: "Host"}
	lobby, err := mgr.CreateLobby("one piece", host)
	if err != nil {
		t.Fatalf("CreateLobby failed: %v", err)
	}

	if lobby.Theme != "one piece" {
		t.Errorf("expected theme 'one piece', got %s", lobby.Theme)
	}

	if lobby.State != models.LobbyStateWaiting {
		t.Errorf("expected state to be waiting, got %s", lobby.State)
	}

	if len(lobby.Players) != 1 || lobby.Players[0].Name != "Host" {
		t.Errorf("expected lobby to have host, got players: %v", lobby.Players)
	}
}

func TestManager_JoinLobby(t *testing.T) {
	store := NewMemoryStore()
	gen := &mockGenerator{}
	engine := game.NewEngine()
	repo := &mockUserRepo{}
	mgr := NewManager(store, gen, engine, repo)

	host := models.Player{ID: "p1", Name: "Host"}
	lobby, _ := mgr.CreateLobby("one piece", host)

	guest := models.Player{ID: "p2", Name: "Guest", SocketID: "s2"}
	updatedLobby, err := mgr.JoinLobby(lobby.ID, guest)
	if err != nil {
		t.Fatalf("JoinLobby failed: %v", err)
	}

	if len(updatedLobby.Players) != 2 {
		t.Errorf("expected 2 players, got %d", len(updatedLobby.Players))
	}

	// Test reconnection (join same player name, update SocketID)
	reconnectedGuest := models.Player{ID: "p2", Name: "Guest", SocketID: "s_new"}
	reconnectedLobby, err := mgr.JoinLobby(lobby.ID, reconnectedGuest)
	if err != nil {
		t.Fatalf("Reconnection JoinLobby failed: %v", err)
	}

	if len(reconnectedLobby.Players) != 2 {
		t.Errorf("expected players count to remain 2, got %d", len(reconnectedLobby.Players))
	}

	if reconnectedLobby.Players[1].SocketID != "s_new" {
		t.Errorf("expected socket ID to be updated to 's_new', got %s", reconnectedLobby.Players[1].SocketID)
	}
}

func TestManager_AddBot(t *testing.T) {
	store := NewMemoryStore()
	gen := &mockGenerator{}
	engine := game.NewEngine()
	repo := &mockUserRepo{}
	mgr := NewManager(store, gen, engine, repo)

	host := models.Player{ID: "p1", Name: "Host"}
	lobby, _ := mgr.CreateLobby("one piece", host)

	updatedLobby, err := mgr.AddBot(lobby.ID)
	if err != nil {
		t.Fatalf("AddBot failed: %v", err)
	}

	if len(updatedLobby.Players) != 2 || !updatedLobby.Players[1].IsBot {
		t.Errorf("expected second player to be a bot, got: %v", updatedLobby.Players)
	}
}

func TestManager_StartGame(t *testing.T) {
	store := NewMemoryStore()
	gen := &mockGenerator{}
	engine := game.NewEngine()
	repo := &mockUserRepo{}
	mgr := NewManager(store, gen, engine, repo)

	host := models.Player{ID: "p1", Name: "Host"}
	lobby, _ := mgr.CreateLobby("one piece", host)

	// Need at least 2 players to start
	_, err := mgr.StartGame(lobby.ID)
	if err == nil {
		t.Error("expected start game to fail with 1 player, got nil")
	}

	_, _ = mgr.AddBot(lobby.ID)

	startedLobby, err := mgr.StartGame(lobby.ID)
	if err != nil {
		t.Fatalf("StartGame failed: %v", err)
	}

	if startedLobby.State != models.LobbyStatePlaying {
		t.Errorf("expected lobby state playing, got %s", startedLobby.State)
	}

	if len(startedLobby.Players[0].Hand) != models.CardsPerPlayer {
		t.Errorf("expected dealt hand size %d, got %d", models.CardsPerPlayer, len(startedLobby.Players[0].Hand))
	}
}

func TestManager_PlayRound_MatchHistory(t *testing.T) {
	store := NewMemoryStore()
	// Create generator that returns cards so that 1 card is dealt per player
	// We want game to finish after 1 round to test match history persistence
	// Wait, models.TotalCards is models.CardsPerPlayer * models.MaxPlayers = 24.
	// But we can customize hands directly on the lobby object after dealing or creating!
	gen := &mockGenerator{}
	engine := game.NewEngine()
	repo := &mockUserRepo{}
	mgr := NewManager(store, gen, engine, repo)

	hostID := "00000000-0000-0000-0000-000000000001"
	host := models.Player{ID: models.PlayerID(hostID), Name: "Host", UserID: hostID}
	lobby, _ := mgr.CreateLobby("one piece", host)
	_, _ = mgr.AddBot(lobby.ID)

	// Deal
	lobby, _ = mgr.StartGame(lobby.ID)

	// Manually set hands to 1 card each to trigger LobbyStateFinished on round resolve
	lobby.Players[0].Hand = models.Deck{{ID: "1", Name: "C1", Stats: models.Stats{Strength: 90}}}
	lobby.Players[1].Hand = models.Deck{{ID: "2", Name: "C2", Stats: models.Stats{Strength: 80}}}
	_ = store.Update(lobby)

	// Play the round
	updatedLobby, res, err := mgr.PlayRound(lobby.ID, models.PlayerID(hostID), "strength")
	if err != nil {
		t.Fatalf("PlayRound failed: %v", err)
	}

	if updatedLobby.State != models.LobbyStateFinished {
		t.Errorf("expected game to finish, got state %s", updatedLobby.State)
	}

	if res.WinnerIDs[0] != models.PlayerID(hostID) {
		t.Errorf("expected winner to be %s, got %v", hostID, res.WinnerIDs)
	}

	// Verify SaveMatch was called
	if repo.savedMatch == nil {
		t.Fatal("expected match to be saved, got nil")
	}

	if repo.savedMatch.LobbyCode != lobby.ID || repo.savedMatch.WinnerID != hostID {
		t.Errorf("incorrect saved match details: %+v", repo.savedMatch)
	}

	if len(repo.savedPlayers) != 2 || repo.savedPlayers[0].UserID != hostID || repo.savedPlayers[0].Score != 1 {
		t.Errorf("incorrect saved match players details: %+v", repo.savedPlayers)
	}
}
