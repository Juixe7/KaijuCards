// Package models defines the core domain types for Meta Clash.
// These structs are shared across game engine, lobby management,
// WebSocket messaging, and database persistence layers.
package models

// Attribute represents a single card stat category used in battle comparisons.
type Attribute string

const (
	AttrRank     Attribute = "rank"
	AttrStrength Attribute = "strength"
	AttrSpeed    Attribute = "speed"
	AttrIQ       Attribute = "iq"
)

// AllAttributes returns the ordered list of valid battle attributes.
func AllAttributes() []Attribute {
	return []Attribute{AttrRank, AttrStrength, AttrSpeed, AttrIQ}
}

// Stats holds the four comparable integer values for a card.
// Each stat is in the range [1, 99].
type Stats struct {
	Rank     int `json:"rank"`
	Strength int `json:"strength"`
	Speed    int `json:"speed"`
	IQ       int `json:"iq"`
}

// Get returns the stat value for the given attribute.
func (s Stats) Get(attr Attribute) int {
	switch attr {
	case AttrRank:
		return s.Rank
	case AttrStrength:
		return s.Strength
	case AttrSpeed:
		return s.Speed
	case AttrIQ:
		return s.IQ
	default:
		return 0
	}
}

// Card represents a single playable character card.
type Card struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Image string `json:"image"`
	Stats Stats  `json:"stats"`
}

// CardSource indicates where a card's data originated.
type CardSource string

const (
	CardSourceJikan     CardSource = "jikan"     // External Jikan API (anime)
	CardSourcePack      CardSource = "pack"      // Curated fallback pack
	CardSourceSuperhero CardSource = "superhero" // Superhero API (Marvel/DC/comics)
	CardSourceGemini    CardSource = "gemini"    // Gemini AI full generation (any universe)
	CardSourceGenerate  CardSource = "generate"  // Deterministic hash-based generation (last resort)
)

// Deck is an ordered collection of cards used in a game session.
type Deck []Card

// Constants governing card distribution.
const (
	CardsPerPlayer = 6
	MaxPlayers     = 4
	TotalCards     = CardsPerPlayer * MaxPlayers // 24
)

// CardGenerator defines the interface for producing a themed deck.
// Implementations include Jikan API client, curated packs, and
// deterministic hash-based generation.
type CardGenerator interface {
	// Generate produces a full deck of TotalCards cards for the given theme.
	// Returns the cards, the source they came from, and any error.
	Generate(theme string) (Deck, CardSource, error)
}
