package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	caseInsensitive := flag.Bool("i", false, "case-insensitive search")
	recursive := flag.Bool("r", false, "search directories recursively")
	extension := flag.String("e", "", "only search files with this extension")

	flag.Parse()

	var extensions []string

	if *extension != "" {
		extensions = strings.Split(*extension, ",")

		for i := range extensions {
			extensions[i] = strings.TrimPrefix(strings.TrimSpace(extensions[i]), ".")
		}
	}

	args := flag.Args()

	if len(args) < 2 {
		fmt.Println("Usage: gosearch [options] <query> <filepath>")
		flag.PrintDefaults()
		return
	}

	query := args[0]
	path := args[1]

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if info.IsDir() {
		if !*recursive {
			fmt.Println("Error: path is a directory (use -r to search recursively)")
			return
		}

		err = searchDirectory(path, query, *caseInsensitive, extensions)
		if err != nil {
			fmt.Println("Error searching directory:", err)
			return
		}

		return
	}

	found, err := searchFile(path, query, *caseInsensitive)
	if err != nil {
		fmt.Println("Error searching the file:", err)
		return
	}

	if found == 0 {
		fmt.Println("no matches found")
	} else {
		fmt.Printf("%d matches found\n", found)
	}
}

func searchFile(file, query string, caseInsensitive bool) (int, error) {
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
	found := 0

	for scanner.Scan() {
		lineNumber++

		line := scanner.Text()
		searchLine := line

		if caseInsensitive {
			searchLine = strings.ToLower(line)
		}

		if strings.Contains(searchLine, query) {
			fmt.Printf("%s:%d: %s\n", file, lineNumber, line)
			found++
		}
	}

	if err := scanner.Err(); err != nil {
		return found, err
	}

	return found, nil
}

func searchDirectory(
	dir string,
	query string,
	caseInsensitive bool,
	extensions []string,
) error {
	totalMatches := 0

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

		found, err := searchFile(path, query, caseInsensitive)
		if err != nil {
			return err
		}

		totalMatches += found

		return nil
	})

	if err != nil {
		return err
	}

	if totalMatches == 0 {
		fmt.Println("no matches found")
	} else {
		fmt.Printf("%d matches found\n", totalMatches)
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
