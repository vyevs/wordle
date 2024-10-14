package wordle

import (
	"slices"
	"strings"

	"github.com/vyevs/ansi"
)

type Path [][2]byte

func (p Path) clone() Path {
	return slices.Clone(p)
}

func (p Path) trimLast() Path {
	return p[:len(p)-1]
}

func (p Path) String(grid [][]byte) string {
	var b strings.Builder
	b.Grow(128)
	color := ansi.FGRed

	pathCells := make(map[[2]byte]struct{}, len(p))
	for _, cell := range p {
		pathCells[cell] = struct{}{}
	}

	for r, row := range grid {
		for c, char := range row {
			cell := [2]byte{byte(r), byte(c)}

			if _, partOfPath := pathCells[cell]; partOfPath {
				b.WriteString(color)
				b.WriteByte(char)
				b.WriteString(ansi.Clear)
			} else {
				b.WriteByte(char)
			}
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func getPossiblePaths(grid [][]byte, word string) []Path {
	pf := pathFinder{
		grid:    grid,
		curPath: make(Path, 0, len(word)),
	}

	for r, row := range grid {
		for c := range row {
			pf.walkPossiblePath(word, byte(r), byte(c))
		}
	}

	return pf.allPaths
}

type pathFinder struct {
	grid [][]byte

	curPath  Path
	allPaths []Path
}

func (pf *pathFinder) walkPossiblePath(word string, r, c byte) {
	// If row is out of bounds, we can't place a char in this direction.
	if r >= byte(len(pf.grid)) {
		return
	}
	// If col is out of bounds, we can't place a char in this direction.
	if c >= byte(len(pf.grid[r])) {
		return
	}

	char := word[0]
	if char != pf.grid[r][c] {
		return
	}

	pf.curPath = append(pf.curPath, [2]byte{r, c})
	defer func() {
		pf.curPath = pf.curPath.trimLast()
	}()

	if len(word) == 1 {
		pf.allPaths = append(pf.allPaths, pf.curPath.clone())
		return
	}

	// Mark this grid cell as being unusable for the rest of this path walk.
	pf.grid[r][c] = 0
	defer func() {
		pf.grid[r][c] = char
	}()

	restOfWord := word[1:]

	pf.walkPossiblePath(restOfWord, r-1, c)
	pf.walkPossiblePath(restOfWord, r+1, c)
	pf.walkPossiblePath(restOfWord, r, c-1)
	pf.walkPossiblePath(restOfWord, r, c+1)
	pf.walkPossiblePath(restOfWord, r-1, c-1)
	pf.walkPossiblePath(restOfWord, r-1, c+1)
	pf.walkPossiblePath(restOfWord, r+1, c-1)
	pf.walkPossiblePath(restOfWord, r+1, c+1)
}
