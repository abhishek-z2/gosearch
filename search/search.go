package search

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func SearchFile(w io.Writer, file, query string, caseInsensitive bool) (int, error) {
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

func SearchDirectory(
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

		matches, err := SearchFile(w, path, query, caseInsensitive)
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
