package main

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/ophis"
	"github.com/ophis/terraform"
)

const logFilePath = "/Users/nickpowell/claude/ophis/app.log"

func slogToFile(level slog.Level) {
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(fmt.Sprintf("Failed to open log file: %v", err))
	}

	// Create handler with proper level setting
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(logFile, handlerOptions)
	slog.SetDefault(slog.New(handler))

	slog.Info("Logging initialized", "level", level.String(), "file", "/Users/nickpowell/claude/cobra-mcp-bridge/app.log")
}

func start() error {
	slogToFile(slog.LevelDebug)
	appName := "ophis"
	slog.Info("Starting MCP bridge server", "app", appName, "version", "1.0.0")

	// Basic hello world commands
	//cmd := basic.NewRootCmd()

	cmd := terraform.CreateTerraformCmd()

	bridge := ophis.NewCobraToMCPBridge(cmd, appName, "0.0.0")
	slog.Info("Bridge created, starting server...")

	err := bridge.StartServer()
	if err != nil {
		slog.Error("Server failed to start", "error", err)
	}
	return err
}

func main() {
	if err := start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
