package main

import (
	"fmt"
	"os"
)

func main() {
	//fmt.Println(os.Args)
	if len(os.Args) < 3 {
		fmt.Println("not enough arguments")
		return
	}
	query := os.Args[1]
	file := os.Args[2]
	fmt.Printf("query is %s , file is %s", query, file)
}
