package ws

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/Juixe7/KaijuCards/backend/internal/auth"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 4096
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow cross-origin requests for flexible frontend deployments
	},
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte

	// Reference to the global handlers instance
	handlers *Handlers

	// The authenticated user's ID, or empty for guests.
	UserID string
}

// SocketMessage defines the generic format for requests from frontend
type SocketMessage struct {
	Action  string          `json:"action"`
	ReqID   string          `json:"reqId,omitempty"` // For callbacks
	Payload json.RawMessage `json:"payload"`
}

// SocketResponse defines the generic format for replies to frontend
type SocketResponse struct {
	Action  string      `json:"action"`
	ReqID   string      `json:"reqId,omitempty"` // To resolve frontend promises
	Ok      bool        `json:"ok"`
	Error   string      `json:"error,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func (c *Client) Reply(reqID, action string, ok bool, err string, payload interface{}) {
	if reqID == "" && action == "" {
		return
	}
	res := SocketResponse{
		Action:  action,
		ReqID:   reqID,
		Ok:      ok,
		Error:   err,
		Payload: payload,
	}
	b, _ := json.Marshal(res)
	
	// Ensure we only queue if channel isn't closed
	defer func() {
		if r := recover(); r != nil {
			slog.Warn("Attempted to send on closed channel")
		}
	}()
	c.send <- b
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Error("readPump error", "error", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		
		// Parse standard message format
		var msg SocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			slog.Error("Invalid message format", "error", err)
			continue
		}

		// Dispatch to handlers
		c.handlers.Dispatch(c, msg)
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// ServeWs handles websocket requests from the peer.
func ServeWs(hub *Hub, handlers *Handlers, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("WS Upgrade error", "error", err)
		return
	}
	
	// Extract user ID if the client is authenticated
	userID := auth.UserIDFromContext(r.Context())

	client := &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, 256),
		handlers: handlers,
		UserID:   userID,
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
