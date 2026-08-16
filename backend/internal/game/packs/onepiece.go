// Package packs provides curated, lore-accurate card data for fallback when
// the Jikan API is unavailable or unreliable.
package packs

import "github.com/nikhilsaxena04/meta_clash/backend/internal/models"

// OnePiece returns 24 hand-curated One Piece characters with lore-accurate stats.
// Stats range 1-99: Rank (overall power level), Strength, Speed, IQ.
func OnePiece() models.Deck {
	return models.Deck{
		{ID: "op01", Name: "Monkey D. Luffy", Image: "/images/onepiece/op01.png", Stats: models.Stats{Rank: 95, Strength: 92, Speed: 88, IQ: 45}},
		{ID: "op02", Name: "Roronoa Zoro", Image: "/images/onepiece/op02.png", Stats: models.Stats{Rank: 90, Strength: 95, Speed: 80, IQ: 50}},
		{ID: "op03", Name: "Sanji", Image: "/images/onepiece/op03.png", Stats: models.Stats{Rank: 85, Strength: 82, Speed: 92, IQ: 70}},
		{ID: "op04", Name: "Nami", Image: "/images/onepiece/op04.png", Stats: models.Stats{Rank: 60, Strength: 30, Speed: 65, IQ: 90}},
		{ID: "op05", Name: "Nico Robin", Image: "/images/onepiece/op05.png", Stats: models.Stats{Rank: 75, Strength: 70, Speed: 60, IQ: 95}},
		{ID: "op06", Name: "Tony Tony Chopper", Image: "/images/onepiece/op06.png", Stats: models.Stats{Rank: 55, Strength: 60, Speed: 50, IQ: 85}},
		{ID: "op07", Name: "Franky", Image: "/images/onepiece/op07.png", Stats: models.Stats{Rank: 70, Strength: 85, Speed: 45, IQ: 80}},
		{ID: "op08", Name: "Brook", Image: "/images/onepiece/op08.png", Stats: models.Stats{Rank: 65, Strength: 55, Speed: 85, IQ: 60}},
		{ID: "op09", Name: "Usopp", Image: "/images/onepiece/op09.png", Stats: models.Stats{Rank: 50, Strength: 35, Speed: 55, IQ: 75}},
		{ID: "op10", Name: "Jinbe", Image: "/images/onepiece/op10.png", Stats: models.Stats{Rank: 80, Strength: 88, Speed: 55, IQ: 72}},
		{ID: "op11", Name: "Shanks", Image: "/images/onepiece/op11.png", Stats: models.Stats{Rank: 98, Strength: 90, Speed: 90, IQ: 85}},
		{ID: "op12", Name: "Kaido", Image: "/images/onepiece/op12.png", Stats: models.Stats{Rank: 97, Strength: 99, Speed: 70, IQ: 55}},
		{ID: "op13", Name: "Big Mom", Image: "/images/onepiece/op13.png", Stats: models.Stats{Rank: 96, Strength: 95, Speed: 60, IQ: 65}},
		{ID: "op14", Name: "Blackbeard", Image: "/images/onepiece/op14.png", Stats: models.Stats{Rank: 94, Strength: 88, Speed: 50, IQ: 80}},
		{ID: "op15", Name: "Trafalgar Law", Image: "/images/onepiece/op15.png", Stats: models.Stats{Rank: 82, Strength: 70, Speed: 75, IQ: 92}},
		{ID: "op16", Name: "Boa Hancock", Image: "/images/onepiece/op16.png", Stats: models.Stats{Rank: 83, Strength: 78, Speed: 80, IQ: 68}},
		{ID: "op17", Name: "Doflamingo", Image: "/images/onepiece/op17.png", Stats: models.Stats{Rank: 88, Strength: 82, Speed: 78, IQ: 90}},
		{ID: "op18", Name: "Crocodile", Image: "/images/onepiece/op18.png", Stats: models.Stats{Rank: 78, Strength: 75, Speed: 62, IQ: 88}},
		{ID: "op19", Name: "Portgas D. Ace", Image: "/images/onepiece/op19.png", Stats: models.Stats{Rank: 86, Strength: 85, Speed: 82, IQ: 55}},
		{ID: "op20", Name: "Sabo", Image: "/images/onepiece/op20.png", Stats: models.Stats{Rank: 87, Strength: 86, Speed: 84, IQ: 72}},
		{ID: "op21", Name: "Mihawk", Image: "/images/onepiece/op21.png", Stats: models.Stats{Rank: 92, Strength: 90, Speed: 88, IQ: 78}},
		{ID: "op22", Name: "Aokiji", Image: "/images/onepiece/op22.png", Stats: models.Stats{Rank: 91, Strength: 80, Speed: 75, IQ: 88}},
		{ID: "op23", Name: "Kizaru", Image: "/images/onepiece/op23.png", Stats: models.Stats{Rank: 89, Strength: 78, Speed: 99, IQ: 60}},
		{ID: "op24", Name: "Akainu", Image: "/images/onepiece/op24.png", Stats: models.Stats{Rank: 93, Strength: 92, Speed: 72, IQ: 75}},
	}
}

// IsOnePieceTheme reports whether the given theme string matches "One Piece".
func IsOnePieceTheme(theme string) bool {
	switch theme {
	case "one piece", "onepiece", "One Piece", "OnePiece", "ONE PIECE":
		return true
	}
	return false
}
