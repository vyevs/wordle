package wordle

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func ReadDictionaryFromFile(file string) ([]string, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}
	return ReadDictionary(bytes.NewReader(bs))
}

func ReadDictionary(r io.Reader) ([]string, error) {
	sc := bufio.NewScanner(r)

	dict := make([]string, 0, 1<<19)
	for sc.Scan() {
		line := sc.Text()
		line = strings.TrimSpace(line)
		if line != "" {
			dict = append(dict, line)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %v", err)
	}

	return dict, nil
}

func ReadPuzzleFromFile(file string) ([][]byte, []byte, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to read file: %v", err)
	}
	return ReadPuzzle(bytes.NewReader(bs))
}

func ReadPuzzle(r io.Reader) ([][]byte, []byte, error) {
	grid := make([][]byte, 0)

	sc := bufio.NewScanner(r)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			break
		}

		gridLine := make([]byte, 0, len(line))
		for i := 0; i < len(line); i++ {
			c := line[i]
			gridLine = append(gridLine, c)
		}
		grid = append(grid, gridLine)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading grid: %v", err)
	}

	wordLens := make([]byte, 0)
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			break
		}

		wordLen, err := strconv.Atoi(line)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid word length %q: %v", line, err)
		}
		wordLens = append(wordLens, byte(wordLen))
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading word lengths: %v", err)
	}

	return grid, wordLens, nil
}
