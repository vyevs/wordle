package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime/pprof"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/vyevs/ansi"
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
	var dictFilePath, puzzleFilePath string
	flag.StringVar(&dictFilePath, "d", "dictionaries/words_alpha.txt", "path to the dictionary file, optional")
	flag.StringVar(&puzzleFilePath, "p", "", "path to the puzzle file")

	var verbose bool
	flag.BoolVar(&verbose, "v", false, "whether to print verbose info")

	flag.Parse()

	if puzzleFilePath == "" {
		return fmt.Errorf("provide a puzzle file using -p switch")
	}

	dict, err := readDictionaryFromFile(dictFilePath)
	if err != nil {
		return fmt.Errorf("failed to get dictionary: %v", err)
	}

	fmt.Printf("The dictionary contains %d words\n", len(dict))

	grid, wordLens, err := readPuzzleFromFile(puzzleFilePath)
	if err != nil {
		return fmt.Errorf("failed to read puzzle: %v", err)
	}

	fmt.Println("grid:")
	fmt.Print(gridStr(grid))
	fmt.Printf("looking for %d words of lengths %v\n", len(wordLens), wordLens)

	solutions, err := solve(grid, wordLens, dict)
	if err != nil {
		return fmt.Errorf("failed to solve: %v", err)
	}

	fmt.Printf("found %d solutions\n", len(solutions))
	if verbose {
		for i, s := range solutions {
			fmt.Printf("%3d\n%v\n", i+1, s.String(grid))
		}
	}

	return nil
}

func solve(grid [][]byte, wordLens []byte, dictionary []string) ([]solution, error) {
	if err := validateInput(grid, wordLens); err != nil {
		return nil, err
	}

	// We want our word lens in decreasing order, so we look for longest words first.
	slices.Sort(wordLens)
	slices.Reverse(wordLens)

	s := solver{
		grid:     grid,
		used:     makeBoolGrid(grid),
		wordLens: wordLens,

		solutions: make([]solution, 0, 1024),
	}

	for _, row := range s.grid {
		for _, c := range row {
			s.availableChars[c-'a']++
		}
	}

	for r, row := range s.grid {
		for c, char := range row {
			loc := [2]byte{byte(r), byte(c)}

			s.charLocations[char-'a'] = append(s.charLocations[char-'a'], loc)
		}
	}

	initialCandidates := s.makeInitialCandidates(dictionary)

	for i, cands := range initialCandidates {
		fmt.Printf("word length %d has %d candidates\n", wordLens[i], len(cands))
	}

	s.wordLenCandidates = make(map[byte][]string, len(wordLens))
	for i, l := range wordLens {
		s.wordLenCandidates[l] = initialCandidates[i]
	}

	return s.solve()
}

func validateInput(grid [][]byte, wordLens []byte) error {
	gridSz := numItems(grid)
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
	used           [][]bool // Whether we've used a specific grid position.
	availableChars [26]byte // Count of each available alphabetic character 'a' thru 'z'.

	// The locations of each char 'a' through 'z' in the grid. charLocations[char][i] is [2]int{rowIndex, colIndex}.
	// Never changes.
	charLocations [26][][2]byte
	// The lengths of the words we're looking for. Never changes.
	wordLens []byte

	// wordLenCandidates maps from the word length that we are looking for to it's current list of candidates.
	// The lists get pruned as we descend into the search so that we aren't looking at duplicate same-letter words.
	wordLenCandidates map[byte][]string

	// curSol is the solution we are in the progress of building.
	curSol    solution
	solutions []solution
}

type solution struct {
	words []string
	cells [][][2]byte
}

func (s solution) clone() solution {
	return solution{
		words: slices.Clone(s.words),
		cells: slices.Clone(s.cells),
	}
}

func (s solution) String(grid [][]byte) string {
	var b strings.Builder
	b.Grow(128)

	colors := [9]string{"red", "light gray", "green", "yellow", "cyan", "orange", "pink", "purple", "chartreuse"}

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
	for i, wordCells := range s.cells {
		cellColor := colors[i]
		for _, cell := range wordCells {
			cellToColor[cell] = cellColor
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
	defer timeIt(time.Now(), "Solving")

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
		if s.haveEnoughCharsForWord(candidate) {
			s.wordLenCandidates[wordLen] = cands[i+1:]
			s.placeWord(candidate)
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

func (s *solver) placeWordRec(r, c byte, candidate string, charIdx int, path [][2]byte) {
	if s.used[r][c] {
		return
	}

	char := candidate[charIdx]
	if char != s.grid[r][c] {
		return
	}

	s.used[r][c] = true
	defer func() {
		s.used[r][c] = false
	}()

	path = append(path, [2]byte{r, c})

	if charIdx == len(candidate)-1 {
		s.curSol.words = append(s.curSol.words, candidate)
		s.curSol.cells = append(s.curSol.cells, slices.Clone(path))

		for _, c := range candidate {
			s.availableChars[c-'a']--
		}

		s.findSolutions()

		for _, c := range candidate {
			s.availableChars[c-'a']++
		}

		s.curSol.words = s.curSol.words[:len(s.curSol.words)-1]
		s.curSol.cells = s.curSol.cells[:len(s.curSol.cells)-1]

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

func (s *solver) makeInitialCandidates(dict []string) [][]string {
	initialCandidates := make([][]string, 0, len(s.wordLens))
	// Initial candidates are all the words with the same len as the words we've looking for.
	for _, l := range s.wordLens {
		initialCandidates = append(initialCandidates, getWordsOfLen(dict, l))
	}

	// Prune candidates to words for which we have the correct character counts.
	{
		for i, cands := range initialCandidates {
			initialCandidates[i] = s.pruneCandidatesByCharCounts(cands)
		}
	}

	// Prune candidates to words that can be formed contiguously on the grid.
	{
		for i, cands := range initialCandidates {
			initialCandidates[i] = s.pruneCandidatesByPlacement(cands)
		}
	}

	return initialCandidates
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
	var wCts [26]byte
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

func (s *solver) canPlaceWordRec(r, c byte, candidate string, charIdx int) bool {
	// If row is out of bounds, we can't place a char in this direction.
	if r >= byte(len(s.grid)) {
		return false
	}
	// If col is out of bounds, we can't place a char in this direction.
	if c >= byte(len(s.grid[r])) {
		return false
	}

	if s.used[r][c] {
		return false
	}

	char := candidate[charIdx]
	if char != s.grid[r][c] {
		return false
	}

	if charIdx == len(candidate)-1 {
		return true
	}

	s.used[r][c] = true
	defer func() {
		s.used[r][c] = false
	}()

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

func getWordsOfLen(dict []string, l byte) []string {
	out := make([]string, 0, 1024*16)
	for _, w := range dict {
		if l == byte(len(w)) {
			out = append(out, w)
		}
	}
	return out
}

func readDictionaryFromFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open dictionary file: %v", err)
	}
	defer f.Close()
	return readDictionary(f)
}

func readDictionary(r io.Reader) ([]string, error) {
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

func readPuzzleFromFile(file string) ([][]byte, []byte, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open puzzle file: %v", err)
	}
	defer f.Close()
	return readPuzzle(f)
}

func readPuzzle(r io.Reader) ([][]byte, []byte, error) {
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
