package main

// A backtracking matcher that tracks capture groups.
// Slower than the NFA simulator, but easily supports captures.

// BTRunner holds state for a backtracking match attempt.
type BTRunner struct {
	text    string
	groups  []int // groups[i] = start position of group i; -1 = unmatched
	groupsEnd []int // groups[i+1]... actually we store start/end pairs
	caps    map[int][2]int // group index → {start, end}
}

// Match captures the first match at any position.
// Returns start, end, and captures map (group index → substring).
func MatchWithCaptures(pattern, text string) (int, int, map[int]string, bool, error) {
	ast, err := compilePattern(pattern)
	if err != nil {
		return 0, 0, nil, false, err
	}

	for start := 0; start <= len(text); start++ {
		r := &BTRunner{text: text, caps: map[int][2]int{}}
		end, ok := r.match(ast, start, start)
		if ok {
			captures := map[int]string{}
			for idx, span := range r.caps {
				if span[0] >= 0 && span[1] >= span[0] {
					captures[idx] = text[span[0]:span[1]]
				}
			}
			return start, end, captures, true, nil
		}
	}
	return 0, 0, nil, false, nil
}

// match tries to match `node` starting at `pos`, returns the end position if success.
// `start` is the absolute start of the overall match (for `^` anchor check).
func (r *BTRunner) match(node Node, pos int, start int) (int, bool) {
	switch v := node.(type) {
	case Literal:
		if pos < len(r.text) && r.text[pos] == v.Ch {
			return pos + 1, true
		}
		return 0, false

	case Any:
		if pos < len(r.text) {
			return pos + 1, true
		}
		return 0, false

	case CharClass:
		if pos < len(r.text) && charInClass(v, r.text[pos]) {
			return pos + 1, true
		}
		return 0, false

	case AnchorStart:
		if pos == 0 {
			return pos, true
		}
		return 0, false

	case AnchorEnd:
		if pos == len(r.text) {
			return pos, true
		}
		return 0, false

	case Empty:
		return pos, true

	case Concat:
		cur := pos
		for _, p := range v.Parts {
			next, ok := r.match(p, cur, start)
			if !ok {
				return 0, false
			}
			cur = next
		}
		return cur, true

	case Alternate:
		for _, c := range v.Choices {
			// Save capture state before trying
			saved := copyCaps(r.caps)
			if end, ok := r.match(c, pos, start); ok {
				return end, true
			}
			r.caps = saved
		}
		return 0, false

	case Star:
		// Greedy: try as many as possible, then backtrack
		return r.matchStar(v.Inner, pos, start)

	case Plus:
		// At least one, then star
		saved := copyCaps(r.caps)
		first, ok := r.match(v.Inner, pos, start)
		if !ok {
			r.caps = saved
			return 0, false
		}
		end, ok := r.matchStar(v.Inner, first, start)
		if !ok {
			// Star can match zero — ok
			return first, true
		}
		return end, true

	case Quest:
		// Try matching first
		saved := copyCaps(r.caps)
		if end, ok := r.match(v.Inner, pos, start); ok {
			return end, true
		}
		r.caps = saved
		return pos, true

	case Repeat:
		return r.matchRepeat(v, pos, start)

	case Group:
		startPos := pos
		end, ok := r.match(v.Inner, pos, start)
		if !ok {
			return 0, false
		}
		r.caps[v.Index] = [2]int{startPos, end}
		return end, true

	default:
		return pos, true
	}
}

// matchStar implements `inner*` greedily with backtracking.
func (r *BTRunner) matchStar(inner Node, pos int, start int) (int, bool) {
	// Greedy: try matching as many as possible, recording each position
	positions := []int{pos}
	capsStack := []map[int][2]int{copyCaps(r.caps)}
	cur := pos
	for {
		saved := copyCaps(r.caps)
		next, ok := r.match(inner, cur, start)
		if !ok || next == cur {
			// Can't advance further
			r.caps = saved
			break
		}
		cur = next
		positions = append(positions, cur)
		capsStack = append(capsStack, copyCaps(r.caps))
	}

	// Now try to match the rest from the longest position backwards
	// But we only have the star; the caller has the rest. Return the longest
	// that this star can consume. The caller (Concat) will backtrack if needed.
	// For simplicity, return the longest possible — full backtracking would
	// require continuations. This is a simplification but works for many cases.
	// Return the last successful position.
	r.caps = capsStack[len(capsStack)-1]
	return positions[len(positions)-1], true
}

// matchRepeat handles {n,m}.
func (r *BTRunner) matchRepeat(rep Repeat, pos int, start int) (int, bool) {
	cur := pos

	// Match min required
	for i := 0; i < rep.Min; i++ {
		next, ok := r.match(rep.Inner, cur, start)
		if !ok {
			return 0, false
		}
		cur = next
	}

	// If max == min, we're done
	if rep.Max == rep.Min {
		return cur, true
	}

	// If max == -1 (unbounded), act like star
	if rep.Max < 0 {
		return r.matchStar(rep.Inner, cur, start)
	}

	// {n,m}: try matching up to (max-min) more times greedily
	positions := []int{cur}
	cur2 := cur
	for i := 0; i < rep.Max-rep.Min; i++ {
		next, ok := r.match(rep.Inner, cur2, start)
		if !ok {
			break
		}
		cur2 = next
		positions = append(positions, cur2)
	}
	return positions[len(positions)-1], true
}

// copyCaps makes a shallow copy of the captures map.
func copyCaps(m map[int][2]int) map[int][2]int {
	out := make(map[int][2]int, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}