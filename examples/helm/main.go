package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ophis"
)

// Configuration constants
const (
	AppName    = "helm"
	AppVersion = "0.0.1"
)

func main() {
	// Parse command line flags
	p := flag.String("loglevel", "", "slog log level")
	flag.Parse()
	loglevel := *p

	bridge := ophis.NewCobraToMCPBridge(&HelmCommandFactory{}, &ophis.MCPCommandConfig{
		AppName:    AppName,
		AppVersion: AppVersion,
		LogLevel:   loglevel,
	})

	if err := bridge.StartServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
