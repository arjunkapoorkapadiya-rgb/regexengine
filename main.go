package main

import (
	"fmt"
	"os"
)

const usage = `regexengine — a regex engine built from scratch

usage:
  regexengine match <pattern> <text>       test if text matches pattern
  regexengine find <pattern> <text>        find first match (prints index + length)
  regexengine findall <pattern> <text>     find all matches
  regexengine replace <pattern> <text> <replacement>
  regexengine groups <pattern> <text>      show capture groups
  regexengine version                      print version

examples:
  regexengine match "abc" "xabcy"
  regexengine find "a+" "aaabbb"
  regexengine replace "\d+" "abc123def" "X"
`

const version = "0.2.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "match":
		if err := cmdMatch(args); err != nil {
			fatal(err)
		}
	case "find":
		if err := cmdFind(args); err != nil {
			fatal(err)
		}
	case "findall":
		if err := cmdFindAll(args); err != nil {
			fatal(err)
		}
	case "replace":
		if err := cmdReplace(args); err != nil {
			fatal(err)
		}
	case "groups":
		if err := cmdGroups(args); err != nil {
			fatal(err)
		}
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fatal(fmt.Errorf("unknown command: %s", cmd))
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "regexengine: %v\n", err)
	os.Exit(1)
}

func cmdMatch(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: regexengine match <pattern> <text>")
	}
	matched, err := Match(args[0], args[1])
	if err != nil {
		return err
	}
	if matched {
		fmt.Println("match")
	} else {
		fmt.Println("no match")
	}
	return nil
}

func cmdFind(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: regexengine find <pattern> <text>")
	}
	start, length, found, err := Find(args[0], args[1])
	if err != nil {
		return err
	}
	if found {
		fmt.Printf("match at index %d, length %d\n", start, length)
	} else {
		fmt.Println("no match")
	}
	return nil
}

func cmdFindAll(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: regexengine findall <pattern> <text>")
	}
	matches, err := FindAll(args[0], args[1])
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Println("no matches")
		return nil
	}
	for _, m := range matches {
		fmt.Printf("match at index %d, length %d: %q\n", m.Start, m.Length, args[1][m.Start:m.Start+m.Length])
	}
	return nil
}

func cmdReplace(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: regexengine replace <pattern> <text> <replacement>")
	}
	result, err := Replace(args[0], args[1], args[2])
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}

// cmdGroups prints capture groups for the first match.
func cmdGroups(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: regexengine groups <pattern> <text>")
	}
	pattern, text := args[0], args[1]

	// Count groups in pattern
	ast, err := compilePattern(pattern)
	if err != nil {
		return err
	}
	groupCount := countGroups(ast)

	if groupCount == 0 {
		fmt.Println("(no capture groups in pattern)")
		return nil
	}

	// For now, just show that we detect them
	result, found, err := FindWithGroups(pattern, text)
	if err != nil {
		return err
	}
	if !found {
		fmt.Println("no match")
		return nil
	}

	fmt.Printf("match at index %d, length %d\n", result.Start, result.Length)
	fmt.Printf("pattern has %d capture group(s)\n", groupCount)
	// Actual capture extraction is coming next.
	return nil
}

// countGroups counts Group nodes in the AST.
func countGroups(node Node) int {
	switch v := node.(type) {
	case Group:
		return 1 + countGroups(v.Inner)
	case Concat:
		n := 0
		for _, p := range v.Parts {
			n += countGroups(p)
		}
		return n
	case Alternate:
		n := 0
		for _, c := range v.Choices {
			n += countGroups(c)
		}
		return n
	case Star:
		return countGroups(v.Inner)
	case Plus:
		return countGroups(v.Inner)
	case Quest:
		return countGroups(v.Inner)
	case Repeat:
		return countGroups(v.Inner)
	}
	return 0
}