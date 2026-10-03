package main

import "fmt"

// TokenType identifies the kind of a token.
type TokenType int

const (
	TokEOF TokenType = iota
	TokLiteral       // any literal char
	TokAny           // .
	TokStar          // *
	TokPlus          // +
	TokQuest         // ?
	TokPipe          // |
	TokLParen        // (
	TokRParen        // )
	TokLBracket      // [
	TokRBracket      // ]
	TokDash          // -
	TokCaret         // ^
	TokDollar        // $
	TokLBrace        // {
	TokRBrace        // }
	TokComma         // ,
	TokEscape        // \x — the escaped char is stored in Ch
	TokDigit         // \d
	TokWord          // \w
	TokSpace         // \s
)

// Token is one lexeme from the pattern.
type Token struct {
	Type TokenType
	Ch   byte // for TokLiteral or TokEscape
	Pos  int  // byte offset in the pattern (for error messages)
}

// String is for debugging.
func (t Token) String() string {
	switch t.Type {
	case TokEOF:
		return "EOF"
	case TokLiteral:
		return fmt.Sprintf("LITERAL(%c)", t.Ch)
	case TokEscape:
		return fmt.Sprintf("ESCAPE(%c)", t.Ch)
	case TokAny:
		return "ANY"
	case TokStar:
		return "STAR"
	case TokPlus:
		return "PLUS"
	case TokQuest:
		return "QUEST"
	case TokPipe:
		return "PIPE"
	case TokLParen:
		return "LPAREN"
	case TokRParen:
		return "RPAREN"
	case TokLBracket:
		return "LBRACKET"
	case TokRBracket:
		return "RBRACKET"
	case TokDash:
		return "DASH"
	case TokCaret:
		return "CARET"
	case TokDollar:
		return "DOLLAR"
	case TokLBrace:
		return "LBRACE"
	case TokRBrace:
		return "RBRACE"
	case TokComma:
		return "COMMA"
	case TokDigit:
		return "DIGIT"
	case TokWord:
		return "WORD"
	case TokSpace:
		return "SPACE"
	}
	return "UNKNOWN"
}