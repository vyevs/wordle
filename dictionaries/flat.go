package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
)

func main() {
	fName := os.Args[1]
	bs, _ := os.ReadFile(os.Args[1])

	outBs := make([]byte, 0, len(bs))
	sc := bufio.NewScanner(bytes.NewReader(bs))
	for sc.Scan() {
		line := sc.Text()

		line = strings.ToLower(line)
		line = strings.ReplaceAll(line, " ", "")
		line = strings.ReplaceAll(line, "'", "")
		line = strings.ReplaceAll(line, "-", "")

		outBs = append(outBs, []byte(line)...)
		outBs = append(outBs, '\n')
	}

	os.WriteFile("new"+fName, outBs, 0777)

}
