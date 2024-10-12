package main

import (
	"testing"
)

func BenchmarkSolve(b *testing.B) {
	grid, wordLens, err := readPuzzleFromFile("puzzles/flowers3.txt")
	if err != nil {
		b.Fatalf("failed to read puzzle: %v", err)
	}

	dict, err := readDictionaryFromFile("dictionaries/words_alpha.txt")
	if err != nil {
		b.Fatalf("failed to get dictionary: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err = solve(grid, wordLens, dict)
		if err != nil {
			b.Fatalf("solve failed: %v", err)
		}
	}
}
