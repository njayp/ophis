package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/ophis"
	"github.com/ophis/cmds/basic"
)

// Configuration constants
const (
	AppName    = "ophis"
	AppVersion = "0.0.1"
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
	logger := slogToFile(slog.LevelInfo)
	logger.Info("Starting MCP bridge server", "app", AppName, "version", AppVersion)

	// Create a factory function that generates fresh command trees for each execution
	/*
		commandFactory := func() *cobra.Command {
			// Create a root command that includes multiple tool sets
			rootCmd := &cobra.Command{
				Use:   AppName,
				Short: "Ophis MCP Server - Multiple CLI Tools Bridge",
				Long:  "Ophis converts multiple CLI tools into MCP (Model Context Protocol) servers, making them accessible to AI assistants.",
			}

			// Add all available tool commands
			rootCmd.AddCommand(basic.NewRootCmd())

			return rootCmd
		}
	*/

	bridge := ophis.NewCobraToMCPBridge(basic.NewRootCmd, AppName, AppVersion, logger)

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
