package packs

import "github.com/nikhilsaxena04/meta_clash/backend/internal/models"

// Pokemon returns 24 hand-curated Pokémon (gen 1-3 favorites) with balanced stats.
// Stats range 1-99: Rank (overall tier), Strength (Atk), Speed, IQ (Sp.Atk/strategy).
func Pokemon() models.Deck {
	return models.Deck{
		{ID: "pk01", Name: "Charizard", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/6.png", Stats: models.Stats{Rank: 85, Strength: 84, Speed: 80, IQ: 60}},
		{ID: "pk02", Name: "Pikachu", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/25.png", Stats: models.Stats{Rank: 70, Strength: 55, Speed: 90, IQ: 65}},
		{ID: "pk03", Name: "Mewtwo", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/150.png", Stats: models.Stats{Rank: 98, Strength: 90, Speed: 85, IQ: 99}},
		{ID: "pk04", Name: "Gengar", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/94.png", Stats: models.Stats{Rank: 82, Strength: 65, Speed: 88, IQ: 90}},
		{ID: "pk05", Name: "Dragonite", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/149.png", Stats: models.Stats{Rank: 88, Strength: 92, Speed: 70, IQ: 72}},
		{ID: "pk06", Name: "Snorlax", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/143.png", Stats: models.Stats{Rank: 75, Strength: 88, Speed: 20, IQ: 45}},
		{ID: "pk07", Name: "Gyarados", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/130.png", Stats: models.Stats{Rank: 84, Strength: 90, Speed: 68, IQ: 50}},
		{ID: "pk08", Name: "Alakazam", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/65.png", Stats: models.Stats{Rank: 80, Strength: 40, Speed: 85, IQ: 98}},
		{ID: "pk09", Name: "Machamp", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/68.png", Stats: models.Stats{Rank: 76, Strength: 95, Speed: 55, IQ: 35}},
		{ID: "pk10", Name: "Blastoise", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/9.png", Stats: models.Stats{Rank: 79, Strength: 78, Speed: 60, IQ: 68}},
		{ID: "pk11", Name: "Venusaur", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/3.png", Stats: models.Stats{Rank: 78, Strength: 76, Speed: 58, IQ: 72}},
		{ID: "pk12", Name: "Arcanine", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/59.png", Stats: models.Stats{Rank: 77, Strength: 85, Speed: 82, IQ: 55}},
		{ID: "pk13", Name: "Lapras", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/131.png", Stats: models.Stats{Rank: 74, Strength: 70, Speed: 50, IQ: 80}},
		{ID: "pk14", Name: "Scizor", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/212.png", Stats: models.Stats{Rank: 83, Strength: 90, Speed: 55, IQ: 60}},
		{ID: "pk15", Name: "Tyranitar", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/248.png", Stats: models.Stats{Rank: 90, Strength: 95, Speed: 48, IQ: 62}},
		{ID: "pk16", Name: "Salamence", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/373.png", Stats: models.Stats{Rank: 89, Strength: 92, Speed: 80, IQ: 58}},
		{ID: "pk17", Name: "Metagross", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/376.png", Stats: models.Stats{Rank: 87, Strength: 92, Speed: 55, IQ: 85}},
		{ID: "pk18", Name: "Gardevoir", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/282.png", Stats: models.Stats{Rank: 81, Strength: 55, Speed: 70, IQ: 92}},
		{ID: "pk19", Name: "Blaziken", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/257.png", Stats: models.Stats{Rank: 86, Strength: 90, Speed: 75, IQ: 55}},
		{ID: "pk20", Name: "Swampert", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/260.png", Stats: models.Stats{Rank: 81, Strength: 85, Speed: 50, IQ: 60}},
		{ID: "pk21", Name: "Flygon", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/330.png", Stats: models.Stats{Rank: 73, Strength: 78, Speed: 80, IQ: 58}},
		{ID: "pk22", Name: "Absol", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/359.png", Stats: models.Stats{Rank: 71, Strength: 88, Speed: 72, IQ: 52}},
		{ID: "pk23", Name: "Milotic", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/350.png", Stats: models.Stats{Rank: 79, Strength: 50, Speed: 68, IQ: 88}},
		{ID: "pk24", Name: "Rayquaza", Image: "https://raw.githubusercontent.com/PokeAPI/sprites/master/sprites/pokemon/other/official-artwork/384.png", Stats: models.Stats{Rank: 99, Strength: 95, Speed: 88, IQ: 75}},
	}
}

// IsPokemonTheme reports whether the given theme string matches "Pokémon".
func IsPokemonTheme(theme string) bool {
	switch theme {
	case "pokemon", "Pokemon", "POKEMON", "pokémon", "Pokémon":
		return true
	}
	return false
}
