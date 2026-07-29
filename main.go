package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: gosearch <query> <filename>")
		return
	}
	//query := os.Args[1]
	file := os.Args[2]

	filehandle, err := os.Open(file)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	defer filehandle.Close()

	scanner := bufio.NewScanner(filehandle)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
