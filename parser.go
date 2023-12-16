/* 
LL(1) Parser for JSON
See https://en.wikipedia.org/wiki/LL_parser
*/

package main

import (
	"fmt"
)

/*
JSON Grammar (from http://www.json.org/) adapted to be LL(1)
and with simplified terminals (lexer creates tokens for these)

json
   value

value
   object
   array
   string
   number
   "true"
   "false"
   "null"

object
    '{' members '}'

members
    e
    member more-members

more-members
    e
    ',' member more-members

member
    string ':' value

array
    '[' values ']'

values
    e
    value more-values

more-values
    e
    ',' value more-values

*/

type NonTerminalType int

const (
	JSON         NonTerminalType = iota // value
	VALUE                               // object array string number true false null
	MORE_VALUES                         // , value more-values | e
	OBJECT                              // { members }
	MEMBERS                             // member more-members | e
	MEMBER                              // string : value
	MORE_MEMBERS                        // , member more-members | e
	ARRAY                               // [ values ]
	VALUES                              // value values | e
)

var NON_TERMINAL_TYPE_NAMES = []string{
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

type StackItemType int

const (
	TOKEN StackItemType = iota
	NON_TERMINAL
)

type StackItem struct {
	itemType StackItemType
	value    interface{} // TokenType or NonTerminalType
}

type Parser struct {
	tokens []Token
	pos    int
	stack  []StackItem
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
			p.matchNonTerminal(token, stackItem.value.(NonTerminalType))
		}
	}

	p.panicUnexpectedToken(Token{EOI, "EOI"})
}

func (p *Parser) matchNonTerminal(token Token, nonTerminal NonTerminalType) {
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
			fmt.Printf("%s", NON_TERMINAL_TYPE_NAMES[item.value.(NonTerminalType)])
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
