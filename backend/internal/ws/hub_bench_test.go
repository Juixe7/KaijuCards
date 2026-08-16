package ws

import (
	"encoding/json"
	"sync"
	"testing"
)

// mockClient simulates a connected WebSocket client with a buffered send channel.
func mockClient(hub *Hub) *Client {
	return &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}
}

// BenchmarkHubBroadcast_10 measures broadcasting to 10 clients.
func BenchmarkHubBroadcast_10(b *testing.B) {
	benchBroadcast(b, 10)
}

// BenchmarkHubBroadcast_100 measures broadcasting to 100 clients.
func BenchmarkHubBroadcast_100(b *testing.B) {
	benchBroadcast(b, 100)
}

// BenchmarkHubBroadcast_1000 measures broadcasting to 1000 clients.
func BenchmarkHubBroadcast_1000(b *testing.B) {
	benchBroadcast(b, 1000)
}

func benchBroadcast(b *testing.B, numClients int) {
	hub := NewHub()

	// Register mock clients directly (no hub.Run needed for direct map access)
	clients := make([]*Client, numClients)
	for i := 0; i < numClients; i++ {
		c := mockClient(hub)
		clients[i] = c
		hub.mu.Lock()
		hub.clients[c] = true
		hub.mu.Unlock()
	}

	// Drain channels in the background so sends don't block
	var wg sync.WaitGroup
	for _, c := range clients {
		wg.Add(1)
		go func(ch chan []byte) {
			defer wg.Done()
			for range ch {
			}
		}(c.send)
	}

	msg := SocketResponse{
		Action:  "roundResult",
		Ok:      true,
		Payload: map[string]interface{}{"round": 1, "winnerId": "p1"},
	}
	msgBytes, _ := json.Marshal(msg)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hub.mu.RLock()
		for client := range hub.clients {
			select {
			case client.send <- msgBytes:
			default:
			}
		}
		hub.mu.RUnlock()
	}

	// Cleanup
	b.StopTimer()
	for _, c := range clients {
		close(c.send)
	}
	wg.Wait()
}

// BenchmarkMessageSerialization measures JSON marshal of a typical game event.
func BenchmarkMessageSerialization(b *testing.B) {
	msg := SocketResponse{
		Action: "roundResult",
		Ok:     true,
		Payload: map[string]interface{}{
			"round":     3,
			"winnerId":  "p2",
			"attr":      "strength",
			"reveals":   []map[string]interface{}{{"name": "Luffy", "stats": map[string]int{"rank": 95, "strength": 99, "speed": 88, "iq": 60}}},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(msg)
	}
}
