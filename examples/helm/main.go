package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/ophis"
)

// Configuration constants
const (
	AppName    = "ophis"
	AppVersion = "0.0.0"
)

func slogToFile(level slog.Level, logFile string) *slog.Logger {
	// Try to create log file, fallback to stderr if it fails
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file %s: %v", logFile, err))
	}

	// Create handler with proper level setting
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(file, handlerOptions)
	logger := slog.New(handler)
	logger.Info("Logging initialized", "level", level.String(), "target", file.Name())
	return logger
}

func start(logFile string) error {
	var logger *slog.Logger

	if logFile == "" {
		// Use info level for production to reduce noise
		logger := slogToFile(slog.LevelDebug, logFile)
		logger.Info("Starting MCP bridge server", "app", AppName, "version", AppVersion)
	} else {
		logger = slog.Default()
	}

	bridge := ophis.NewCobraToMCPBridge(&HelmCommandFactory{}, AppName, AppVersion, logger)

	logger.Info("Bridge created with command factory, starting server...")
	err := bridge.StartServer()
	if err != nil {
		logger.Error("Server failed to start", "error", err)
	}
	return err
}

func main() {
	// Parse command line flags
	logFile := flag.String("logfile", "", "Path to the log file")
	flag.Parse()

	if err := start(*logFile); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
