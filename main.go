package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
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

	caseInsensitive := flags.Bool("i", false, "case-insensitive search")
	recursive := flags.Bool("r", false, "search directories recursively")
	extension := flags.String("e", "", "only search files with this extension")

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

		err = searchDirectory(os.Stdout, path, query, *caseInsensitive, extensions)
		if err != nil {
			return fmt.Errorf("error searching directory: %w", err)
		}

		return nil
	}

	matches, err := searchFile(os.Stdout, path, query, *caseInsensitive)

	if err != nil {
		return fmt.Errorf("error searching file: %w", err)
	}

	if matches == 0 {
		fmt.Println("no matches found")
	} else {
		fmt.Printf("%d matches found\n", matches)
	}
	return nil
}

func searchFile(w io.Writer, file, query string, caseInsensitive bool) (int, error) {
	fileHandle, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer fileHandle.Close()

	if caseInsensitive {
		query = strings.ToLower(query)
	}

	scanner := bufio.NewScanner(fileHandle)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	lineNumber := 0
	matchCount := 0

	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		searchLine := line

		if caseInsensitive {
			searchLine = strings.ToLower(line)
		}

		if strings.Contains(searchLine, query) {
			matchCount++
			fmt.Fprintf(w, "%s:%d: %s\n", file, lineNumber, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return matchCount, err
	}

	return matchCount, nil
}

func searchDirectory(
	w io.Writer,
	dir string,
	query string,
	caseInsensitive bool,
	extensions []string,
) error {
	totalCount := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if !hasExtension(path, extensions) {
			return nil
		}

		matches, err := searchFile(w, path, query, caseInsensitive)
		totalCount += matches
		if err != nil {
			return fmt.Errorf("error searching file: %w", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if totalCount == 0 {
		fmt.Fprintln(w, "no matches found")
	} else {
		fmt.Fprintf(w, "%d matches found\n", totalCount)
	}
	return nil
}

func hasExtension(path string, extensions []string) bool {
	if len(extensions) == 0 {
		return true
	}

	extension := strings.TrimPrefix(filepath.Ext(path), ".")

	for _, allowed := range extensions {
		if extension == allowed {
			return true
		}
	}

	return false
}

func printMatches(matches []Match) {
	for _, match := range matches {
		fmt.Printf("%v:%v: %v\n",
			match.File,
			match.LineNumber,
			match.Line,
		)
	}
}
