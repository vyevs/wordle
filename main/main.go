package main

import (
	"flag"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/vyevs/vtools"
	"github.com/vyevs/wordle"
)

func main() {
	defer vtools.TimeIt(time.Now(), "everything")

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

	if puzzleFilePath != "" {
		return solveSinglePuzzle(puzzleFilePath, dictFilePath)
	}

	return solveEmAll(dictFilePath)
}

func solveSinglePuzzle(puzzleFilePath, dictFilePath string) error {
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

	defer vtools.TimeIt(time.Now(), "solving and printing")
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

func solveEmAll(dictFilePath string) error {
	dict, err := wordle.ReadDictionaryFromFile(dictFilePath)
	if err != nil {
		return fmt.Errorf("failed to get dictionary: %v", err)
	}

	walkDir("../puzzles", dict)

	return nil
}

func walkDir(dir string, dict []string) {
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir {
			return nil
		}
		if d.IsDir() {
			return nil
		}

		grid, wordLens, err := wordle.ReadPuzzleFromFile(path)
		if err != nil {
			return fmt.Errorf("failed to read puzzle: %v", err)
		}

		start := time.Now()
		solutions, err := wordle.Solve(grid, wordLens, dict)
		if err != nil {
			return fmt.Errorf("failed to solve: %v", err)
		}

		fmt.Printf("found %5d solutions for %-50q (%s)\n", len(solutions), path, time.Since(start).Round(time.Millisecond))

		return nil
	})

	if err != nil {
		fmt.Printf("walk error: %v", err)
	}
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
