package wordle

import (
	"slices"
	"testing"
)

func TestPathFinder(t *testing.T) {
	grid, _, err := ReadPuzzleFromFile("puzzles/wordle3/countries/13.txt")
	if err != nil {
		t.Fatalf("failed to read puzzle: %v", err)
	}

	tests := []struct {
		word      string
		wantPaths []Path
	}{
		{
			word: "mexico",
			wantPaths: []Path{
				{
					{6, 0}, {5, 1}, {6, 1}, {5, 0}, {4, 0}, {3, 0},
				},
			},
		},
		{
			word: "iraq",
			wantPaths: []Path{
				{
					{2, 0}, {1, 0}, {1, 1}, {2, 1},
				},
			},
		},
		{
			word: "usa",
			wantPaths: []Path{
				{
					{0, 0}, {0, 1}, {1, 1},
				},
				{
					{0, 0}, {0, 1}, {1, 2},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.word, func(t *testing.T) {
			paths := getPossiblePaths(grid, tt.word)

			if !pathSlicesEqual(paths, tt.wantPaths) {
				t.Fatalf("list of paths not equal\nwantPaths: %v\ngotPaths: %v", tt.wantPaths, paths)
			}
		})
	}
}

func pathSlicesEqual(ps1, ps2 []Path) bool {
	return slices.EqualFunc(ps1, ps2, func(a, b Path) bool {
		return slices.Equal(a, b)
	})
}
