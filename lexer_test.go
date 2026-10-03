package main

import (
	"testing"
)

func TestLexerBasic(t *testing.T) {
	cases := []struct {
		pattern string
		want    int // expected token count (including EOF)
	}{
		{"a", 2},
		{"abc", 4},
		{"a*", 3},
		{"a|b", 4},
		{"a(b|c)*", 8},
		{"[a-z]", 6},
		{"\\d+", 3},
		{"^abc$", 6},
	}

	for _, tc := range cases {
		lex := NewLexer(tc.pattern)
		toks, err := lex.Tokenize()
		if err != nil {
			t.Errorf("pattern %q: unexpected error: %v", tc.pattern, err)
			continue
		}
		if len(toks) != tc.want {
			t.Errorf("pattern %q: got %d tokens, want %d. Tokens: %v",
				tc.pattern, len(toks), tc.want, toks)
		}
	}
}

func TestLexerContent(t *testing.T) {
	// Verify specific token types for a complex pattern
	lex := NewLexer("a(b|c)*")
	toks, err := lex.Tokenize()
	if err != nil {
		t.Fatal(err)
	}
	want := []TokenType{
		TokLiteral, TokLParen, TokLiteral, TokPipe,
		TokLiteral, TokRParen, TokStar, TokEOF,
	}
	if len(toks) != len(want) {
		t.Fatalf("got %d tokens, want %d", len(toks), len(want))
	}
	for i, w := range want {
		if toks[i].Type != w {
			t.Errorf("token %d: got %v, want %v", i, toks[i].Type, w)
		}
	}
}