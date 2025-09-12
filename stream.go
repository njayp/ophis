package ophis

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
)

// StreamCommandFlags holds configuration flags for the stream command.
type StreamCommandFlags struct {
	LogLevel string
	Address  string
	CertFile string
	KeyFile  string
}

// streamCommand creates a Cobra command for starting the MCP server in HTTP streaming mode.
func streamCommand(config *Config) *cobra.Command {
	streamFlags := &StreamCommandFlags{}
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Start the MCP server in HTTP streaming mode",
		Long: `Start MCP server in HTTP streaming mode to expose CLI commands to AI assistants.
This mode allows connections over HTTP with server-sent events (SSE) for real-time streaming.`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if config == nil {
				config = &Config{}
			}

			if streamFlags.LogLevel != "" {
				level := parseLogLevel(streamFlags.LogLevel)
				// Ensure SloggerOptions is initialized
				if config.SloggerOptions == nil {
					config.SloggerOptions = &slog.HandlerOptions{}
				}
				// Set the log level based on the flag
				config.SloggerOptions.Level = level
			}

			// Validate address
			if streamFlags.Address == "" {
				return fmt.Errorf("address is required for HTTP streaming mode")
			}

			switch {
			case (streamFlags.CertFile == "") != (streamFlags.KeyFile == ""):
				return fmt.Errorf("both cert-file and key-file must be provided to enable TLS")
			case streamFlags.CertFile != "" && streamFlags.KeyFile != "":
				slog.Info("TLS enabled for HTTP streaming server")
				config.StreamOptions = append(config.StreamOptions,
					server.WithTLSCert(streamFlags.CertFile, streamFlags.KeyFile),
				)
			default:
				slog.Info("TLS not enabled; using plain HTTP for streaming server")
			}

			// Create and start the bridge in HTTP mode
			return config.bridgeConfig(cmd).ServeHTTP(streamFlags.Address)
		},
	}

	// Add flags
	flags := cmd.Flags()
	flags.StringVar(&streamFlags.LogLevel, "log-level", "", "Log level (debug, info, warn, error)")
	flags.StringVarP(&streamFlags.Address, "address", "a", ":8080", "Address to listen on (e.g., :8080, localhost:3000)")
	flags.StringVar(&streamFlags.CertFile, "cert-file", "", "Path to TLS certificate file (enables HTTPS)")
	flags.StringVar(&streamFlags.KeyFile, "key-file", "", "Path to TLS key file (enables HTTPS)")

	return cmd
}
