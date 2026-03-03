package main

import (
	"fmt"
	"os"

	"github.com/kevinle-00/fornax/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
