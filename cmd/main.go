package main

import (
	"fmt"
	"os"

	"log/slog"

	"github.com/ophis"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "myapp",
	Short: "My CLI application",
	Long:  "A longer description of my CLI application",
}

var helloCmd = &cobra.Command{
	Use:   "hello [name]",
	Short: "Say hello to someone",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := "World"
		if len(args) > 0 {
			name = args[0]
		}
		greeting, _ := cmd.Flags().GetString("greeting")
		cmd.Printf("%s, %s!\n", greeting, name)
	},
}

func init() {
	helloCmd.Flags().String("greeting", "Hello", "The greeting to use")
	rootCmd.AddCommand(helloCmd)
}

func slogToFile(level slog.Level) {
	logFile, err := os.OpenFile("/Users/nickpowell/claude/cobra-mcp-bridge/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
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

func start() error { // Create the Cobra to MCP bridge
	slogToFile(slog.LevelDebug)
	slog.Info("Starting MCP bridge server", "app", "myapp", "version", "1.0.0")

	bridge := ophis.NewCobraToMCPBridge(rootCmd, "myapp", "1.0.0")
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
