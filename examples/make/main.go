package main

import (
	"os"
)

// Configuration constants
const (
	AppName    = "make"
	AppVersion = "0.0.1"
)

func main() {
	if err := createMakeCommands().Execute(); err != nil {
		os.Exit(1)
	}
}
