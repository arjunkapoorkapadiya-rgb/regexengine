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
  regexengine version                      print version

examples:
  regexengine match "abc" "xabcy"
  regexengine find "a+" "aaabbb"
  regexengine replace "\d+" "abc123def" "X"
`

const version = "0.1.0"

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
	pattern, text := args[0], args[1]

	matched, err := Match(pattern, text)
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
	pattern, text := args[0], args[1]

	start, length, found, err := Find(pattern, text)
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
	pattern, text := args[0], args[1]

	matches, err := FindAll(pattern, text)
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		fmt.Println("no matches")
		return nil
	}
	for _, m := range matches {
		fmt.Printf("match at index %d, length %d: %q\n", m.Start, m.Length, text[m.Start:m.Start+m.Length])
	}
	return nil
}

func cmdReplace(args []string) error {
	if len(args) != 3 {
		return fmt.Errorf("usage: regexengine replace <pattern> <text> <replacement>")
	}
	pattern, text, replacement := args[0], args[1], args[2]

	result, err := Replace(pattern, text, replacement)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}