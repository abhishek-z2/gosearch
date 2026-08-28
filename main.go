package main

import (
	"flag"
	"fmt"
	"gosearch/search"
	"os"
	"strings"
)

type Match struct {
	File       string
	LineNumber int
	Line       string
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {

	flags := flag.NewFlagSet("gosearch", flag.ContinueOnError)

	invertMatch := flags.Bool("v", false, "invert match: select non-matching lines")
	caseInsensitive := flags.Bool("i", false, "case-insensitive search")
	recursive := flags.Bool("r", false, "search directories recursively")
	countOnly := flags.Bool("c", false, "returns only the number of matches")
	extension := flags.String("e", "", "only search files with this extension")
	afterContext := flags.Int("A", 0, "print N lines after match")
	beforeContext := flags.Int("B", 0, "print N lines before match")
	colorOutput := flags.Bool("color", false, "enable highlighted color ouput")

	if err := flags.Parse(args); err != nil {
		return err
	}

	var extensions []string

	if *extension != "" {
		extensions = strings.Split(*extension, ",")

		for i := range extensions {
			extensions[i] = strings.TrimPrefix(strings.TrimSpace(extensions[i]), ".")
		}
	}

	opts := search.Options{
		CaseInsensitive: *caseInsensitive,
		InvertMatch:     *invertMatch,
		Recursive:       *recursive,
		CountOnly:       *countOnly,
		Extensions:      extensions,
		AfterContext:    *afterContext,
		BeforeContext:   *beforeContext,
		Color:           *colorOutput,
	}

	args = flags.Args()
	flags.SetOutput(os.Stderr)
	if len(args) < 2 {
		flags.PrintDefaults() // uses local flag instance
		return fmt.Errorf("usage: gosearch [options] <query> <filepath>")
	}

	query := args[0]
	path := args[1]

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("couldn't find path %q: %w", path, err)
	}

	if info.IsDir() {
		if !*recursive {
			return fmt.Errorf("Error: path is a directory (use -r to search recursively)")
		}

		err = search.SearchDirectoryConcurrent(os.Stdout, path, query, opts)
		if err != nil {
			return fmt.Errorf("error searching directory: %w", err)
		}

		return nil
	}

	results, matches, err := search.SearchFile(path, query, opts)

	if err != nil {
		return fmt.Errorf("error searching file: %w", err)
	}
	search.PrintResult(os.Stdout, results, query, opts)

	if matches == 0 {
		fmt.Println("no matches found")
	} else {
		fmt.Printf("%d total matches found\n", matches)
	}
	return nil
}
