package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ophis"
)

// Configuration constants
const (
	AppName    = "ophis"
	AppVersion = "0.0.0"
	LogFile    = "/Users/nickpowell/claude/ophis/app.log"
)

func slogToFile(level slog.Level) *slog.Logger {
	// Try to create log file, fallback to stderr if it fails
	logFile, err := os.OpenFile(LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file %s: %v", LogFile, err))
	}

	// Create handler with proper level setting
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(logFile, handlerOptions)
	logger := slog.New(handler)
	logger.Info("Logging initialized", "level", level.String(), "target", logFile.Name())
	return logger
}

func start() error {
	// Set environment variable to indicate MCP server is running
	os.Setenv("MCP_SERVER_RUNNING", "true")

	// Use info level for production to reduce noise
	logger := slogToFile(slog.LevelDebug)
	logger.Info("Starting MCP bridge server", "app", AppName, "version", AppVersion)

	bridge := ophis.NewCobraToMCPBridge(&HelmCommandFactory{}, AppName, AppVersion, logger)

	logger.Info("Bridge created with command factory, starting server...")
	err := bridge.StartServer()
	if err != nil {
		logger.Error("Server failed to start", "error", err)
	}
	return err
}

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
