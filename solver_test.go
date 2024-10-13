package wordle

import (
	"testing"
)

func BenchmarkSolve(b *testing.B) {
	grid, wordLens, err := ReadPuzzleFromFile("puzzles/wordle3/flowers/7.txt")
	if err != nil {
		b.Fatalf("failed to read puzzle: %v", err)
	}

	dict, err := ReadDictionaryFromFile("dictionaries/words_alpha.txt")
	if err != nil {
		b.Fatalf("failed to get dictionary: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = Solve(grid, wordLens, dict)
		if err != nil {
			b.Fatalf("solve failed: %v", err)
		}
	}
}
