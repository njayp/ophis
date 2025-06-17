package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ophis/bridge"
)

// Configuration constants
const (
	AppName    = "helm"
	AppVersion = "0.0.1"
)

func main() {
	// Parse command line flags
	p := flag.String("log-level", "", "slog log level")
	flag.Parse()
	loglevel := *p

	bridge, err := bridge.NewCobraToMCPBridge(&HelmCommandFactory{}, &bridge.MCPCommandConfig{
		AppName:    AppName,
		AppVersion: AppVersion,
		LogLevel:   loglevel,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating bridge: %v\n", err)
		os.Exit(1)
	}

	if err := bridge.StartServer(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
