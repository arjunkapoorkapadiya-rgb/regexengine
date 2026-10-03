package main

import (
	"fmt"
	"os"
	"sort"
)

const usage = `regexengine — a regex engine built from scratch

usage:
  regexengine match <pattern> <text>       test if text matches pattern
  regexengine find <pattern> <text>        find first match
  regexengine findall <pattern> <text>     find all matches
  regexengine replace <pattern> <text> <replacement>
  regexengine groups <pattern> <text>      show capture groups
  regexengine version                      print version
`

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		fmt.Print(usage)
		os.Exit(1)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "match":
		handle(cmdMatch(args))
	case "find":
		handle(cmdFind(args))
	case "findall":
		handle(cmdFindAll(args))
	case "replace":
		handle(cmdReplace(args))
	case "groups":
		handle(cmdGroups(args))
	case "version":
		fmt.Println(version)
	case "help", "-h", "--help":
		fmt.Print(usage)
	default:
		fatal(fmt.Errorf("unknown command: %s", cmd))
	}
}

func handle(err error) {
	if err != nil {
		fatal(err)
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

func cmdGroups(args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("usage: regexengine groups <pattern> <text>")
	}
	pattern, text := args[0], args[1]

	start, end, captures, found, err := MatchWithCaptures(pattern, text)
	if err != nil {
		return err
	}
	if !found {
		fmt.Println("no match")
		return nil
	}

	fmt.Printf("match at index %d, length %d\n", start, end-start)
	fmt.Printf("matched: %q\n", text[start:end])

	if len(captures) == 0 {
		fmt.Println("(no capture groups)")
		return nil
	}

	// Sort group indices
	indices := make([]int, 0, len(captures))
	for i := range captures {
		indices = append(indices, i)
	}
	sort.Ints(indices)

	for _, i := range indices {
		fmt.Printf("  group %d: %q\n", i, captures[i])
	}
	return nil
}