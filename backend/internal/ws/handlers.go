package ws

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/game"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// Handlers binds the websocket actions to the lobby manager
type Handlers struct {
	Manager models.LobbyManager
	Hub     *Hub
}

func NewHandlers(m models.LobbyManager, h *Hub) *Handlers {
	return &Handlers{Manager: m, Hub: h}
}

// BroadcastToLobby sends an event message to all connected clients regarding this lobby.
// Clients independently filter and process messages intended for their active lobby.
func (h *Handlers) BroadcastToLobby(action string, payload interface{}) {
	res := SocketResponse{
		Action:  action,
		Ok:      true,
		Payload: payload,
	}
	b, _ := json.Marshal(res)
	
	h.Hub.mu.RLock()
	for client := range h.Hub.clients {
		select {
		case client.send <- b:
		default:
		}
	}
	h.Hub.mu.RUnlock()
}

func (h *Handlers) Dispatch(c *Client, msg SocketMessage) {
	switch msg.Action {
	case "createLobby":
		h.handleCreateLobby(c, msg)
	case "joinLobby":
		h.handleJoinLobby(c, msg)
	case "addBot":
		h.handleAddBot(c, msg)
	case "startGame":
		h.handleStartGame(c, msg)
	case "chooseAttribute":
		h.handleChooseAttribute(c, msg)
	default:
		slog.Warn("Unknown ws action", "action", msg.Action)
	}
}

func (h *Handlers) handleCreateLobby(c *Client, msg SocketMessage) {
	var req struct {
		Name  string `json:"name"`
		Theme string `json:"theme"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.Reply(msg.ReqID, msg.Action, false, "Invalid payload", nil)
		return
	}

	playerID := "p_" + msg.ReqID
	if c.UserID != "" {
		playerID = c.UserID
	}

	host := models.Player{
		ID:     models.PlayerID(playerID),
		Name:   req.Name,
		UserID: c.UserID,
	}

	l, err := h.Manager.CreateLobby(req.Theme, host)
	if err != nil {
		c.Reply(msg.ReqID, msg.Action, false, err.Error(), nil)
		return
	}

	c.Reply(msg.ReqID, msg.Action, true, "", map[string]interface{}{"lobby": l})
}

func (h *Handlers) handleJoinLobby(c *Client, msg SocketMessage) {
	var req struct {
		LobbyID string `json:"lobbyId"`
		Name    string `json:"name"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.Reply(msg.ReqID, msg.Action, false, "Invalid payload", nil)
		return
	}

	playerID := "p_" + msg.ReqID
	if c.UserID != "" {
		playerID = c.UserID
	}

	p := models.Player{
		ID:     models.PlayerID(playerID),
		Name:   req.Name,
		UserID: c.UserID,
	}

	l, err := h.Manager.JoinLobby(req.LobbyID, p)
	if err != nil {
		c.Reply(msg.ReqID, msg.Action, false, err.Error(), nil)
		return
	}

	c.Reply(msg.ReqID, msg.Action, true, "", map[string]interface{}{"lobby": l})
	h.BroadcastToLobby("lobbyUpdate", l)
}

func (h *Handlers) handleAddBot(c *Client, msg SocketMessage) {
	var req struct {
		LobbyID string `json:"lobbyId"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.Reply(msg.ReqID, msg.Action, false, "Invalid payload", nil)
		return
	}

	l, err := h.Manager.AddBot(req.LobbyID)
	if err != nil {
		c.Reply(msg.ReqID, msg.Action, false, err.Error(), nil)
		return
	}

	c.Reply(msg.ReqID, msg.Action, true, "", map[string]interface{}{"lobby": l})
	h.BroadcastToLobby("lobbyUpdate", l)
}

func (h *Handlers) handleStartGame(c *Client, msg SocketMessage) {
	var req struct {
		LobbyID string `json:"lobbyId"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.Reply(msg.ReqID, msg.Action, false, "Invalid payload", nil)
		return
	}

	l, err := h.Manager.StartGame(req.LobbyID)
	if err != nil {
		c.Reply(msg.ReqID, msg.Action, false, err.Error(), nil)
		return
	}

	c.Reply(msg.ReqID, msg.Action, true, "", nil)
	h.BroadcastToLobby("gameStarted", l)
	
	// Trigger bot loop for the first turn if a bot was randomly chosen to start
	go h.processBotTurns(l, false)
}

func (h *Handlers) handleChooseAttribute(c *Client, msg SocketMessage) {
	var req struct {
		LobbyID  string `json:"lobbyId"`
		PlayerID string `json:"playerId"`
		Attr     string `json:"attr"`
	}
	if err := json.Unmarshal(msg.Payload, &req); err != nil {
		c.Reply(msg.ReqID, msg.Action, false, "Invalid payload", nil)
		return
	}

	_, res, err := h.Manager.PlayRound(req.LobbyID, models.PlayerID(req.PlayerID), req.Attr)
	if err != nil {
		c.Reply(msg.ReqID, msg.Action, false, err.Error(), nil)
		return
	}

	c.Reply(msg.ReqID, msg.Action, true, "", nil)
	
	// Broadcast round results
	h.BroadcastToLobby("roundResult", map[string]interface{}{
		"attr":      res.Attr,
		"winnerIds": res.WinnerIDs,
		"reveals":  res.Reveals,
		"lobby":    res.LobbyObj, // Include full updated state too
	})
	
	// Wait, standardizing roundResult payload so it mimics the old socket.js
	// the old socket.js expects: { attr, winnerId, reveals, lobby }

	// Trigger bot loop in the background if the next turn belongs to a bot
	go h.processBotTurns(res.LobbyObj, true)
}

func (h *Handlers) processBotTurns(lobby *models.Lobby, isContinuation bool) {
	for lobby.State == models.LobbyStatePlaying {
		currentPlayer := lobby.Players[lobby.CurrentPlayerIndex]
		if !currentPlayer.IsBot {
			break // human's turn, stop bot processing
		}

		// Wait briefly so the frontend has time to show animations for the previous round
		if isContinuation {
			time.Sleep(15 * time.Second)
		} else {
			time.Sleep(4 * time.Second)
		}
		isContinuation = true

		if !currentPlayer.HasCards() {
			break
		}

		botLogic := game.NewMaxStatBot()
		chosenAttr := botLogic.ChooseAttribute(currentPlayer.Hand[0])

		// Play the round
		_, res, err := h.Manager.PlayRound(lobby.ID, currentPlayer.ID, string(chosenAttr))
		if err != nil {
			slog.Error("Bot failed to play round", "err", err)
			break
		}

		h.BroadcastToLobby("roundResult", map[string]interface{}{
			"attr":      res.Attr,
			"winnerIds": res.WinnerIDs,
			"reveals":  res.Reveals,
			"lobby":    res.LobbyObj,
		})

		lobby = res.LobbyObj
	}
}
