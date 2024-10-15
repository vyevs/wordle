package wordle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/vyevs/ansi"
)

const emptyCellChar = '.'

// Solve returns all the ways to place len(wordLens) strings on the provided grid.
// wordLens contains the lengths of the strings to be place.
// dict is the dictionary of strings to be considered for placement.
// grid must contains only lowercase alphabet characters ('a' through 'z') or '.' to indicate an empty grid square.
func Solve(grid [][]byte, wordLens []byte, dictionary []string) ([]Solution, error) {
	if err := validateInput(grid, wordLens); err != nil {
		return nil, err
	}

	// We want our word lens in decreasing order, so we look for longest words first.
	slices.Sort(wordLens)
	slices.Reverse(wordLens)

	s := solver{
		grid:     grid,
		wordLens: wordLens,

		solutions: make([]Solution, 0, 1024),
	}

	for _, row := range s.grid {
		for _, c := range row {
			if c != emptyCellChar {
				s.availableChars[c-'a']++
			}
		}
	}

	s.wordLenCandidates = s.makeInitialCandidates(dictionary)

	return s.solve()
}

func validateInput(grid [][]byte, wordLens []byte) error {
	for r, row := range grid {
		for c, char := range row {
			if (char >= 'a' && char <= 'z') || char == emptyCellChar {
				continue
			}

			return fmt.Errorf("grid contains invalid character %c at [%d, %d]", char, r, c)
		}
	}

	gridSz := countAlphaChars(grid)
	var wordLensSum int
	for _, l := range wordLens {
		wordLensSum += int(l)
	}
	if wordLensSum < gridSz {
		return fmt.Errorf("the word lengths provided (sum %d) will not use all %d grid characters", wordLensSum, gridSz)
	}
	if wordLensSum > gridSz {
		return fmt.Errorf("the word lengths provided (sum %d) will not all fit in the grid (%d characters)", wordLensSum, gridSz)
	}

	return nil
}

type solver struct {
	grid           [][]byte // Character grid from which to make words.
	availableChars [26]byte // Count of each available alphabetic character 'a' thru 'z'.

	// The lengths of the words we're looking for. Never changes.
	wordLens []byte

	// wordLenCandidates maps from the word length that we are looking for to it's current list of candidates.
	// The lists get pruned as we descend into the search so that we aren't looking at duplicate same-letter words.
	// len(wordLenCandidates) == max(wordLens) + 1 so there is an entry for each word length.
	// i.e. wordLenCandidates[5] are the candidate words of length 5.
	wordLenCandidates [][]word

	// curSol is the solution we are in the progress of building.
	curSol    Solution
	solutions []Solution
}

type word struct {
	str           string
	possiblePaths []Path
	// The count of each char in str. charCts[i][0] is a letter (between 'a' and 'z') and charCts[i][1] is the count of that letter.
	charCts [][2]byte
}

type Solution struct {
	Words []string
	Paths []Path
}

func (s Solution) clone() Solution {
	return Solution{
		Words: slices.Clone(s.Words),
		Paths: slices.Clone(s.Paths),
	}
}

func (s Solution) String(grid [][]byte) string {
	var b strings.Builder
	b.Grow(128)

	colors := [9]string{"red", "green", "yellow", "cyan", "orange", "pink", "purple", "chartreuse", "light gray"}

	{
		for i, word := range s.Words {
			colorForWord := colors[i]

			b.WriteString(ansi.FGColorName(colorForWord))
			b.WriteString(word)
			b.WriteByte(' ')
		}

		b.WriteByte('\n')
	}

	cellToColor := make(map[[2]byte]string, len(s.Words))
	for i, path := range s.Paths {
		pathColor := colors[i]
		for _, cell := range path {
			cellToColor[cell] = pathColor
		}
	}

	for r, row := range grid {
		for c, char := range row {
			cell := [2]byte{byte(r), byte(c)}
			color := cellToColor[cell]

			b.WriteString(ansi.FGColorName(color))

			b.WriteByte(char)
		}
		b.WriteByte('\n')
	}

	b.WriteString(ansi.Clear)

	return b.String()
}

func (s *solver) solve() ([]Solution, error) {
	s.findSolutions()

	return s.solutions, nil
}

