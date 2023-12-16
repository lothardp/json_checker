package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	lexer := Lexer{
		state:      GOING,
		buf:        make([]rune, 0),
		readTokens: make([]Token, 0),
		row:        1,
		col:        1,
	}

	for scanner.Scan() {
		for _, c := range scanner.Text() {
			lexer.readRune(c)
			lexer.col++
		}

		lexer.row++
		lexer.col = 1
	}

	lexer.readTokens = append(lexer.readTokens, Token{EOI, "EOI"})

	// lexer.printTokens()

	parser := Parser{
		tokens: lexer.readTokens,
		pos:    0,
		stack:  make([]StackItem, 0),
	}

	parser.parse()

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}

	fmt.Println("\nValid JSON!")
}

