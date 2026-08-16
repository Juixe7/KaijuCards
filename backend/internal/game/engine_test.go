package game

import (
	"testing"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

func TestEngine_Deal(t *testing.T) {
	engine := NewEngine()

	deck := models.Deck{
		{ID: "1", Name: "Card 1"},
		{ID: "2", Name: "Card 2"},
		{ID: "3", Name: "Card 3"},
		{ID: "4", Name: "Card 4"},
		{ID: "5", Name: "Card 5"},
		{ID: "6", Name: "Card 6"},
		{ID: "7", Name: "Card 7"},
		{ID: "8", Name: "Card 8"},
		{ID: "9", Name: "Card 9"},
		{ID: "10", Name: "Card 10"},
		{ID: "11", Name: "Card 11"},
		{ID: "12", Name: "Card 12"},
	}

	lobby := &models.Lobby{
		State: models.LobbyStateWaiting,
		Players: []models.Player{
			{ID: "p1", Name: "Player 1"},
			{ID: "p2", Name: "Player 2"},
		},
		Deck: deck,
	}

	err := engine.Deal(lobby)
	if err != nil {
		t.Fatalf("unexpected error dealing cards: %v", err)
	}

	if lobby.State != models.LobbyStatePlaying {
		t.Errorf("expected lobby state to be playing, got %s", lobby.State)
	}

	if lobby.Round != 1 {
		t.Errorf("expected round to be 1, got %d", lobby.Round)
	}

	for _, p := range lobby.Players {
		if len(p.Hand) != models.CardsPerPlayer {
			t.Errorf("expected player %s to have %d cards, got %d", p.Name, models.CardsPerPlayer, len(p.Hand))
		}
	}
}

func TestEngine_Deal_InvalidState(t *testing.T) {
	engine := NewEngine()
	lobby := &models.Lobby{
		State: models.LobbyStatePlaying,
	}

	err := engine.Deal(lobby)
	if err == nil {
		t.Error("expected error dealing in playing state, got nil")
	}
}

func TestEngine_Deal_NotEnoughPlayers(t *testing.T) {
	engine := NewEngine()
	lobby := &models.Lobby{
		State: models.LobbyStateWaiting,
		Players: []models.Player{
			{ID: "p1", Name: "Player 1"},
		},
	}

	err := engine.Deal(lobby)
	if err == nil {
		t.Error("expected error dealing with 1 player, got nil")
	}
}

func TestEngine_ResolveRound(t *testing.T) {
	engine := NewEngine()

	lobby := &models.Lobby{
		State: models.LobbyStatePlaying,
		Players: []models.Player{
			{
				ID:   "p1",
				Name: "Player 1",
				Hand: models.Deck{
					{ID: "1", Name: "Card 1", Stats: models.Stats{Strength: 90}},
				},
			},
			{
				ID:   "p2",
				Name: "Player 2",
				Hand: models.Deck{
					{ID: "2", Name: "Card 2", Stats: models.Stats{Strength: 80}},
				},
			},
		},
		Round: 1,
	}

	res, err := engine.ResolveRound(lobby, models.AttrStrength)
	if err != nil {
		t.Fatalf("unexpected error resolving round: %v", err)
	}

	if len(res.WinnerIDs) != 1 || res.WinnerIDs[0] != "p1" {
		t.Errorf("expected winner to be p1, got %v", res.WinnerIDs)
	}

	if lobby.Players[0].Score != 1 {
		t.Errorf("expected Player 1 score to be 1, got %d", lobby.Players[0].Score)
	}

	if len(lobby.Players[0].Hand) != 0 || len(lobby.Players[1].Hand) != 0 {
		t.Error("expected top cards to be removed from hands")
	}

	if lobby.State != models.LobbyStateFinished {
		t.Errorf("expected game to finish, got state %s", lobby.State)
	}
}

func TestEngine_ResolveRound_Tie(t *testing.T) {
	engine := NewEngine()

	lobby := &models.Lobby{
		State: models.LobbyStatePlaying,
		Players: []models.Player{
			{
				ID:   "p1",
				Name: "Player 1",
				Hand: models.Deck{
					{ID: "1", Name: "Card 1", Stats: models.Stats{IQ: 50}},
					{ID: "3", Name: "Card 3", Stats: models.Stats{IQ: 70}},
				},
			},
			{
				ID:   "p2",
				Name: "Player 2",
				Hand: models.Deck{
					{ID: "2", Name: "Card 2", Stats: models.Stats{IQ: 50}},
					{ID: "4", Name: "Card 4", Stats: models.Stats{IQ: 60}},
				},
			},
		},
		Round: 1,
	}

	res, err := engine.ResolveRound(lobby, models.AttrIQ)
	if err != nil {
		t.Fatalf("unexpected error resolving round: %v", err)
	}

	if len(res.WinnerIDs) != 2 {
		t.Errorf("expected a tie with 2 winners, got %d winners", len(res.WinnerIDs))
	}

	if lobby.Players[0].Score != 1 || lobby.Players[1].Score != 1 {
		t.Errorf("expected tie: both players should score 1, got scores p1=%d, p2=%d", lobby.Players[0].Score, lobby.Players[1].Score)
	}

	if lobby.State != models.LobbyStatePlaying {
		t.Errorf("expected game to remain playing, got state %s", lobby.State)
	}

	if lobby.Round != 2 {
		t.Errorf("expected round to advance to 2, got %d", lobby.Round)
	}
}

func TestEngine_DetermineWinner(t *testing.T) {
	engine := NewEngine()

	lobby := &models.Lobby{
		Players: []models.Player{
			{ID: "p1", Name: "Player 1", Score: 2},
			{ID: "p2", Name: "Player 2", Score: 4},
			{ID: "p3", Name: "Player 3", Score: 1},
		},
	}

	winner := engine.DetermineWinner(lobby)
	if winner == nil || winner.ID != "p2" {
		t.Errorf("expected winner p2, got %v", winner)
	}
}
