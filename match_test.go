package main

import (
	"testing"
)

func TestMatchBasic(t *testing.T) {
	cases := []struct {
		pattern string
		text    string
		want    bool
	}{
		{"abc", "xabcy", true},
		{"abc", "xyz", false},
		{"a", "a", true},
		{"a", "b", false},
		{"a*", "aaa", true},
		{"a*", "bbb", true},    // * matches zero — should be true
		{"a+", "aaa", true},
		{"a+", "", false},       // + requires at least one
		{"a?", "", true},        // ? allows zero
		{"a?", "a", true},
		{"a|b", "b", true},
		{"a|b", "c", false},
		{".", "x", true},
		{"[abc]", "b", true},
		{"[abc]", "z", false},
		{"[a-z]+", "hello", true},
		{"[0-9]+", "abc", false},
		{"\\d+", "123", true},
		{"\\d+", "abc", false},
		{"^abc", "abc", true},
		{"^abc", "xabc", false},
		{"abc$", "xabc", true},
		{"abc$", "abcx", false},
		{"(a|b)*c", "aabbc", true},
		{"(a|b)*c", "aabbb", false},
		{"a{2}", "aa", true},
		{"a{2}", "a", false},
		{"a{2,4}", "aaa", true},
		{"a{2,4}", "aaaaa", true}, // can match "aaaa" substring
	}

	for _, tc := range cases {
		got, err := Match(tc.pattern, tc.text)
		if err != nil {
			t.Errorf("pattern %q text %q: error %v", tc.pattern, tc.text, err)
			continue
		}
		if got != tc.want {
			t.Errorf("pattern %q text %q: got %v, want %v", tc.pattern, tc.text, got, tc.want)
		}
	}
}

func TestFind(t *testing.T) {
	cases := []struct {
		pattern string
		text    string
		start   int
		length  int
	}{
		{"abc", "xabcy", 1, 3},
		{"a+", "xxaaayy", 2, 3},
		{"\\d+", "abc123def", 3, 3},
		{"[a-z]+", "123hello456", 3, 5},
	}
	for _, tc := range cases {
		s, l, found, err := Find(tc.pattern, tc.text)
		if err != nil {
			t.Errorf("pattern %q: error %v", tc.pattern, err)
			continue
		}
		if !found {
			t.Errorf("pattern %q text %q: no match, expected at %d len %d",
				tc.pattern, tc.text, tc.start, tc.length)
			continue
		}
		if s != tc.start || l != tc.length {
			t.Errorf("pattern %q text %q: got (%d,%d), want (%d,%d)",
				tc.pattern, tc.text, s, l, tc.start, tc.length)
		}
	}
}

func TestReplace(t *testing.T) {
	cases := []struct {
		pattern     string
		text        string
		replacement string
		want        string
	}{
		{"\\d+", "abc123def456", "X", "abcXdefX"},
		{"a", "banana", "o", "bonono"},
		{"[aeiou]", "hello world", "_", "h_ll_ w_rld"},
	}
	for _, tc := range cases {
		got, err := Replace(tc.pattern, tc.text, tc.replacement)
		if err != nil {
			t.Errorf("pattern %q: error %v", tc.pattern, err)
			continue
		}
		if got != tc.want {
			t.Errorf("pattern %q text %q: got %q, want %q", tc.pattern, tc.text, got, tc.want)
		}
	}
}