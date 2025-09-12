package bridge

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/mark3labs/mcp-go/server"
)

// ServeIO starts the MCP server using stdio for communication.
func (c *Config) ServeIO() error {
	m := c.newManager()
	return server.ServeStdio(m.server)
}

// ServeHTTP starts the MCP server with an HTTP transport on the specified address.
func (c *Config) ServeHTTP(addr string) error {
	m := c.newManager()
	server := server.NewStreamableHTTPServer(m.server, c.StreamOptions...)

	// shutdown server gracefully on interrupt signal
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		slog.Info("shutting down MCP HTTP server")
		if err := server.Shutdown(context.Background()); err != nil {
			slog.Error("error shutting down MCP HTTP server", "error", err)
		}
	}()

	slog.Info("starting MCP HTTP server", "address", addr)
	return server.Start(addr)
}
