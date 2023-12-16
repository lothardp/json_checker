package main

import (
	"bufio"
	"fmt"
	"os"
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

type NonTerminal int

const (
	JSON         NonTerminal = iota // value
	VALUE                           // object array string number true false null
	MORE_VALUES                     // , value more-values | e
	OBJECT                          // { members }
	MEMBERS                         // member more-members | e
	MEMBER                          // string : value
	MORE_MEMBERS                    // , member more-members | e
	ARRAY                           // [ values ]
	VALUES                          // value values | e
)

var NON_TERMINAL_NAMES = []string{
	JSON:         "JSON",
	VALUE:        "VALUE",
	MORE_VALUES:  "MORE_VALUES",
	OBJECT:       "OBJECT",
	MEMBERS:      "MEMBERS",
	MEMBER:       "MEMBER",
	MORE_MEMBERS: "MORE_MEMBERS",
	ARRAY:        "ARRAY",
	VALUES:       "VALUES",
}

type ItemType int

const (
	TOKEN ItemType = iota
	NON_TERMINAL
)

type StackItem struct {
	itemType ItemType
	value    interface{}
}

type Parser struct {
	tokens []Token
	pos    int
	stack  []StackItem
}

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

func (p *Parser) parse() {
	p.stack = append(p.stack, StackItem{NON_TERMINAL, JSON})

	for p.pos < len(p.tokens) {
		// p.printState()

		token := p.tokens[p.pos]

		if len(p.stack) == 0 {
			if token.tokenType == EOI {
				return
			} else {
				p.panicUnexpectedToken(token)
			}
		}

		stackItem := p.stack[len(p.stack)-1]

		switch stackItem.itemType {
		case TOKEN:
			if stackItem.value == token.tokenType {
				p.pos++
				p.popStack()
			} else {
				p.panicUnexpectedToken(token)
			}
		case NON_TERMINAL:
			p.matchNonTerminal(token, stackItem.value.(NonTerminal))
		}
	}

	p.panicUnexpectedToken(Token{EOI, "EOI"})
}

func (p *Parser) matchNonTerminal(token Token, nonTerminal NonTerminal) {
	switch nonTerminal {
	case JSON:
		switch token.tokenType {
		case OPEN_BRACE, OPEN_BRACKET, STRING, NUMBER, BOOLEAN, NULL:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, VALUE})
		default:
			p.panicUnexpectedToken(token)
		}

	case VALUE:
		switch token.tokenType {
		case OPEN_BRACE:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, OBJECT})
		case OPEN_BRACKET:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, ARRAY})
		case STRING, NUMBER, BOOLEAN, NULL:
			p.popStack()

			p.pushStack(StackItem{TOKEN, token.tokenType})
		default:
			p.panicUnexpectedToken(token)
		}

	case OBJECT:
		switch token.tokenType {
		case OPEN_BRACE:
			p.popStack()

			p.pushStack(StackItem{TOKEN, CLOSE_BRACE})
			p.pushStack(StackItem{NON_TERMINAL, MEMBERS})
			p.pushStack(StackItem{TOKEN, OPEN_BRACE})
		default:
			p.panicUnexpectedToken(token)
		}

	case MEMBERS:
		switch token.tokenType {
		case STRING:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, MORE_MEMBERS})
			p.pushStack(StackItem{NON_TERMINAL, MEMBER})
		case CLOSE_BRACE:
			p.popStack()
		default:
			p.panicUnexpectedToken(token)
		}

	case MEMBER:
		switch token.tokenType {
		case STRING:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, VALUE})
			p.pushStack(StackItem{TOKEN, COLON})
			p.pushStack(StackItem{TOKEN, STRING})
		default:
			p.panicUnexpectedToken(token)
		}

	case MORE_MEMBERS:
		switch token.tokenType {
		case COMMA:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, MORE_MEMBERS})
			p.pushStack(StackItem{NON_TERMINAL, MEMBER})
			p.pushStack(StackItem{TOKEN, COMMA})
		case CLOSE_BRACE:
			p.popStack()
		default:
			p.panicUnexpectedToken(token)
		}

	case ARRAY:
		switch token.tokenType {
		case OPEN_BRACKET:
			p.popStack()

			p.pushStack(StackItem{TOKEN, CLOSE_BRACKET})
			p.pushStack(StackItem{NON_TERMINAL, VALUES})
			p.pushStack(StackItem{TOKEN, OPEN_BRACKET})
		default:
			p.panicUnexpectedToken(token)
		}

	case VALUES:
		switch token.tokenType {
		case OPEN_BRACE, OPEN_BRACKET, STRING, NUMBER, BOOLEAN, NULL:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, MORE_VALUES})
			p.pushStack(StackItem{NON_TERMINAL, VALUE})
		case CLOSE_BRACKET:
			p.popStack()
		default:
			p.panicUnexpectedToken(token)
		}

	case MORE_VALUES:
		switch token.tokenType {
		case COMMA:
			p.popStack()

			p.pushStack(StackItem{NON_TERMINAL, MORE_VALUES})
			p.pushStack(StackItem{NON_TERMINAL, VALUE})
			p.pushStack(StackItem{TOKEN, COMMA})
		case CLOSE_BRACKET:
			p.popStack()
		default:
			p.panicUnexpectedToken(token)
		}
	}
}

func (p *Parser) pushStack(item StackItem) {
	p.stack = append(p.stack, item)
}

func (p *Parser) popStack() StackItem {
	item := p.stack[len(p.stack)-1]
	p.stack = p.stack[:len(p.stack)-1]
	return item
}

func (p *Parser) panicUnexpectedToken(token Token) {
	panic(fmt.Sprintf("Parser Error: Unexpected token %s", TOKEN_TYPE_NAMES[token.tokenType]))
}

func (p *Parser) printState() {
	fmt.Printf("Pos: %d\n", p.pos)
	p.printStack()
	p.printInput()
	fmt.Println("---------------------")
}

func (p *Parser) printStack() {
	fmt.Printf("Stack: ")
	for i := len(p.stack) - 1; i >= 0; i-- {
		item := p.stack[i]
		switch item.itemType {
		case NON_TERMINAL:
			fmt.Printf("%s", NON_TERMINAL_NAMES[item.value.(NonTerminal)])
		case TOKEN:
			fmt.Printf("%s", TOKEN_TYPE_NAMES[item.value.(TokenType)])
		}
		if i > 0 {
			fmt.Printf(" - ")
		}
	}
	fmt.Println()
}

func (p *Parser) printInput() {
	fmt.Printf("Input: ")
	for i := p.pos; i < len(p.tokens); i++ {
		token := p.tokens[i]
		fmt.Printf("%s ", TOKEN_TYPE_NAMES[token.tokenType])
	}
	fmt.Println()
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
