# wordle

A fast Wordle solver.

## What is wordle?

Wordle is a text-based word search game that can be found on Steam (Wordle, Wordle 2, 3, 4, 5).

Gameplay consists of finding English words in a nxm grid and using those words to fill up the entire grid, thereby using all the characters in the grid.

Here is an example:

![image](https://github.com/user-attachments/assets/28a886e1-d317-484e-a079-ecca4e5cdb76)

These puzzles can be quite difficult when the grid is large and there are a lot of word possibilities. Try finding some words in the above image!

This program finds solutions given a grid and list of word lengths that must cover the grid.

## The package

The most important function in this package is

`func Solve(grid [][]byte, wordLens []byte, dict []string) ([]Solution, error)`

which will find all the unique word combinations that cover the grid, and the words' paths through the grid.

## The solving tool

`main/main.go` is a program that takes puzzle input and prints out solutions in a clear and colorful way.

The program can be used

```
go build main.go
./main -p path/to/your/puzzle.txt
```

where `path/to/your/puzzle.txt` is a puzzle file containing a puzzle. See the `puzzles` directory for examples, the format is simple.

By default, the program uses a dictionary of 360k+ words located at `dictionaries/words_alpha.txt`.
