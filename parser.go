package main

import "fmt"

// Parser turns tokens into an AST.
type Parser struct {
	tokens     []Token
	pos        int
	groupCount int // tracks how many capturing groups we've seen
}

// NewParser builds a parser from a pattern string.
func NewParser(pattern string) (*Parser, error) {
	lex := NewLexer(pattern)
	toks, err := lex.Tokenize()
	if err != nil {
		return nil, err
	}
	return &Parser{tokens: toks, pos: 0}, nil
}

// Parse builds the AST.
func (p *Parser) Parse() (Node, error) {
	node, err := p.parseAlternation()
	if err != nil {
		return nil, err
	}
	if p.peek().Type != TokEOF {
		return nil, fmt.Errorf("unexpected token at position %d: %s", p.peek().Pos, p.peek())
	}
	return node, nil
}

// peek returns the current token without consuming.
func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF}
	}
	return p.tokens[p.pos]
}

// advance consumes and returns the current token.
func (p *Parser) advance() Token {
	t := p.peek()
	p.pos++
	return t
}

// parseAlternation: concatenation ('|' concatenation)*
func (p *Parser) parseAlternation() (Node, error) {
	first, err := p.parseConcat()
	if err != nil {
		return nil, err
	}

	if p.peek().Type != TokPipe {
		return first, nil
	}

	choices := []Node{first}
	for p.peek().Type == TokPipe {
		p.advance() // consume '|'
		next, err := p.parseConcat()
		if err != nil {
			return nil, err
		}
		choices = append(choices, next)
	}

	return Alternate{Choices: choices}, nil
}

