package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"time"
)

func main() {
	if true {
		f, err := os.Create("cpu.prof")
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer f.Close() // error handling omitted for example
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	defer timeIt(time.Now(), "Everything")

	err := myMain()
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}
}

func myMain() error {
	dict, err := getDictionary("dictionaries/words_alpha.txt")
	if err != nil {
		return fmt.Errorf("failed to get dictionary: %v", err)
	}

	fmt.Printf("The dictionary contains %d words\n", len(dict))

	grid, wordLens, err := readPuzzle(os.Stdin)
	if err != nil {
		return fmt.Errorf("failed to read puzzle: %v", err)
	}

	fmt.Println("grid:")
	fmt.Print(gridStr(grid))
	fmt.Printf("looking for words of lengths %v\n", wordLens)

	solutions, err := solve(grid, wordLens, dict)
	if err != nil {
		return fmt.Errorf("failed to solve: %v", err)
	}

	fmt.Printf("found %d solutions\n", len(solutions))
	if false {
		for i, s := range solutions {
			fmt.Printf("%3d: %v\n", i+1, s)
		}
	}

	checkForDuplicates(solutions)

	return nil
}

const N = 7

func checkForDuplicates(solutions [][]string) {
	uniqSols := make(map[[N]string]struct{})
	for _, s := range solutions {
		uniqSols[[N]string(s)] = struct{}{}
	}

	fmt.Printf("%d unique solutions\n", len(uniqSols))
	if false {
		var i int
		for sol := range uniqSols {
			fmt.Printf("%3d: %v\n", i+1, sol)
			i++
		}
	}
}

func solve(grid [][]byte, wordLens []int, dictionary []string) ([][]string, error) {
	if err := validateInput(grid, wordLens); err != nil {
		return nil, err
	}

	// We want our word lens in decreasing order, so we look for longest words first.
	slices.Sort(wordLens)
	slices.Reverse(wordLens)

	s := solver{
		dict: dictionary,

		grid:     grid,
		used:     makeBoolGrid(grid),
		wordLens: wordLens,

		curSol:    make([]string, 0, len(wordLens)),
		solutions: make([][]string, 0, 1024),
	}

	for _, row := range s.grid {
		for _, c := range row {
			s.availableChars[c-'a']++
		}
	}

	for r, row := range s.grid {
		for c, char := range row {
			loc := [2]int{r, c}

			s.charLocations[char-'a'] = append(s.charLocations[char-'a'], loc)
		}
	}

	s.makeInitialCandidates()

	return s.solve()
}

