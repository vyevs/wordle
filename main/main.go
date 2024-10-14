package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/vyevs/wordle"
)

func main() {
	defer timeIt(time.Now(), "Everything")

	err := myMain()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func myMain() error {
	var dictFilePath, puzzleFilePath string
	flag.StringVar(&dictFilePath, "d", "../dictionaries/words_alpha.txt", "path to the dictionary file, optional")
	flag.StringVar(&puzzleFilePath, "p", "", "path to the puzzle file")

	flag.Parse()

	if puzzleFilePath == "" {
		return fmt.Errorf("provide a puzzle file using -p switch")
	}

	dict, err := wordle.ReadDictionaryFromFile(dictFilePath)
	if err != nil {
		return fmt.Errorf("failed to get dictionary: %v", err)
	}

	fmt.Printf("The dictionary contains %d words\n", len(dict))

	grid, wordLens, err := wordle.ReadPuzzleFromFile(puzzleFilePath)
	if err != nil {
		return fmt.Errorf("failed to read puzzle: %v", err)
	}

	fmt.Println("grid:")
	fmt.Print(gridStr(grid))
	fmt.Printf("looking for %d words of lengths %v\n", len(wordLens), wordLens)

	defer timeIt(time.Now(), "solving and printing")
	solutions, err := wordle.Solve(grid, wordLens, dict)
	if err != nil {
		return fmt.Errorf("failed to solve: %v", err)
	}

	fmt.Printf("found %d solutions:\n", len(solutions))
	for i, s := range solutions {
		fmt.Printf("%3d\n%v\n", i+1, s.String(grid))
	}

	return nil
}

func timeIt(start time.Time, s string) {
	fmt.Printf("%s took %v\n", s, time.Since(start))
}

func gridStr(g [][]byte) string {
	var b strings.Builder
	b.Grow(len(g) * len(g[0]))
	for _, l := range g {
		for _, c := range l {
			b.WriteByte(c)
		}
		b.WriteByte('\n')
	}

	return b.String()
}
