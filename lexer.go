package main

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenType int

const (
	OPEN_PAREN    TokenType = iota // (
	CLOSE_PAREN                    // )
	OPEN_BRACE                     // {
	CLOSE_BRACE                    // }
	OPEN_BRACKET                   // [
	CLOSE_BRACKET                  // ]
	COMMA                          // ,
	COLON                          // :
	STRING                         // "string"
	NUMBER                         // 123
	BOOLEAN                        // true false
	NULL                           // null

	EOI // end of input
)

var TOKEN_TYPE_NAMES = []string{
	OPEN_PAREN:    "OPEN_PARENTHESE",
	CLOSE_PAREN:   "CLOSE_PARENTHESE",
	OPEN_BRACE:    "OPEN_BRACE",
	CLOSE_BRACE:   "CLOSE_BRACE",
	OPEN_BRACKET:  "OPEN_BRACKET",
	CLOSE_BRACKET: "CLOSE_BRACKET",
	COMMA:         "COMMA",
	COLON:         "COLON",
	STRING:        "STRING",
	NUMBER:        "NUMBER",
	BOOLEAN:       "BOOLEAN",
	NULL:          "NULL",

	EOI: "EOI",
}

type Token struct {
	tokenType TokenType
	value     string
}

type LexerState int

const (
	GOING           LexerState = iota // not reading a long token
	READING_STRING                    // in the middle of reading a string
	READING_NUMBER                    // in the middle of reading a number
	READING_BOOLEAN                   // in the middle of reading a boolean
	READING_NULL                      // in the middle of reading a null
)

var STATE_NAMES = []string{
	GOING:           "GOING",
	READING_STRING:  "READING_STRING",
	READING_NUMBER:  "READING_NUMBER",
	READING_BOOLEAN: "READING_BOOLEAN",
	READING_NULL:    "READING_NULL",
}

type Lexer struct {
	state      LexerState
	row        int
	col        int
	buf        []rune
	readTokens []Token
}

func (l *Lexer) readRune(c rune) {
	switch l.state {
	case GOING:
		l.cleanRead(c)
	case READING_STRING:
		l.inStringRead(c)
	case READING_NUMBER:
		l.inNumberRead(c)
	case READING_BOOLEAN:
		l.inBooleanRead(c)
	case READING_NULL:
		l.inNullRead(c)
	}
}

func (l *Lexer) cleanRead(c rune) {
	switch {
	case c == '(':
		l.appendToken(OPEN_PAREN, "(")
	case c == ')':
		l.appendToken(CLOSE_PAREN, ")")
	case c == '{':
		l.appendToken(OPEN_BRACE, "{")
	case c == '}':
		l.appendToken(CLOSE_BRACE, "}")
	case c == '[':
		l.appendToken(OPEN_BRACKET, "[")
	case c == ']':
		l.appendToken(CLOSE_BRACKET, "]")
	case c == ',':
		l.appendToken(COMMA, ",")
	case c == ':':
		l.appendToken(COLON, ":")
	case c == '"':
		l.state = READING_STRING
		l.buf = append(l.buf, c)
	case c == 't' || c == 'f':
		l.state = READING_BOOLEAN
		l.buf = append(l.buf, c)
	case c == 'n':
		l.state = READING_NULL
		l.buf = append(l.buf, c)
	case unicode.IsDigit(c) || c == '-':
		l.state = READING_NUMBER
		l.buf = append(l.buf, c)
	case unicode.IsSpace(c):
	default:
		l.panicUnexpectedCharacter(c)
	}
}

func (l *Lexer) inStringRead(c rune) {
	l.buf = append(l.buf, c)
	if completeString(l.buf) {
		l.appendToken(STRING, string(l.buf))
		l.buf = make([]rune, 0)
		l.state = GOING
	}
}

func completeString(buf []rune) bool {
	if buf[len(buf)-1] != '"' {
		return false
	}

	escaped := false
	for i := len(buf) - 2; i >= 0; i-- {
		if buf[i] == '\\' {
			escaped = !escaped
		} else {
			break
		}
	}

	return !escaped
}

func (l *Lexer) inNumberRead(c rune) {
	validNumberBreak := c == ',' || c == '}' || c == ']' || unicode.IsSpace(c)
	isValidSymbol := c == '-' || c == '+' || c == 'e' || c == 'E' || c == '.'

	if unicode.IsDigit(c) {
		l.buf = append(l.buf, c)
	} else if isValidSymbol {
		// TODO: implement logic for weird numbers like 1.2e-3
		l.buf = append(l.buf, c)
	} else if validNumberBreak {
		l.appendToken(NUMBER, string(l.buf))

		l.buf = make([]rune, 0)
		l.state = GOING
		l.cleanRead(c)
	} else {
		l.panicUnexpectedCharacter(c)
	}
}

func (l *Lexer) inBooleanRead(c rune) {
	fullBool := string(l.buf) == "true" || string(l.buf) == "false"

	if fullBool {
		validBoolBreak := c == ',' || c == '}' || c == ']' || unicode.IsSpace(c)

		if validBoolBreak {
			l.appendToken(BOOLEAN, string(l.buf))

			l.buf = make([]rune, 0)
			l.state = GOING
			l.cleanRead(c)
		} else {
			l.panicUnexpectedCharacter(c)
		}
	} else {
		l.buf = append(l.buf, c)
		goodSoFar := strings.Contains("true", string(l.buf)) || strings.Contains("false", string(l.buf))

		if !goodSoFar {
			l.panicUnexpectedCharacter(c)
		}
	}
}

func (l *Lexer) inNullRead(c rune) {
	fullNull := string(l.buf) == "null"

	if fullNull {
		validNullBreak := c == ',' || c == '}' || c == ']' || unicode.IsSpace(c)

		if validNullBreak {
			l.appendToken(NULL, string(l.buf))

			l.buf = make([]rune, 0)
			l.state = GOING
			l.cleanRead(c)
		} else {
			l.panicUnexpectedCharacter(c)
		}
	} else {
		l.buf = append(l.buf, c)
		goodSoFar := strings.Contains("null", string(l.buf))

		if !goodSoFar {
			l.panicUnexpectedCharacter(c)
		}
	}
}

func (l *Lexer) appendToken(tokenType TokenType, value string) {
	l.readTokens = append(l.readTokens, Token{tokenType: tokenType, value: value})
}

func (l *Lexer) printTokens() {
	fmt.Println("Tokens:")

	for _, token := range l.readTokens {
		fmt.Println(fmt.Sprintf("%-14s: %s", TOKEN_TYPE_NAMES[token.tokenType], token.value))
	}
}

func (l *Lexer) panicUnexpectedCharacter(c rune) {
	panic(fmt.Sprintf("Lexer Error: Unexpected character '%c' in row %d col %d in state %s", c, l.row, l.col, l.stateName()))
}

func (l *Lexer) stateName() string {
	return STATE_NAMES[l.state]
}
