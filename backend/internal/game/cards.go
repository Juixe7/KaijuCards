// Package game implements the Meta Clash game engine, card generation,
// and bot AI logic.
package game

import (
	"fmt"
	"hash/fnv"
	"log/slog"
	"math/rand"
	"net/url"
	"time"

	"github.com/nikhilsaxena04/meta_clash/backend/internal/game/packs"
	"github.com/nikhilsaxena04/meta_clash/backend/internal/models"
)

// Generator implements models.CardGenerator with a 5-tier fallback chain:
//  1. Curated packs — if theme matches a hardcoded pack (One Piece, Pokemon)
//  2. Jikan API — fetch characters for anime themes
//  3. Superhero API — fetch characters for Marvel/DC/comics themes
//  4. Gemini full-generation — AI generates names + stats for any universe
//  5. Deterministic hash — generate placeholder cards with FNV-based stable stats
type Generator struct {
	jikanClient     *JikanClient
	superheroClient *SuperheroClient
	geminiAPIKey    string
}

// NewGenerator creates a CardGenerator with the given API clients.
func NewGenerator(jikanBaseURL string, timeout time.Duration, geminiAPIKey string, superheroClient *SuperheroClient) *Generator {
	return &Generator{
		jikanClient:     NewJikanClient(jikanBaseURL, timeout, geminiAPIKey),
		superheroClient: superheroClient,
		geminiAPIKey:    geminiAPIKey,
	}
}

// Generate produces a full deck of TotalCards cards for the given theme.
// Fallback chain: Curated pack → Jikan API → Superhero API → Gemini → deterministic hash.
func (g *Generator) Generate(theme string) (models.Deck, models.CardSource, error) {
	// Tier 1: Check curated packs
	if pack, ok := g.matchPack(theme); ok {
		return pack, models.CardSourcePack, nil
	}

	// Tier 2: Try Jikan API for anime (includes caching)
	deck, err := g.jikanClient.FetchDeck(theme)
	if err == nil && len(deck) >= models.TotalCards {
		return deck[:models.TotalCards], models.CardSourceJikan, nil
	}

	// Tier 3: Try Superhero API for Marvel/DC/comics
	if g.superheroClient != nil {
		deck, err = g.superheroClient.FetchDeck(theme)
		if err == nil && len(deck) >= models.TotalCards {
			return deck[:models.TotalCards], models.CardSourceSuperhero, nil
		}
	}

	// Tier 4: Gemini full-generation for any universe
	if g.geminiAPIKey != "" {
		deck, err = fetchGeminiFullDeck(theme, g.geminiAPIKey)
		if err == nil && len(deck) >= models.TotalCards {
			return deck[:models.TotalCards], models.CardSourceGemini, nil
		}
		if err != nil {
			slog.Warn("gemini full-gen failed, falling back to deterministic", "theme", theme, "err", err)
		}
	}

	// Tier 5: Deterministic hash-based generation (last resort)
	deck = g.generateDeterministic(theme)
	return deck, models.CardSourceGenerate, nil
}

// --- Tier 2: Curated packs ---

func (g *Generator) matchPack(theme string) (models.Deck, bool) {
	if packs.IsOnePieceTheme(theme) {
		return packs.OnePiece(), true
	}
	if packs.IsPokemonTheme(theme) {
		return packs.Pokemon(), true
	}
	return nil, false
}

// --- Tier 3: Deterministic hash-based generation ---

func (g *Generator) generateDeterministic(theme string) models.Deck {
	deck := make(models.Deck, 0, models.TotalCards)
	for i := range models.TotalCards {
		name := fmt.Sprintf("%s #%d", theme, i+1)
		deck = append(deck, models.Card{
			ID:    generateCardID(name),
			Name:  name,
			Image: fmt.Sprintf("https://picsum.photos/seed/%s/320/420", url.QueryEscape(fmt.Sprintf("%s|%d", theme, i))),
			Stats: generateDeterministicStats(name),
		})
	}
	return deck
}

// generateDeterministicStats produces stable stats from a character name using FNV-1a hashing.
// The same name always produces the same stats across all games and servers.
func generateDeterministicStats(name string) models.Stats {
	return models.Stats{
		Rank:     hashStat(name, "rank"),
		Strength: hashStat(name, "strength"),
		Speed:    hashStat(name, "speed"),
		IQ:       hashStat(name, "iq"),
	}
}

// hashStat produces a deterministic integer in [10, 99] from name + attribute.
func hashStat(name, attr string) int {
	h := fnv.New32a()
	h.Write([]byte(name + "|" + attr))
	return int(h.Sum32()%90) + 10
}

// generateCardID produces a short hex ID from a character name.
func generateCardID(name string) string {
	h := fnv.New32a()
	h.Write([]byte(name))
	return fmt.Sprintf("%08x", h.Sum32())
}

// Compile-time interface check.
var _ models.CardGenerator = (*Generator)(nil)

// ShuffleDeck returns a new deck with cards in random order (Fisher-Yates).
func ShuffleDeck(deck models.Deck) models.Deck {
	shuffled := make(models.Deck, len(deck))
	copy(shuffled, deck)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := len(shuffled) - 1; i > 0; i-- {
		j := r.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}
	return shuffled
}
