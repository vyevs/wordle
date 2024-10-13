package main

import "slices"

type pathFinder struct {
	grid [][]byte

	curPath  path
	allPaths []path
}

type path [][2]byte

func (p path) clone() path {
	return slices.Clone(p)
}

func (p path) trimLast() path {
	return p[:len(p)-1]
}

func getPossiblePaths(grid [][]byte, word string) []path {
	pf := pathFinder{
		grid:    grid,
		curPath: make(path, 0, len(word)),
	}

	for r, row := range grid {
		for c := range row {
			pf.walkPossiblePath(word, byte(r), byte(c))
		}
	}

	return pf.allPaths
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

	if len(word) == 1 {
		pf.allPaths = append(pf.allPaths, pf.curPath.clone())
		return
	}

	// Mark this grid cell as being unusable for the rest of this path walk.
	pf.grid[r][c] = 0
	pf.curPath = append(pf.curPath, [2]byte{r, c})
	defer func() {
		pf.grid[r][c] = char
		pf.curPath = pf.curPath.trimLast()
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
