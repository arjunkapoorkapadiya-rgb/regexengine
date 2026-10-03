# regexengine

A regex engine built from scratch in Go. No `regexp` package. No libraries. Just:
lexer → parser → NFA compiler → simulator.

## Features

- **Lexer** — tokenizes pattern strings
- **Recursive-descent parser** — builds an AST with proper operator precedence
- **Thompson's construction** — compiles AST → NFA (nondeterministic finite automaton)
- **NFA simulator** — matches text using ε-closure + state sets (no backtracking)
- **Command-line interface** — `match`, `find`, `findall`, `replace`

## Supported regex syntax

| Syntax | Meaning |
|---|---|
| `abc` | Literal characters |
| `.` | Any character |
| `a*` | Zero or more |
| `a+` | One or more |
| `a?` | Zero or one |
| `a{3}` | Exactly 3 |
| `a{2,5}` | Between 2 and 5 |
| `a{2,}` | 2 or more |
| `a\|b` | Alternation (OR) |
| `(...)` | Grouping |
| `[...]` | Character class |
| `[a-z]` | Range |
| `[^abc]` | Negated class |
| `\d` | Digit `[0-9]` |
| `\w` | Word `[a-zA-Z0-9_]` |
| `\s` | Whitespace |
| `^` | Start of text |
| `$` | End of text |

## Installation

```bash
git clone https://github.com/arjunkapoorkapadiya-rgb/regexengine.git
cd regexengine
go build -o regexengine.exe
```

## Usage

```bash
# Test if a pattern matches anywhere in the text
regexengine match "abc" "xabcy"
# → match

# Find the first match
regexengine find "a+" "xxaaayy"
# → match at index 2, length 3

# Find all matches
regexengine findall "[a-z]+" "abc 123 xyz 456"
# → match at index 0, length 3: "abc"
# → match at index 8, length 3: "xyz"

# Replace all matches
regexengine replace "\d+" "abc123def456" "X"
# → abcXdefX
```

## Architecture

```
pattern string
      │
      ▼
   ┌──────┐
   │Lexer │  → tokens (LITERAL, STAR, LPAREN, ...)
   └──────┘
      │
      ▼
   ┌──────┐
   │Parser│  → AST (Literal, Concat, Star, Alternate, ...)
   └──────┘
      │
      ▼
   ┌──────┐
   │Compile│ → NFA (Thompson's construction)
   └──────┘
      │
      ▼
   ┌──────────┐
   │Simulator │ → match / no match + positions
   └──────────┘
```

## File map

| File | Purpose |
|---|---|
| `main.go` | CLI entry point + subcommands |
| `token.go` | Token type definitions |
| `lexer.go` | Pattern string → token stream |
| `ast.go` | AST node definitions |
| `parser.go` | Token stream → AST (recursive descent) |
| `nfa.go` | AST → NFA (Thompson's construction) |
| `match.go` | NFA simulator: match, find, findall, replace |
| `*_test.go` | Tests for each component |

## Testing

```bash
go test -v
```

All tests pass. Coverage spans: lexer, parser, NFA compilation, matching, finding, replacing.

## Why build this?

Building a regex engine from scratch teaches:
- **Lexing & parsing** — the same techniques used in compilers
- **Automata theory** — NFAs, ε-closures, Thompson's construction
- **Backtracking vs. set-based matching** — why some engines are fast and others aren't
- **Content-addressable design** — how to structure a clean compiler pipeline

## LOC

~1800 lines of Go across 7 source files and 4 test files.

## License

MIT