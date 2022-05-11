/*
	Stang is a simple interpreter implemented in Go
*/
package stang

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

const prompt = ">> "

func RunProgram(filename string) {
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

func StartCommandLine(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	scope := evaluator.NewScope(nil)
	for {
		fmt.Printf(prompt)
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