// parseConcat: repetition+
func (p *Parser) parseConcat() (Node, error) {
	var parts []Node

	for {
		t := p.peek()
		if t.Type == TokEOF || t.Type == TokPipe || t.Type == TokRParen {
			break
		}
		part, err := p.parseRepetition()
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	if len(parts) == 0 {
		return Empty{}, nil
	}
	if len(parts) == 1 {
		return parts[0], nil
	}
	return Concat{Parts: parts}, nil
}

// parseRepetition: atom ('*' | '+' | '?' | '{n,m}')*
func (p *Parser) parseRepetition() (Node, error) {
	atom, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for {
		t := p.peek()
		switch t.Type {
		case TokStar:
			p.advance()
			atom = Star{Inner: atom}
		case TokPlus:
			p.advance()
			atom = Plus{Inner: atom}
		case TokQuest:
			p.advance()
			atom = Quest{Inner: atom}
		case TokLBrace:
			p.advance()
			min, max, err := p.parseRepeatCount()
			if err != nil {
				return nil, err
			}
			atom = Repeat{Inner: atom, Min: min, Max: max}
		default:
			return atom, nil
		}
	}
}

// parseRepeatCount reads "{n}" or "{n,m}" or "{n,}". Assumes '{' already consumed.
func (p *Parser) parseRepeatCount() (int, int, error) {
	// Read first number
	t := p.advance()
	if t.Type != TokLiteral || t.Ch < '0' || t.Ch > '9' {
		return 0, 0, fmt.Errorf("expected digit after '{' at position %d", t.Pos)
	}
	min := int(t.Ch - '0')
	for {
		t = p.peek()
		if t.Type == TokLiteral && t.Ch >= '0' && t.Ch <= '9' {
			p.advance()
			min = min*10 + int(t.Ch-'0')
		} else {
			break
		}
	}

	// Either '}' or ',' follows
	t = p.advance()
	if t.Type == TokRBrace {
		// {n} means exactly n
		return min, min, nil
	}
	if t.Type != TokComma {
		return 0, 0, fmt.Errorf("expected ',' or '}' at position %d, got %s", t.Pos, t)
	}

	// After comma: either digit(s) then '}' or immediate '}'
	t = p.peek()
	if t.Type == TokRBrace {
		p.advance()
		return min, -1, nil // {n,} means n or more
	}

	if t.Type != TokLiteral || t.Ch < '0' || t.Ch > '9' {
		return 0, 0, fmt.Errorf("expected digit or '}' after ',' at position %d", t.Pos)
	}
	p.advance()
	max := int(t.Ch - '0')
	for {
		t = p.peek()
		if t.Type == TokLiteral && t.Ch >= '0' && t.Ch <= '9' {
			p.advance()
			max = max*10 + int(t.Ch-'0')
		} else {
			break
		}
	}

	t = p.advance()
	if t.Type != TokRBrace {
		return 0, 0, fmt.Errorf("expected '}' at position %d, got %s", t.Pos, t)
	}
	return min, max, nil
}

// parseAtom parses a single atom (literal, group, class, etc.).
func (p *Parser) parseAtom() (Node, error) {
	t := p.advance()

	switch t.Type {
	case TokLiteral:
		return Literal{Ch: t.Ch}, nil
	case TokEscape:
		return Literal{Ch: t.Ch}, nil
	case TokAny:
		return Any{}, nil
	case TokCaret:
		return AnchorStart{}, nil
	case TokDollar:
		return AnchorEnd{}, nil
	case TokDigit:
		// \d = [0-9]
		return CharClass{Ranges: [][2]byte{{'0', '9'}}}, nil
	case TokWord:
		// \w = [a-zA-Z0-9_]
		return CharClass{Ranges: [][2]byte{
			{'a', 'z'}, {'A', 'Z'}, {'0', '9'}, {'_', '_'},
		}}, nil
	case TokSpace:
		// \s = [ \t\n\r\f\v]
		return CharClass{Ranges: [][2]byte{
			{' ', ' '}, {'\t', '\t'}, {'\n', '\n'},
			{'\r', '\r'}, {'\f', '\f'}, {'\v', '\v'},
		}}, nil
	case TokLParen:
		p.groupCount++
		idx := p.groupCount
		inner, err := p.parseAlternation()
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokRParen {
			return nil, fmt.Errorf("expected ')' at position %d", p.peek().Pos)
		}
		p.advance()
		return Group{Index: idx, Inner: inner}, nil
	case TokLBracket:
		return p.parseCharClass()
	default:
		return nil, fmt.Errorf("unexpected token at position %d: %s", t.Pos, t)
	}
}

// parseCharClass reads "[...]" contents. Assumes '[' already consumed.
func (p *Parser) parseCharClass() (Node, error) {
	cc := CharClass{}

	// Optional negation: [^...]
	if p.peek().Type == TokCaret {
		p.advance()
		cc.Negated = true
	}

	first := true
	for {
		t := p.peek()
		if t.Type == TokEOF {
			return nil, fmt.Errorf("unterminated character class")
		}
		if t.Type == TokRBracket && !first {
			p.advance()
			break
		}
		if t.Type == TokRBracket && first {
			// "[]" is technically a literal ']' in many regex flavors;
			// we'll just include ']' literally.
			p.advance()
			cc.Ranges = append(cc.Ranges, [2]byte{']', ']'})
			first = false
			continue
		}

		// Read lo
		lo, err := p.readClassChar()
		if err != nil {
			return nil, err
		}

		// Check for range "lo-hi"
		if p.peek().Type == TokDash {
			// Look ahead: if next is ']' then it's a literal dash
			save := p.pos
			p.advance() // consume '-'
			if p.peek().Type == TokRBracket {
				// Trailing '-' is literal
				cc.Ranges = append(cc.Ranges, [2]byte{lo, lo})
				cc.Ranges = append(cc.Ranges, [2]byte{'-', '-'})
				p.pos = save
				p.advance() // re-consume '-' ... actually we want to keep parsing
				// Simplify: just push literal '-' and continue
				cc.Ranges = cc.Ranges[:len(cc.Ranges)-1]
				cc.Ranges = append(cc.Ranges, [2]byte{'-', '-'})
				first = false
				continue
			}
			hi, err := p.readClassChar()
			if err != nil {
				return nil, err
			}
			if hi < lo {
				return nil, fmt.Errorf("invalid range %c-%c", lo, hi)
			}
			cc.Ranges = append(cc.Ranges, [2]byte{lo, hi})
		} else {
			cc.Ranges = append(cc.Ranges, [2]byte{lo, lo})
		}
		first = false
	}

	return cc, nil
}

// readClassChar reads one character inside a char class (handles escapes).
func (p *Parser) readClassChar() (byte, error) {
	t := p.advance()
	switch t.Type {
	case TokLiteral:
		return t.Ch, nil
	case TokEscape:
		return t.Ch, nil
	case TokDigit:
		// For simplicity, \d inside class expands to first range only (rare case)
		return '0', nil
	case TokWord:
		return '_', nil
	case TokSpace:
		return ' ', nil
	case TokLBracket, TokRBracket, TokDash, TokCaret, TokDollar, TokPipe:
		// Treat operator chars as literals inside a class
		switch t.Type {
		case TokLBracket:
			return '[', nil
		case TokRBracket:
			return ']', nil
		case TokDash:
			return '-', nil
		case TokCaret:
			return '^', nil
		case TokDollar:
			return '$', nil
		case TokPipe:
			return '|', nil
		}
	}
	return 0, fmt.Errorf("invalid character in class at position %d: %s", t.Pos, t)
}

// compile is a helper for one-shot parsing.
func compilePattern(pattern string) (Node, error) {
	p, err := NewParser(pattern)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}