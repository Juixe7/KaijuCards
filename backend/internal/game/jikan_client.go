package game

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// JikanClient handles outbound requests to the Jikan API with an in-memory cache
// to prevent rate-limiting on popular themes.
type JikanClient struct {
	baseURL string
	client  *http.Client
	geminiAPIKey string
	
	// Simple thread-safe cache: Theme Name -> Deck
	cacheMu sync.RWMutex
	cache   map[string]cacheEntry
}

type cacheEntry struct {
	deck      models.Deck
	timestamp time.Time
}

// NewJikanClient creates a new Jikan API client.
func NewJikanClient(baseURL string, timeout time.Duration, geminiKey string) *JikanClient {
	return &JikanClient{
		baseURL: baseURL,
		client:  &http.Client{Timeout: timeout},
		geminiAPIKey: geminiKey,
		cache:   make(map[string]cacheEntry),
	}
}

// FetchDeck searches for an anime by theme and returns a deck of its characters.
// It checks the cache first.
func (c *JikanClient) FetchDeck(theme string) (models.Deck, error) {
	themeKey := strings.ToLower(strings.TrimSpace(theme))

	// 1. Check cache (Valid for 1 hour)
	c.cacheMu.RLock()
	if entry, ok := c.cache[themeKey]; ok {
		if time.Since(entry.timestamp) < time.Hour {
			c.cacheMu.RUnlock()
			return copyDeck(entry.deck), nil
		}
	}
	c.cacheMu.RUnlock()

	// 2. Fetch from Jikan
	deck, err := c.fetchFromAPI(theme)
	if err != nil {
		return nil, err
	}

	// 3. Save to cache
	c.cacheMu.Lock()
	// Basic cache eviction if it grows too large (prevent memory leaks)
	if len(c.cache) > 100 {
		c.cache = make(map[string]cacheEntry) // clear all for simplicity
	}
	c.cache[themeKey] = cacheEntry{
		deck:      deck,
		timestamp: time.Now(),
	}
	c.cacheMu.Unlock()

	return copyDeck(deck), nil
}

func (c *JikanClient) fetchFromAPI(theme string) (models.Deck, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.client.Timeout)
	defer cancel()

	// Step 1: Search Anime
	searchURL := fmt.Sprintf("%s/anime?q=%s&limit=1", c.baseURL, url.QueryEscape(theme))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("search request error: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jikan api connection failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search returned status %d", resp.StatusCode)
	}

	var searchResp jikanSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, fmt.Errorf("failed to decode search JSON: %w", err)
	}
	if len(searchResp.Data) == 0 {
		return nil, fmt.Errorf("no anime found for %q", theme)
	}

	// Step 2: Fetch Characters
	malID := searchResp.Data[0].MalID
	charsURL := fmt.Sprintf("%s/anime/%d/characters", c.baseURL, malID)
	
	req2, err := http.NewRequestWithContext(ctx, http.MethodGet, charsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("chars request error: %w", err)
	}

	// Be nice to the API
	time.Sleep(500 * time.Millisecond)

	resp2, err := c.client.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("jikan api chars connection failed: %w", err)
	}
	defer resp2.Body.Close()

	if resp2.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("rate limited (429) on characters")
	}
	if resp2.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("characters returned status %d", resp2.StatusCode)
	}

	var charsResp jikanCharactersResponse
	if err := json.NewDecoder(resp2.Body).Decode(&charsResp); err != nil {
		return nil, fmt.Errorf("failed to decode chars JSON: %w", err)
	}
	
	// Build the deck
	seen := make(map[string]bool)
	var deck models.Deck
	var names []string

	type charInfo struct {
		Name string
		Image string
		Favorites int
	}
	var chars []charInfo

	for _, c := range charsResp.Data {
		name := c.Character.Name
		if seen[name] {
			continue
		}
		seen[name] = true

		image := c.Character.Images.JPG.ImageURL
		if image == "" {
			image = fmt.Sprintf("https://picsum.photos/seed/%s/320/420", url.QueryEscape(name))
		}

		chars = append(chars, charInfo{Name: name, Image: image, Favorites: c.Favorites})
		names = append(names, name)

		if len(chars) >= models.TotalCards {
			break
		}
	}

	var llmStats map[string]models.Stats
	if c.geminiAPIKey != "" {
		llmStats = fetchLLMStats(ctx, theme, names, c.geminiAPIKey)
	}

	for _, ch := range chars {
		stats, ok := llmStats[ch.Name]
		if !ok {
			baseRank := ch.Favorites % 50 
			stats = generateDeterministicStats(ch.Name)
			stats.Rank = min(99, stats.Rank + baseRank)
		}

		deck = append(deck, models.Card{
			ID:    generateCardID(ch.Name),
			Name:  ch.Name,
			Image: ch.Image,
			Stats: stats,
		})
	}

	if len(deck) < models.TotalCards {
		return nil, fmt.Errorf("not enough characters found (%d/%d)", len(deck), models.TotalCards)
	}

	return deck, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func copyDeck(d models.Deck) models.Deck {
	cp := make(models.Deck, len(d))
	copy(cp, d)
	return cp
}

// JSON Structs
type jikanSearchResponse struct {
	Data []struct {
		MalID int `json:"mal_id"`
	} `json:"data"`
}

type jikanCharactersResponse struct {
	Data []struct {
		Favorites int `json:"favorites"`
		Character struct {
			Name   string `json:"name"`
			Images struct {
				JPG struct {
					ImageURL string `json:"image_url"`
				} `json:"jpg"`
			} `json:"images"`
		} `json:"character"`
	} `json:"data"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
	GenerationConfig struct {
		ResponseMimeType string `json:"response_mime_type"`
	} `json:"generationConfig"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func fetchLLMStats(ctx context.Context, theme string, names []string, apiKey string) map[string]models.Stats {
	prompt := fmt.Sprintf(`Assign lore-accurate power stats for these characters from the universe '%s'. 
Stats must be integers from 1 to 99 for: Rank (overall power), Strength, Speed, IQ.
Respond ONLY with a JSON object mapping the exact character name to the stats object.
Example: {"Goku": {"rank": 99, "strength": 99, "speed": 99, "iq": 70}}

Characters:
%s`, theme, strings.Join(names, "\n"))

	reqBody := geminiRequest{}
	reqBody.Contents = append(reqBody.Contents, geminiContent{
		Parts: []geminiPart{{Text: prompt}},
	})
	reqBody.GenerationConfig.ResponseMimeType = "application/json"

	b, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		slog.Error("LLM req creation failed", "err", err)
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("LLM API call failed", "err", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		slog.Error("LLM API error", "status", resp.StatusCode)
		return nil
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text
	var result map[string]models.Stats
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		slog.Error("LLM JSON parse failed", "err", err, "text", text)
		return nil
	}

	return result
}
