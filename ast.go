package main

// AST for regex patterns. We parse a pattern string into one of these nodes.

// Node is any AST node.
type Node interface {
	node()
}

// Literal matches a single literal character.
type Literal struct {
	Ch byte
}

func (Literal) node() {}

// Any matches any single character (the "." operator).
type Any struct{}

func (Any) node() {}

// CharClass matches any single character in a set (e.g. [a-z0-9]).
// Ranges are stored as pairs. Negation flips the set.
type CharClass struct {
	Negated bool
	Ranges  [][2]byte // each is [lo, hi] inclusive
}

func (CharClass) node() {}

// Concat matches a sequence of nodes in order.
type Concat struct {
	Parts []Node
}

func (Concat) node() {}

// Alternate matches any of the children (the "|" operator).
type Alternate struct {
	Choices []Node
}

func (Alternate) node() {}

// Star matches zero or more of the inner node (the "*" operator).
type Star struct {
	Inner Node
}

func (Star) node() {}

// Plus matches one or more of the inner node (the "+" operator).
type Plus struct {
	Inner Node
}

func (Plus) node() {}

// Quest matches zero or one of the inner node (the "?" operator).
type Quest struct {
	Inner Node
}

func (Quest) node() {}

// Repeat matches a bounded repetition {min,max}. If Max < 0, it's unbounded.
type Repeat struct {
	Inner Node
	Min   int
	Max   int // -1 means unlimited
}

func (Repeat) node() {}

// Group captures a subexpression. Index identifies the group.
type Group struct {
	Index int
	Inner Node
}

func (Group) node() {}

// AnchorStart matches the beginning of text.
type AnchorStart struct{}

func (AnchorStart) node() {}

// AnchorEnd matches the end of text.
type AnchorEnd struct{}

func (AnchorEnd) node() {}

// Empty matches nothing (used as a placeholder).
type Empty struct{}

func (Empty) node() {}