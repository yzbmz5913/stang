package main

import (
	"github.com/yzbmz5913/stang"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 1 {
		stang.StartCommandLine(os.Stdin, os.Stdout)
	} else {
		stang.RunProgram(os.Args[0])
	}
}
