package claude

import "fmt"

// Server represents an MCP server configuration entry for Claude Desktop.
type Server struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// Print displays the server configuration details.
func (s Server) Print() {
	fmt.Printf("  Command: %s\n", s.Command)
	if len(s.Args) > 0 {
		fmt.Printf("  Args: %v\n", s.Args)
	}
	if len(s.Env) > 0 {
		fmt.Printf("  Environment:\n")
		for key, value := range s.Env {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
}
