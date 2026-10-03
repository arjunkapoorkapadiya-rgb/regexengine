package main

import "fmt"

// Lexer turns a pattern string into a slice of tokens.
type Lexer struct {
	pattern string
	pos     int
}

// NewLexer creates a lexer for the given pattern.
func NewLexer(pattern string) *Lexer {
	return &Lexer{pattern: pattern, pos: 0}
}

// Tokenize returns all tokens (with a trailing EOF).
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok, err := l.next()
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TokEOF {
			break
		}
	}
	return tokens, nil
}

// next reads one token from the pattern.
func (l *Lexer) next() (Token, error) {
	if l.pos >= len(l.pattern) {
		return Token{Type: TokEOF, Pos: l.pos}, nil
	}

	start := l.pos
	ch := l.pattern[l.pos]
	l.pos++

	switch ch {
	case '.':
		return Token{Type: TokAny, Pos: start}, nil
	case '*':
		return Token{Type: TokStar, Pos: start}, nil
	case '+':
		return Token{Type: TokPlus, Pos: start}, nil
	case '?':
		return Token{Type: TokQuest, Pos: start}, nil
	case '|':
		return Token{Type: TokPipe, Pos: start}, nil
	case '(':
		return Token{Type: TokLParen, Pos: start}, nil
	case ')':
		return Token{Type: TokRParen, Pos: start}, nil
	case '[':
		return Token{Type: TokLBracket, Pos: start}, nil
	case ']':
		return Token{Type: TokRBracket, Pos: start}, nil
	case '-':
		return Token{Type: TokDash, Pos: start}, nil
	case '^':
		return Token{Type: TokCaret, Pos: start}, nil
	case '$':
		return Token{Type: TokDollar, Pos: start}, nil
	case '{':
		return Token{Type: TokLBrace, Pos: start}, nil
	case '}':
		return Token{Type: TokRBrace, Pos: start}, nil
	case ',':
		return Token{Type: TokComma, Pos: start}, nil
	case '\\':
		// Escape sequence
		if l.pos >= len(l.pattern) {
			return Token{}, fmt.Errorf("trailing backslash at position %d", start)
		}
		esc := l.pattern[l.pos]
		l.pos++
		switch esc {
		case 'd':
			return Token{Type: TokDigit, Pos: start}, nil
		case 'w':
			return Token{Type: TokWord, Pos: start}, nil
		case 's':
			return Token{Type: TokSpace, Pos: start}, nil
		case 'D', 'W', 'S':
			// Negated shorthands — treat as escapes for now, handle in parser
			return Token{Type: TokEscape, Ch: esc, Pos: start}, nil
		default:
			return Token{Type: TokEscape, Ch: esc, Pos: start}, nil
		}
	default:
		return Token{Type: TokLiteral, Ch: ch, Pos: start}, nil
	}
}