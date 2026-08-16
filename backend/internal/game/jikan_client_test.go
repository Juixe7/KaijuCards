package game

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

func TestJikanClient_FetchDeck_Success(t *testing.T) {
	// Mock Jikan API Server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/anime" {
			json.NewEncoder(w).Encode(jikanSearchResponse{
				Data: []struct {
					MalID int `json:"mal_id"`
				}{{MalID: 1}},
			})
			return
		}

		if r.URL.Path == "/anime/1/characters" {
			chars := jikanCharactersResponse{}
			for i := 0; i < models.TotalCards; i++ {
				chars.Data = append(chars.Data, struct {
					Favorites int `json:"favorites"`
					Character struct {
						Name   string `json:"name"`
						Images struct {
							JPG struct {
								ImageURL string `json:"image_url"`
							} `json:"jpg"`
						} `json:"images"`
					} `json:"character"`
				}{
					Favorites: i * 5,
					Character: struct {
						Name   string `json:"name"`
						Images struct {
							JPG struct {
								ImageURL string `json:"image_url"`
							} `json:"jpg"`
						} `json:"images"`
					}{
						Name: "Test Char", // Will be deduplicated if names are the same, so let's make them unique
					},
				})
				chars.Data[i].Character.Name = "Char " + string(rune('A'+i))
			}
			json.NewEncoder(w).Encode(chars)
			return
		}
		
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := NewJikanClient(server.URL, 2*time.Second, "")

	// Fetch 1st time (Cache Miss)
	deck, err := client.FetchDeck("test theme")
	if err != nil {
		t.Fatalf("FetchDeck failed: %v", err)
	}

	if len(deck) != models.TotalCards {
		t.Errorf("Expected %d cards, got %d", models.TotalCards, len(deck))
	}

	// Fetch 2nd time (Cache Hit)
	// Even if we shut down the server, it should return from cache
	server.Close()
	
	deck2, err := client.FetchDeck("test theme")
	if err != nil {
		t.Fatalf("FetchDeck (cached) failed: %v", err)
	}
	
	if len(deck2) != models.TotalCards {
		t.Errorf("Expected %d cards from cache, got %d", models.TotalCards, len(deck2))
	}
}

func TestJikanClient_FetchDeck_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewJikanClient(server.URL, 1*time.Second, "")
	_, err := client.FetchDeck("naruto")
	
	if err == nil {
		t.Fatal("Expected error on 429 Rate Limit, got nil")
	}
	if err.Error() != "rate limited (429)" {
		t.Errorf("Expected 'rate limited (429)', got %v", err)
	}
}

func TestDeterministicStats(t *testing.T) {
	stats1 := generateDeterministicStats("Naruto")
	stats2 := generateDeterministicStats("Naruto")
	stats3 := generateDeterministicStats("Sasuke")

	if stats1.Rank != stats2.Rank || stats1.Strength != stats2.Strength {
		t.Error("Deterministic stats failed to produce identical results for the same name")
	}
	
	if stats1.Rank == stats3.Rank && stats1.Strength == stats3.Strength {
		t.Error("Deterministic stats produced identical results for different names")
	}
}
