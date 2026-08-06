package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gosearch <query> <filename>")
		return
	}
	query := os.Args[1]
	file := os.Args[2]

	//checks if its a dir or a file
	info, err := os.Stat(file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	} else {
		fmt.Println("File Info: ", info.Name())
	}

	//searches the file
	err = searchFile(file, query)
	if err != nil {
		fmt.Println("Error searching the file: ", err)
		return
	}

}
func searchFile(file, query string) error {
	fmt.Println(file, query)
	filehandle, err := os.Open(file)
	if err != nil {
		return err
	}
	defer filehandle.Close()

	scanner := bufio.NewScanner(filehandle)
	ln := 0
	found := 0
	for scanner.Scan() {
		ln++
		if line := scanner.Text(); strings.Contains(line, query) {
			fmt.Printf("%v: %v\n", ln, line)
			found++
		}
	}
	if found == 0 {
		fmt.Println("no matches found")
	} else {
		fmt.Printf("%d matches found", found)
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return err
	}
	return nil
}
