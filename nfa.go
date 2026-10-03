package main

// StateID is an index into NFA.States.
type StateID int

// TransitionKind describes what kind of edge this is.
type TransitionKind int

const (
	TransEpsilon TransitionKind = iota // ε (empty) transition — always taken
	TransLiteral                       // matches a specific byte
	TransAny                           // matches any byte (except newline by default)
	TransClass                         // matches if byte is in the class
	TransAnchorStart                   // zero-width: matches at start of text
	TransAnchorEnd                     // zero-width: matches at end of text
)

// Transition is one edge out of a state.
type Transition struct {
	Kind  TransitionKind
	To    StateID
	Ch    byte      // for TransLiteral
	Class CharClass // for TransClass
}

// State is one node in the NFA.
type State struct {
	ID          StateID
	Transitions []Transition
}

// NFA is a compiled regex.
type NFA struct {
	States []State
	Start  StateID
	Accept StateID
}

// NewNFA creates an empty NFA.
func NewNFA() *NFA {
	return &NFA{}
}

// addState creates a new state and returns its ID.
func (n *NFA) addState() StateID {
	id := StateID(len(n.States))
	n.States = append(n.States, State{ID: id})
	return id
}

// addTransition adds an edge from -> to.
func (n *NFA) addTransition(from StateID, t Transition) {
	n.States[from].Transitions = append(n.States[from].Transitions, t)
}

// Compile builds an NFA from an AST.
func Compile(root Node) (*NFA, error) {
	n := NewNFA()
	start, accept := compileNode(n, root)
	n.Start = start
	n.Accept = accept
	return n, nil
}

// compileNode returns (start, accept) state IDs for the given AST node.
func compileNode(n *NFA, node Node) (StateID, StateID) {
	switch v := node.(type) {
	case Literal:
		s := n.addState()
		e := n.addState()
		n.addTransition(s, Transition{Kind: TransLiteral, To: e, Ch: v.Ch})
		return s, e

	case Any:
		s := n.addState()
		e := n.addState()
		n.addTransition(s, Transition{Kind: TransAny, To: e})
		return s, e

	case CharClass:
		s := n.addState()
		e := n.addState()
		n.addTransition(s, Transition{Kind: TransClass, To: e, Class: v})
		return s, e

	case Concat:
		if len(v.Parts) == 0 {
			return compileEmpty(n)
		}
		start, end := compileNode(n, v.Parts[0])
		for i := 1; i < len(v.Parts); i++ {
			s, e := compileNode(n, v.Parts[i])
			n.addTransition(end, Transition{Kind: TransEpsilon, To: s})
			end = e
		}
		return start, end

	case Alternate:
		s := n.addState()
		e := n.addState()
		for _, ch := range v.Choices {
			cs, ce := compileNode(n, ch)
			n.addTransition(s, Transition{Kind: TransEpsilon, To: cs})
			n.addTransition(ce, Transition{Kind: TransEpsilon, To: e})
		}
		return s, e

	case Star:
		s := n.addState()
		e := n.addState()
		is, ie := compileNode(n, v.Inner)
		n.addTransition(s, Transition{Kind: TransEpsilon, To: is})
		n.addTransition(s, Transition{Kind: TransEpsilon, To: e})
		n.addTransition(ie, Transition{Kind: TransEpsilon, To: is})
		n.addTransition(ie, Transition{Kind: TransEpsilon, To: e})
		return s, e

	case Plus:
		s := n.addState()
		e := n.addState()
		is, ie := compileNode(n, v.Inner)
		n.addTransition(s, Transition{Kind: TransEpsilon, To: is})
		n.addTransition(ie, Transition{Kind: TransEpsilon, To: is})
		n.addTransition(ie, Transition{Kind: TransEpsilon, To: e})
		return s, e

	case Quest:
		s := n.addState()
		e := n.addState()
		is, ie := compileNode(n, v.Inner)
		n.addTransition(s, Transition{Kind: TransEpsilon, To: is})
		n.addTransition(s, Transition{Kind: TransEpsilon, To: e})
		n.addTransition(ie, Transition{Kind: TransEpsilon, To: e})
		return s, e

	case Repeat:
		return compileRepeat(n, v)

	case Group:
		return compileNode(n, v.Inner)

	case AnchorStart:
		s := n.addState()
		e := n.addState()
		n.addTransition(s, Transition{Kind: TransAnchorStart, To: e})
		return s, e

	case AnchorEnd:
		s := n.addState()
		e := n.addState()
		n.addTransition(s, Transition{Kind: TransAnchorEnd, To: e})
		return s, e

	case Empty:
		return compileEmpty(n)

	default:
		return compileEmpty(n)
	}
}

// compileEmpty creates an ε-only path from start to accept.
func compileEmpty(n *NFA) (StateID, StateID) {
	s := n.addState()
	e := n.addState()
	n.addTransition(s, Transition{Kind: TransEpsilon, To: e})
	return s, e
}

// compileRepeat handles {min, max}.
//   - max == -1: append star of one more copy ({n,})
//   - max == min: exactly min copies ({n})
//   - otherwise: min copies + (max-min) optional copies ({n,m})
func compileRepeat(n *NFA, r Repeat) (StateID, StateID) {
	if r.Min < 0 {
		r.Min = 0
	}
	if r.Max >= 0 && r.Max < r.Min {
		r.Max = r.Min
	}

	var start, end StateID
	first := true

	// min copies concatenated
	for i := 0; i < r.Min; i++ {
		s, e := compileNode(n, r.Inner)
		if first {
			start = s
			first = false
		} else {
			n.addTransition(end, Transition{Kind: TransEpsilon, To: s})
		}
		end = e
	}

	if first {
		// min was 0 — we haven't created anything yet
		start = n.addState()
		end = start
	}

	if r.Max < 0 {
		// {n,} — star of one more copy
		s, e := compileNode(n, Star{Inner: r.Inner})
		if r.Min == 0 {
			// Simplify: {0,} == star
			return s, e
		}
		n.addTransition(end, Transition{Kind: TransEpsilon, To: s})
		end = e
	} else {
		// {n,m} — add (m-n) optional copies
		for i := r.Min; i < r.Max; i++ {
			s, e := compileNode(n, Quest{Inner: r.Inner})
			if r.Min == 0 && i == 0 {
				// Simplify first optional: chain directly
				start = s
				end = e
				continue
			}
			n.addTransition(end, Transition{Kind: TransEpsilon, To: s})
			end = e
		}
	}

	return start, end
}