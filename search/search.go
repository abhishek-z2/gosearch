package search

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

func SearchFile(w io.Writer, file, query string, opts Options) (int, error) {
	fileHandle, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer fileHandle.Close()

	if opts.CaseInsensitive {
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

		if opts.CaseInsensitive {
			searchLine = strings.ToLower(line)
		}

		matched := strings.Contains(searchLine, query)

		if matched != opts.InvertMatch {
			matchCount++
			if !opts.CountOnly {
				fmt.Fprintf(w, "%s:%d: %s\n", file, lineNumber, line)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return matchCount, err
	}

	if opts.CountOnly && matchCount > 0 {
		fmt.Fprintf(w, "%s: %d matches\n", file, matchCount)
	}

	return matchCount, nil
}

func SearchDirectory(
	w io.Writer,
	dir string,
	query string,
	opts Options,
) error {
	totalCount := 0
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Fprintf(os.Stderr, "gosearch: %v\n", err)
			return nil
		}

		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}

			return nil
		}

		if !hasExtension(path, opts.Extensions) {
			return nil
		}

		binary, err := isBinary(path)
		if err != nil {
			return nil
		}
		if binary {
			return nil
		}

		matches, err := SearchFile(w, path, query, opts)
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

func SearchDirectoryConcurrent(
	w io.Writer,
	dir string,
	query string,
	opts Options,
) error {

	jobs := make(chan string, 100)
	results := make(chan Result, 100)

	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				matches, err := SearchFile(w, path, query, opts)
				results <- Result{Matches: matches, Err: err}
			}
		}()
	}

	go func() {
		defer close(jobs)
		_ = filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				fmt.Fprintf(os.Stderr, "gosearch:%v\n", err)
			}
			if d.IsDir() {
				if d.Name() == "git" {
					return filepath.SkipDir
				}
				return nil
			}
			if !hasExtension(path, opts.Extensions) {
				return nil
			}
			binary, err := isBinary(path)
			if err != nil || binary {
				return nil
			}

			jobs <- path
			return nil
		})
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	totalCount := 0
	for res := range results {
		if res.Err != nil {
			fmt.Fprintf(os.Stderr, "gosearch error reading file: %v\n", res.Err)
			continue
		}
		totalCount += res.Matches
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

// isBinary reads up to 512 bytes from path and returns true if a NUL byte is found.

func isBinary(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 { // NUL byte check
			return true, nil
		}
	}

	return false, nil
}
