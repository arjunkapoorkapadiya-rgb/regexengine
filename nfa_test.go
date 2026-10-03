package main

import (
	"testing"
)

func TestNFACompileBasic(t *testing.T) {
	cases := []string{
		"a",
		"abc",
		"a|b",
		"a*",
		"a+",
		"a?",
		"(a|b)*c",
		"[a-z]+",
		"\\d{3}",
		"^abc$",
		"a{2,4}",
	}

	for _, pat := range cases {
		ast, err := compilePattern(pat)
		if err != nil {
			t.Errorf("pattern %q: parse error: %v", pat, err)
			continue
		}
		nfa, err := Compile(ast)
		if err != nil {
			t.Errorf("pattern %q: compile error: %v", pat, err)
			continue
		}
		if len(nfa.States) == 0 {
			t.Errorf("pattern %q: NFA has no states", pat)
		}
		if nfa.Start == nfa.Accept {
			t.Errorf("pattern %q: start == accept (should be distinct)", pat)
		}
	}
}

func TestNFARepeatCount(t *testing.T) {
	// "a{3}" should have at least 3 literal transitions for 'a'
	ast, _ := compilePattern("a{3}")
	nfa, _ := Compile(ast)

	count := 0
	for _, st := range nfa.States {
		for _, tr := range st.Transitions {
			if tr.Kind == TransLiteral && tr.Ch == 'a' {
				count++
			}
		}
	}
	if count != 3 {
		t.Errorf("a{3}: expected 3 'a' transitions, got %d", count)
	}
}