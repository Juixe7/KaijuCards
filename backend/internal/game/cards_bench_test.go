package game

import (
	"testing"
)

// BenchmarkHashStat measures a single FNV-1a stat hash.
func BenchmarkHashStat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = hashStat("Monkey D. Luffy", "strength")
	}
}

// BenchmarkGenerateDeterministicStats measures generating all 4 stats for one character.
func BenchmarkGenerateDeterministicStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateDeterministicStats("Monkey D. Luffy")
	}
}

// BenchmarkGenerateCardID measures FNV-based card ID generation.
func BenchmarkGenerateCardID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = generateCardID("Monkey D. Luffy")
	}
}

// BenchmarkGenerateDeck measures full 24-card deterministic deck generation.
func BenchmarkGenerateDeck(b *testing.B) {
	gen := &Generator{jikanClient: NewJikanClient("https://api.jikan.moe/v4", 0, "")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = gen.generateDeterministic("one piece")
	}
}

// BenchmarkShuffleDeck measures Fisher-Yates shuffle of a 24-card deck.
func BenchmarkShuffleDeck(b *testing.B) {
	gen := &Generator{jikanClient: NewJikanClient("https://api.jikan.moe/v4", 0, "")}
	deck := gen.generateDeterministic("benchmark")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ShuffleDeck(deck)
	}
}
