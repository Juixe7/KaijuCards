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
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// geminiCharacter represents a single character returned by Gemini full-generation.
type geminiCharacter struct {
	Name     string `json:"name"`
	Rank     int    `json:"rank"`
	Strength int    `json:"strength"`
	Speed    int    `json:"speed"`
	IQ       int    `json:"iq"`
}

// fetchGeminiFullDeck asks Gemini to generate character names + lore-accurate stats
// for any fictional universe. This is the universal fallback when neither Jikan
// (anime) nor the Superhero API (comics) can resolve the theme.
//
// Returns a complete deck of TotalCards characters, or an error if generation fails.
func fetchGeminiFullDeck(theme string, apiKey string) (models.Deck, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("gemini API key not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	prompt := fmt.Sprintf(`You are a fictional universe expert. Generate exactly %d well-known characters from the universe "%s".

For each character provide these fields:
- "name": the character's most recognizable canonical name (no duplicates)
- "rank": overall power level (integer 1-99, relative to this universe)
- "strength": physical strength (integer 1-99)
- "speed": agility and speed (integer 1-99)
- "iq": intelligence and tactical ability (integer 1-99)

Rules:
- Pick the most iconic and recognizable characters from this universe
- Stats must be lore-accurate and differentiated (don't make everyone similar)
- Ensure variety: include both powerful and weaker characters
- Names must be the commonly known English names

Respond ONLY with a JSON array of objects. No markdown, no explanation.
Example: [{"name": "Character Name", "rank": 85, "strength": 80, "speed": 70, "iq": 88}]`, models.TotalCards, theme)

	reqBody := geminiRequest{}
	reqBody.Contents = append(reqBody.Contents, geminiContent{
		Parts: []geminiPart{{Text: prompt}},
	})
	reqBody.GenerationConfig.ResponseMimeType = "application/json"

	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal gemini request: %w", err)
	}

	apiURL := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(b))
	if err != nil {
		return nil, fmt.Errorf("failed to create gemini request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini API call failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("gemini API returned status %d", resp.StatusCode)
	}

	var geminiResp geminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode gemini response: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned empty response")
	}

	text := geminiResp.Candidates[0].Content.Parts[0].Text

	// Parse the JSON array of characters
	var chars []geminiCharacter
	if err := json.Unmarshal([]byte(text), &chars); err != nil {
		slog.Error("gemini full-gen JSON parse failed", "err", err, "text", text)
		return nil, fmt.Errorf("failed to parse gemini character data: %w", err)
	}

	if len(chars) < models.TotalCards {
		return nil, fmt.Errorf("gemini returned only %d characters, need %d", len(chars), models.TotalCards)
	}

	// Build deck from Gemini-generated characters
	seen := make(map[string]bool)
	var deck models.Deck

	for _, ch := range chars {
		name := strings.TrimSpace(ch.Name)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true

		// Use deterministic placeholder images seeded by character name
		image := fmt.Sprintf("https://picsum.photos/seed/%s/320/420", url.QueryEscape(name))

		deck = append(deck, models.Card{
			ID:    generateCardID(name),
			Name:  name,
			Image: image,
			Stats: models.Stats{
				Rank:     clampStat(ch.Rank),
				Strength: clampStat(ch.Strength),
				Speed:    clampStat(ch.Speed),
				IQ:       clampStat(ch.IQ),
			},
		})

		if len(deck) >= models.TotalCards {
			break
		}
	}

	if len(deck) < models.TotalCards {
		return nil, fmt.Errorf("not enough unique characters after dedup (%d/%d)", len(deck), models.TotalCards)
	}

	slog.Info("gemini full-gen succeeded", "theme", theme, "cards", len(deck))
	return deck, nil
}
