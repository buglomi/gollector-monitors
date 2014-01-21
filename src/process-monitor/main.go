package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("enter the full path to a binary")
		os.Exit(1)
	}

	content, _ := json.Marshal(GetPids(os.Args[1:]...))

	fmt.Println(string(content))
}
