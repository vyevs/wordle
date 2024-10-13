package wordle

import (
	"fmt"
	"slices"
	"strings"

	"github.com/vyevs/ansi"
)

const emptyCellChar = '.'

func Solve(grid [][]byte, wordLens []byte, dictionary []string) ([]solution, error) {
	if err := validateInput(grid, wordLens); err != nil {
		return nil, err
	}

	// We want our word lens in decreasing order, so we look for longest words first.
	slices.Sort(wordLens)
	slices.Reverse(wordLens)

	s := solver{
		grid:     grid,
		wordLens: wordLens,

		solutions: make([]solution, 0, 1024),
	}

	for _, row := range s.grid {
		for _, c := range row {
			if c != emptyCellChar {
				s.availableChars[c-'a']++
			}
		}
	}

	for r, row := range s.grid {
		for c, char := range row {
			if char == emptyCellChar {
				continue
			}

			loc := [2]byte{byte(r), byte(c)}

			s.charLocations[char-'a'] = append(s.charLocations[char-'a'], loc)
		}
	}

	s.makeInitialCandidates(dictionary)

	return s.solve()
}

func validateInput(grid [][]byte, wordLens []byte) error {
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

	// The locations of each char 'a' through 'z' in the grid. charLocations[char][i] is [2]int{rowIndex, colIndex}.
	// Never changes.
	charLocations [26][][2]byte
	// The lengths of the words we're looking for. Never changes.
	wordLens []byte

	// wordLenCandidates maps from the word length that we are looking for to it's current list of candidates.
	// The lists get pruned as we descend into the search so that we aren't looking at duplicate same-letter words.
	wordLenCandidates map[byte][]word

	// curSol is the solution we are in the progress of building.
	curSol    solution
	solutions []solution
}

type word struct {
	str           string
	possiblePaths []path
}

type solution struct {
	words []string
	paths []path
}

func (s solution) clone() solution {
	return solution{
		words: slices.Clone(s.words),
		paths: slices.Clone(s.paths),
	}
}

func (s solution) String(grid [][]byte) string {
	var b strings.Builder
	b.Grow(128)

	colors := [9]string{"red", "green", "yellow", "cyan", "orange", "pink", "purple", "chartreuse", "light gray"}

	{
		for i, word := range s.words {
			colorForWord := colors[i]

			b.WriteString(ansi.FGColorName(colorForWord))
			b.WriteString(word)
			b.WriteByte(' ')
		}

		b.WriteByte('\n')
	}

	cellToColor := make(map[[2]byte]string, len(s.words))
	for i, path := range s.paths {
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

func (s *solver) solve() ([]solution, error) {

	s.findSolutions()

	return s.solutions, nil
}

func (s *solver) findSolutions() {
	if len(s.curSol.words) >= len(s.wordLens) {
		s.solutions = append(s.solutions, s.curSol.clone())
		return
	}

	wordIdx := len(s.curSol.words)
	wordLen := s.wordLens[wordIdx]
	cands := s.wordLenCandidates[wordLen]

	for i, candidate := range cands {
		if s.haveEnoughCharsForStr(candidate.str) {
			s.wordLenCandidates[wordLen] = cands[i+1:]
			s.placeWord(candidate.str)
			s.wordLenCandidates[wordLen] = cands
		}
	}
}

func (s *solver) placeWord(word string) {
	firstChar := word[0]
	firstCharLocs := s.charLocations[firstChar-'a']

	var path [16][2]byte
	for _, loc := range firstCharLocs {
		s.placeWordRec(loc[0], loc[1], word, 0, path[:0])
	}
}

// TODO: Instead of walking the path for the current word, computer each word's possible placements at the beginning!
func (s *solver) placeWordRec(r, c byte, candidate string, charIdx int, path [][2]byte) {
	char := candidate[charIdx]
	if char != s.grid[r][c] {
		return
	}

	// Mark this grid cell as not usable for the rest of this word placement.
	s.grid[r][c] = 0
	defer func() {
		s.grid[r][c] = char
	}()

	path = append(path, [2]byte{r, c})

	if charIdx == len(candidate)-1 {
		s.curSol.words = append(s.curSol.words, candidate)
		s.curSol.paths = append(s.curSol.paths, slices.Clone(path))

		for _, c := range candidate {
			s.availableChars[c-'a']--
		}

		s.findSolutions()

		for _, c := range candidate {
			s.availableChars[c-'a']++
		}

		s.curSol.words = s.curSol.words[:len(s.curSol.words)-1]
		s.curSol.paths = s.curSol.paths[:len(s.curSol.paths)-1]

		return
	}

	nextCharIdx := charIdx + 1

	// If we can move up.
	if r > 0 {
		s.placeWordRec(r-1, c, candidate, nextCharIdx, path)

		// If we can move left and up.
		if c > 0 {
			s.placeWordRec(r-1, c-1, candidate, nextCharIdx, path)
		}
		// If we can move right and up.
		if c < byte(len(s.grid[r])-1) {
			s.placeWordRec(r-1, c+1, candidate, nextCharIdx, path)
		}
	}

	// If we can move down.
	if r < byte(len(s.grid))-1 {
		s.placeWordRec(r+1, c, candidate, nextCharIdx, path)

		// If we can move left and down.
		if c > 0 {
			s.placeWordRec(r+1, c-1, candidate, nextCharIdx, path)
		}

		// If we can move right and down.
		if c < byte(len(s.grid[r])-1) {
			s.placeWordRec(r+1, c+1, candidate, nextCharIdx, path)
		}
	}

	// If we can move left.
	if c > 0 {
		s.placeWordRec(r, c-1, candidate, nextCharIdx, path)
	}

	// If we can move right.
	if c < byte(len(s.grid[r])-1) {
		s.placeWordRec(r, c+1, candidate, nextCharIdx, path)
	}
}

func (s *solver) makeInitialCandidates(dict []string) {
	initialCandidates := getStrsOfLens(dict, s.wordLens)
	initialCandidates = s.pruneStrsByCharCounts(initialCandidates)

	wordCandidates := makeWordsFromStrs(s.grid, initialCandidates)

	fmt.Printf("%d unique words can be placed contiguously on the grid\n", len(wordCandidates))

	s.wordLenCandidates = make(map[byte][]word, len(s.wordLens))
	for _, w := range wordCandidates {
		wLen := byte(len(w.str))

		cands := s.wordLenCandidates[wLen]
		if len(cands) == 0 {
			cands = make([]word, 0, 1024)
		}

		cands = append(cands, w)
		s.wordLenCandidates[wLen] = cands
	}

	for l, cands := range s.wordLenCandidates {
		fmt.Printf("%d length %d word candidates\n", len(cands), l)
	}
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
		if s.haveEnoughCharsForStr(w) {
			newCands = append(newCands, w)
		}
	}

	return newCands
}

func countChars(s string) [26]byte {
	var cts [26]byte
	for _, c := range s {
		cts[c-'a']++
	}
	return cts
}

func (s *solver) haveEnoughCharsForStr(str string) bool {
	cts := countChars(str)

	for i, v := range cts {
		if v > s.availableChars[i] {
			return false
		}
	}

	return true
}

// getStrsOfLens returns all words in dict that are of a length in lens.
func getStrsOfLens(dict []string, lens []byte) []string {
	lens = slices.Compact(lens)

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
