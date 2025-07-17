package main

import (
	"os"
)

func main() {
	if err := createMakeCommands().Execute(); err != nil {
		os.Exit(1)
	}
}
