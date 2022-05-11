package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/yzbmz5913/stang/evaluator"
	"github.com/yzbmz5913/stang/lexer"
	"github.com/yzbmz5913/stang/parser"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const PROMPT = ">> "

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	scope := evaluator.NewScope(nil)
	for {
		fmt.Printf(PROMPT)
		scanned := scanner.Scan()
		if !scanned {
			return
		}
		line := scanner.Text()
		if strings.ToLower(line) == "exit" {
			_, _ = io.WriteString(out, "bye")
			return
		}
		l := lexer.New(line)
		p := parser.New(l)
		program := p.ParseProgram()
		if len(p.Errors()) != 0 {
			printParserErrors(out, p.Errors())
			continue
		}
		result := evaluator.Eval(context.Background(), program, scope)
		if result != nil {
			_, _ = io.WriteString(out, result.String(0))
			_, _ = io.WriteString(out, "\n")
		}
	}
}
func printParserErrors(out io.Writer, errors []string) {
	for _, msg := range errors {
		_, _ = io.WriteString(out, "Error: "+msg+"\n")
	}
}
func runProgram(filename string) {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	f, err := ioutil.ReadFile(wd + "/" + filename)
	if err != nil {
		fmt.Println("monkey: ", err.Error())
		os.Exit(1)
	}
	l := lexer.New(string(f))
	p := parser.New(l)
	program := p.ParseProgram()
	if len(p.Errors()) != 0 {
		fmt.Println(p.Errors()[0])
		os.Exit(1)
	}
	scope := evaluator.NewScope(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	e := evaluator.Eval(ctx, program, scope)
	fmt.Println("program returns:\n", e.String(0))
}

func main() {
	args := os.Args[1:]
	if len(args) == 1 {
		fmt.Println("Welcome to use Stan's programming language(Stang)!")
		fmt.Println("type in command line or pass in filenames as parameters to parse source code")
		fmt.Println()
		Start(os.Stdin, os.Stdout)
	} else {
		//for _, arg := range args {
		//	runProgram(arg)
		//}
		runProgram("test.my")
	}
}
