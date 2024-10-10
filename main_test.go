package main

import (
	"strings"
	"testing"
)

func BenchmarkSolve(b *testing.B) {
	const puzzleStr = `appiulv
lposlul
eopueov
lytrtnc
gacsteo
duouune
pilbnac

9
6
4
11
5
5
9`

	grid, wordLens, err := readPuzzle(strings.NewReader(puzzleStr))
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
