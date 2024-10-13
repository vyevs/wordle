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

func (pf *pathFinder) getPossiblePaths(word string) []path {

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
