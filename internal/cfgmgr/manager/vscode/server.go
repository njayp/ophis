package vscode

import "fmt"

// MCPServer represents an MCP server configuration entry for VSCode.
type MCPServer struct {
	Type    string            `json:"type,omitempty"`
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
	URL     string            `json:"url,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Print displays the server configuration details.
func (s MCPServer) Print() {
	fmt.Printf("  Type: %s\n", s.Type)
	if s.Command != "" {
		fmt.Printf("  Command: %s\n", s.Command)
	}
	if s.URL != "" {
		fmt.Printf("  URL: %s\n", s.URL)
	}
	if len(s.Args) > 0 {
		fmt.Printf("  Args: %v\n", s.Args)
	}
	if len(s.Env) > 0 {
		fmt.Printf("  Environment:\n")
		for key, value := range s.Env {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
	if len(s.Headers) > 0 {
		fmt.Printf("  Headers:\n")
		for key, value := range s.Headers {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
}
