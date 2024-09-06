package main

import (
	"fmt"
	"os"
	"tamper/core"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: go run main.go <url>")
		os.Exit(1)
	}

	url := os.Args[1]
	core.Checker([]string{url})
}
