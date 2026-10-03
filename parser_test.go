package main

import (
	"testing"
)

func TestParseBasic(t *testing.T) {
	cases := []struct {
		pattern string
		check   func(Node) bool
	}{
		{"a", func(n Node) bool {
			_, ok := n.(Literal)
			return ok
		}},
		{"ab", func(n Node) bool {
			c, ok := n.(Concat)
			return ok && len(c.Parts) == 2
		}},
		{"a|b", func(n Node) bool {
			a, ok := n.(Alternate)
			return ok && len(a.Choices) == 2
		}},
		{"a*", func(n Node) bool {
			_, ok := n.(Star)
			return ok
		}},
		{"a+", func(n Node) bool {
			_, ok := n.(Plus)
			return ok
		}},
		{"a?", func(n Node) bool {
			_, ok := n.(Quest)
			return ok
		}},
		{"(a)", func(n Node) bool {
			g, ok := n.(Group)
			return ok && g.Index == 1
		}},
		{".", func(n Node) bool {
			_, ok := n.(Any)
			return ok
		}},
		{"^abc$", func(n Node) bool {
			c, ok := n.(Concat)
			if !ok || len(c.Parts) != 5 {
				return false
			}
			_, a := c.Parts[0].(AnchorStart)
			_, e := c.Parts[4].(AnchorEnd)
			return a && e
		}},
	}

	for _, tc := range cases {
		node, err := compilePattern(tc.pattern)
		if err != nil {
			t.Errorf("pattern %q: parse error: %v", tc.pattern, err)
			continue
		}
		if !tc.check(node) {
			t.Errorf("pattern %q: AST check failed. Got: %#v", tc.pattern, node)
		}
	}
}

func TestParsePrecedence(t *testing.T) {
	// "ab|cd" should parse as Alternate{Concat{a,b}, Concat{c,d}}
	node, err := compilePattern("ab|cd")
	if err != nil {
		t.Fatal(err)
	}
	alt, ok := node.(Alternate)
	if !ok {
		t.Fatalf("expected Alternate, got %#v", node)
	}
	if len(alt.Choices) != 2 {
		t.Fatalf("expected 2 choices, got %d", len(alt.Choices))
	}
	for i, ch := range alt.Choices {
		c, ok := ch.(Concat)
		if !ok || len(c.Parts) != 2 {
			t.Errorf("choice %d: expected Concat of 2, got %#v", i, ch)
		}
	}
}

func TestParseErrors(t *testing.T) {
	cases := []string{
		"(a",      // unclosed paren
		"[a-",     // unclosed class
		"a{",      // unclosed brace
		"a{2,",    // unclosed repeat
	}
	for _, pat := range cases {
		_, err := compilePattern(pat)
		if err == nil {
			t.Errorf("pattern %q: expected error, got nil", pat)
		}
	}
}