func validateInput(grid [][]byte, wordLens []int) error {
	gridSz := numItems(grid)
	var wordLensSum int
	for _, l := range wordLens {
		wordLensSum += l
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
	dict []string // All the words to consider.

	grid           [][]byte // Character grid from which to make words.
	used           [][]bool // Whether we've used a specific grid position.
	availableChars [26]int  // Count of each available alphabetic character 'a' thru 'z'.

	// The locations of each char 'a' through 'z' in the grid. charLocations[char][i] is [2]int{rowIndex, colIndex}.
	// Never changes.
	charLocations [26][][2]int
	// The lengths of the words we're looking for. Never changes.
	wordLens []int
	// The initial candidate words for each word length. Never changes.
	initialCandidates [][]string

	// curSol is the solution we are in the progress of building.
	curSol    []string
	solutions [][]string
}

func (s *solver) solve() ([][]string, error) {
	defer timeIt(time.Now(), "Solving")

	s.findSolutions()

	return s.solutions, nil
}

func (s *solver) findSolutions() {
	if len(s.curSol) >= len(s.wordLens) {
		s.solutions = append(s.solutions, slices.Clone(s.curSol))
		return
	}

	nextWordIdx := len(s.curSol)
	cands := s.initialCandidates[nextWordIdx]

	for _, candidate := range cands {
		if s.haveEnoughCharsForWord(candidate) {
			s.placeWord(candidate)
		}
	}
}

func (s *solver) placeWord(word string) {
	firstChar := word[0]
	firstCharLocs := s.charLocations[firstChar-'a']
	for _, loc := range firstCharLocs {
		s.placeWordRec(loc[0], loc[1], word, 0)
	}
}

func (s *solver) placeWordRec(r, c int, candidate string, charIdx int) {
	// If row is out of bounds, we can't solve the puzzle in this direction.
	if r < 0 || r >= len(s.grid) {
		return
	}
	// If col is out of bounds, we can't solve the puzzle going in this direction.
	if c < 0 || c >= len(s.grid[r]) {
		return
	}

	if s.used[r][c] {
		return
	}

	char := candidate[charIdx]
	if candidate[charIdx] != s.grid[r][c] {
		return
	}

	if charIdx == len(candidate)-1 {
		s.curSol = append(s.curSol, candidate)

		s.findSolutions()

		s.curSol = s.curSol[:len(s.curSol)-1]

		return
	}

	s.used[r][c] = true
	s.availableChars[char-'a']--

	nextCharIdx := charIdx + 1
	s.placeWordRec(r-1, c, candidate, nextCharIdx)
	s.placeWordRec(r+1, c, candidate, nextCharIdx)
	s.placeWordRec(r, c-1, candidate, nextCharIdx)
	s.placeWordRec(r, c+1, candidate, nextCharIdx)
	s.placeWordRec(r-1, c-1, candidate, nextCharIdx)
	s.placeWordRec(r-1, c+1, candidate, nextCharIdx)
	s.placeWordRec(r+1, c-1, candidate, nextCharIdx)
	s.placeWordRec(r+1, c+1, candidate, nextCharIdx)

	s.used[r][c] = false
	s.availableChars[char-'a']++
}

func (s *solver) makeInitialCandidates() {
	s.initialCandidates = make([][]string, 0, len(s.wordLens))
	// Initial candidates are all the words with the same len as the words we've looking for.
	for _, l := range s.wordLens {
		s.initialCandidates = append(s.initialCandidates, getWordsOfLen(s.dict, l))
	}

	// Prune candidates to words for which we have the correct character counts.
	{
		for i, cands := range s.initialCandidates {
			s.initialCandidates[i] = s.pruneCandidatesByCharCounts(cands)
		}
	}

	// Prune candidates to words that can be formed contiguously on the grid.
	{
		for i, cands := range s.initialCandidates {
			s.initialCandidates[i] = s.pruneCandidatesByPlacement(cands)
		}
	}

	for i, l := range s.wordLens {
		fmt.Printf("word %d of len %d has %d candidates\n", i, l, len(s.initialCandidates[i]))
	}
}

// pruneCandidatesByCharCounts returns a new slice of candidate strings after
// filtering out strings that can't be placed on the grid due to missing characters.
func (s *solver) pruneCandidatesByCharCounts(cands []string) []string {
	newCands := make([]string, 0, len(cands))

	for _, w := range cands {
		if s.haveEnoughCharsForWord(w) {
			newCands = append(newCands, w)
		}
	}

	return newCands
}

// pruneCandidatesByPlacement returns a new slice of candidate strings after
// filtering out strings that can't be placed contiguously on the grid.
func (s *solver) pruneCandidatesByPlacement(cands []string) []string {
	newCands := make([]string, 0, len(cands))
	for _, cand := range cands {
		if s.canPlaceWordOnGrid(cand) {
			newCands = append(newCands, cand)
		}
	}
	return newCands
}

func (s *solver) haveEnoughCharsForWord(w string) bool {
	var wCts [26]int
	for _, c := range w {
		wCts[c-'a']++
	}

	for i, v := range wCts {
		if v > s.availableChars[i] {
			return false
		}
	}

	return true
}

// canPlaceWordOnGrid returns whether word can be placed on the grid in it's current state.
func (s *solver) canPlaceWordOnGrid(word string) bool {
	firstChar := word[0]
	firstCharLocs := s.charLocations[firstChar-'a']
	for _, loc := range firstCharLocs {
		if s.canPlaceWordRec(loc[0], loc[1], word, 0) {
			return true
		}

	}

	return false
}

func (s *solver) canPlaceWordRec(r, c int, candidate string, charIdx int) bool {
	// If row is out of bounds, we can't place a char in this direction.
	if r < 0 || r >= len(s.grid) {
		return false
	}
	// If col is out of bounds, we can't place a char in this direction.
	if c < 0 || c >= len(s.grid[r]) {
		return false
	}

	if s.used[r][c] {
		return false
	}

	if candidate[charIdx] != s.grid[r][c] {
		return false
	}

	if charIdx == len(candidate)-1 {
		return true
	}

	nextCharIdx := charIdx + 1
	return s.canPlaceWordRec(r-1, c, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r+1, c, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r, c-1, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r, c+1, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r-1, c-1, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r-1, c+1, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r+1, c-1, candidate, nextCharIdx) ||
		s.canPlaceWordRec(r+1, c+1, candidate, nextCharIdx)
}

func getWordsOfLen(dict []string, l int) []string {
	out := make([]string, 0, 1024)
	for _, w := range dict {
		if len(w) == l {
			out = append(out, w)
		}
	}
	return out
}

func getDictionary(file string) ([]string, error) {
	bs, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %v", err)
	}

	sc := bufio.NewScanner(bytes.NewReader(bs))

	dict := make([]string, 0, 1<<20)
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

func readPuzzle(r io.Reader) ([][]byte, []int, error) {
	grid := make([][]byte, 0)

	fmt.Println("enter puzzle lines followed by an empty line:")
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

	wordLens := make([]int, 0)
	fmt.Println("enter target word lengths, one per line:")
	for sc.Scan() {
		line := sc.Text()
		if line == "" {
			break
		}

		wordLen, err := strconv.Atoi(line)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid word length %q: %v", line, err)
		}
		wordLens = append(wordLens, wordLen)
	}
	if err := sc.Err(); err != nil {
		return nil, nil, fmt.Errorf("error reading word lengths: %v", err)
	}

	return grid, wordLens, nil
}

func numItems(s [][]byte) int {
	var ct int
	for _, r := range s {
		ct += len(r)
	}
	return ct
}

func makeBoolGrid(g [][]byte) [][]bool {
	out := make([][]bool, 0, len(g))
	for _, r := range g {
		out = append(out, make([]bool, len(r)))
	}
	return out
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

func timeIt(start time.Time, s string) {
	fmt.Printf("%s took %v\n", s, time.Since(start))
}