func (s *solver) findSolutions() {
	if len(s.curSol.Words) >= len(s.wordLens) {
		s.solutions = append(s.solutions, s.curSol.clone())
		return
	}

	wordIdx := len(s.curSol.Words)
	wordLen := s.wordLens[wordIdx]
	cands := s.wordLenCandidates[wordLen]

	for i, candidate := range cands {
		if s.haveEnoughCharsForStr(candidate.charCts) {
			s.wordLenCandidates[wordLen] = cands[i+1:]
			s.placeWord(candidate)
			s.wordLenCandidates[wordLen] = cands
		}
	}
}

func (s *solver) placeWord(w word) {
	paths := w.possiblePaths

OUTER:
	for _, path := range paths {
		for _, cell := range path {
			r, c := cell[0], cell[1]

			if s.grid[r][c] == 0 {
				continue OUTER
			}
		}

		// The word can be placed on the current path.
		// Mark the cells of that path as used.

		for i, cell := range path {
			r, c := cell[0], cell[1]
			s.grid[r][c] = 0
			s.availableChars[w.str[i]-'a']--
		}

		s.curSol.Words = append(s.curSol.Words, w.str)
		s.curSol.Paths = append(s.curSol.Paths, path)

		s.findSolutions()

		// Mark the cells of the path as unused.
		for i, cell := range path {
			r, c := cell[0], cell[1]
			s.grid[r][c] = w.str[i]
			s.availableChars[w.str[i]-'a']++
		}
		s.curSol.Words = s.curSol.Words[:len(s.curSol.Words)-1]
		s.curSol.Paths = s.curSol.Paths[:len(s.curSol.Paths)-1]
	}
}

func (s *solver) makeInitialCandidates(dict []string) [][]word {
	initialCandidates := getStrsOfLens(dict, s.wordLens)
	initialCandidates = s.pruneStrsByCharCounts(initialCandidates)

	wordCandidates := makeWordsFromStrs(s.grid, initialCandidates)

	wordLenCandidates := make([][]word, s.wordLens[0]+1)
	for _, w := range wordCandidates {
		wLen := len(w.str)

		cands := wordLenCandidates[wLen]
		if len(cands) == 0 {
			cands = make([]word, 0, 1024)
		}

		cands = append(cands, w)
		wordLenCandidates[wLen] = cands
	}

	return wordLenCandidates
}

func makeWordsFromStrs(grid [][]byte, strs []string) []word {
	words := make([]word, 0, len(strs))

	for _, s := range strs {
		paths := getPossiblePaths(grid, s)
		if len(paths) == 0 {
			// This string cannot be placed contiguously on the grid.
			continue
		}

		// This string can be placed contiguously on the grid. Possibly on multiple paths.
		w := word{
			str:           s,
			possiblePaths: paths,
			charCts:       countChars(s),
		}
		words = append(words, w)
	}

	return words
}

// pruneCandidatesByCharCounts returns a new slice of candidate strings after
// filtering out strings that can't be placed on the grid due to missing characters.
func (s *solver) pruneStrsByCharCounts(words []string) []string {
	newCands := make([]string, 0, len(words))

	for _, w := range words {
		cts := countChars(w)
		if s.haveEnoughCharsForStr(cts) {
			newCands = append(newCands, w)
		}
	}

	return newCands
}

func countChars(s string) [][2]byte {
	var cts [26]byte
	for _, c := range s {
		cts[c-'a']++
	}

	shortCts := make([][2]byte, 0, 16)
	for i, ct := range cts {
		if ct == 0 {
			continue
		}

		sc := [2]byte{'a' + byte(i), ct}
		shortCts = append(shortCts, sc)
	}

	return shortCts
}

func (s *solver) haveEnoughCharsForStr(cts [][2]byte) bool {
	for _, v := range cts {
		char := v[0] - 'a'
		if v[1] > s.availableChars[char] {
			return false
		}
	}

	return true
}

// getStrsOfLens returns all words in dict that are of a length in lens.
func getStrsOfLens(dict []string, lens []byte) []string {
	lens = slices.Compact(slices.Clone(lens))

	out := make([]string, 0, len(dict))
	for _, w := range dict {
		for _, l := range lens {
			if l == byte(len(w)) {
				out = append(out, w)
			}
		}
	}
	return out
}

func countAlphaChars(s [][]byte) int {
	var ct int
	for _, r := range s {
		for _, c := range r {
			if c >= 'a' && c <= 'z' {
				ct++
			}
		}
	}
	return ct
}
