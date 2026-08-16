package game

import "github.com/nikhilsaxena04/meta_clash/backend/internal/models"

// MaxStatBot implements models.BotStrategy by always choosing the attribute
// where the card has the highest stat value. Simple but effective.
type MaxStatBot struct{}

// NewMaxStatBot creates a new max-stat bot strategy.
func NewMaxStatBot() *MaxStatBot {
	return &MaxStatBot{}
}

// ChooseAttribute returns the attribute with the highest value on the given card.
// Ties are broken by attribute order: rank > strength > speed > iq.
func (b *MaxStatBot) ChooseAttribute(card models.Card) models.Attribute {
	bestAttr := models.AttrRank
	bestVal := card.Stats.Get(models.AttrRank)

	for _, attr := range models.AllAttributes() {
		val := card.Stats.Get(attr)
		if val > bestVal {
			bestVal = val
			bestAttr = attr
		}
	}

	return bestAttr
}

// Compile-time interface check.
var _ models.BotStrategy = (*MaxStatBot)(nil)
