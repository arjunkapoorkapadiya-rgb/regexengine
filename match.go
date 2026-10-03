package main

// MatchResult describes a single match.
type MatchResult struct {
	Start  int
	Length int
}

// Match reports whether the pattern matches anywhere in text.
func Match(pattern, text string) (bool, error) {
	_, _, found, err := Find(pattern, text)
	return found, err
}

// Find returns the start index and length of the first match.
func Find(pattern, text string) (int, int, bool, error) {
	ast, err := compilePattern(pattern)
	if err != nil {
		return 0, 0, false, err
	}
	nfa, err := Compile(ast)
	if err != nil {
		return 0, 0, false, err
	}

	for start := 0; start <= len(text); start++ {
		length, ok := runNFAAt(nfa, text, start)
		if ok {
			return start, length, true, nil
		}
	}
	return 0, 0, false, nil
}

// FindAll returns every non-overlapping match.
func FindAll(pattern, text string) ([]MatchResult, error) {
	ast, err := compilePattern(pattern)
	if err != nil {
		return nil, err
	}
	nfa, err := Compile(ast)
	if err != nil {
		return nil, err
	}

	var out []MatchResult
	pos := 0
	for pos <= len(text) {
		found := false
		for start := pos; start <= len(text); start++ {
			length, ok := runNFAAt(nfa, text, start)
			if ok {
				out = append(out, MatchResult{Start: start, Length: length})
				pos = start + length
				if length == 0 {
					pos = start + 1
				}
				found = true
				break
			}
		}
		if !found {
			break
		}
	}
	return out, nil
}

// Replace substitutes all matches with the replacement.
func Replace(pattern, text, replacement string) (string, error) {
	matches, err := FindAll(pattern, text)
	if err != nil {
		return "", err
	}
	if len(matches) == 0 {
		return text, nil
	}

	var out []byte
	last := 0
	for _, m := range matches {
		out = append(out, text[last:m.Start]...)
		out = append(out, replacement...)
		last = m.Start + m.Length
	}
	out = append(out, text[last:]...)
	return string(out), nil
}

// runNFAAt runs the NFA starting at position `start` in text.
// Anchor semantics are absolute: `^` requires start==0 AND text position==0;
// `$` requires text position == len(text).
func runNFAAt(nfa *NFA, text string, start int) (int, bool) {
	// Build initial active set with anchors resolved at text position `start`.
	active := map[StateID]bool{}
	active = addStateWithAnchors(nfa, active, nfa.Start, text, start)

	if len(active) == 0 {
		return 0, false
	}

	lastAccept := -1
	if active[nfa.Accept] {
		// Zero-width match at `start`
		if start == 0 || !requiresStartAnchorToBeAtZero(nfa, active) {
			lastAccept = 0
		} else {
			// has ^ and start != 0 → no match
			return 0, false
		}
	}

	for i := start; i < len(text); i++ {
		ch := text[i]

		next := map[StateID]bool{}
		for s := range active {
			for _, tr := range nfa.States[s].Transitions {
				if transitionMatchesByte(tr, ch) {
					next[tr.To] = true
				}
			}
		}

		if len(next) == 0 {
			break
		}

		// Resolve ε and anchors at new position (i+1)
		active = map[StateID]bool{}
		for s := range next {
			active = addStateWithAnchors(nfa, active, s, text, i+1)
		}

		if active[nfa.Accept] {
			// Greedy: keep consuming if possible
			lastAccept = i + 1 - start
		}
	}

	if lastAccept >= 0 {
		return lastAccept, true
	}
	return 0, false
}

// addStateWithAnchors adds a state plus its ε-closure, resolving
// zero-width anchor transitions based on the current text position.
// `start` for `^` means the ABSOLUTE start of the text (position 0),
// and `$` means position == len(text).
func addStateWithAnchors(nfa *NFA, set map[StateID]bool, start StateID, text string, pos int) map[StateID]bool {
	stack := []StateID{start}
	for len(stack) > 0 {
		s := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if set[s] {
			continue
		}
		set[s] = true

		for _, tr := range nfa.States[s].Transitions {
			switch tr.Kind {
			case TransEpsilon:
				if !set[tr.To] {
					stack = append(stack, tr.To)
				}
			case TransAnchorStart:
				if pos == 0 && !set[tr.To] {
					stack = append(stack, tr.To)
				}
			case TransAnchorEnd:
				if pos == len(text) && !set[tr.To] {
					stack = append(stack, tr.To)
				}
			}
		}
	}
	return set
}

// requiresStartAnchorToBeAtZero is a helper — unused for now.
func requiresStartAnchorToBeAtZero(nfa *NFA, set map[StateID]bool) bool {
	return false
}

// transitionMatchesByte reports whether a transition can consume the byte.
func transitionMatchesByte(tr Transition, ch byte) bool {
	switch tr.Kind {
	case TransLiteral:
		return tr.Ch == ch
	case TransAny:
		return true
	case TransClass:
		return charInClass(tr.Class, ch)
	}
	return false
}

// charInClass checks if a byte is in a CharClass.
func charInClass(cc CharClass, ch byte) bool {
	in := false
	for _, r := range cc.Ranges {
		if ch >= r[0] && ch <= r[1] {
			in = true
			break
		}
	}
	if cc.Negated {
		return !in
	}
	return in
}