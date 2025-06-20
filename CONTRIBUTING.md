# Contributing to Ophis

Thank you for your interest in contributing to Ophis! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [How to Contribute](#how-to-contribute)
- [Development Setup](#development-setup)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Feature Requests](#feature-requests)
- [Questions](#questions)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct:

- **Be respectful**: Treat everyone with respect. No harassment, discrimination, or inappropriate behavior will be tolerated.
- **Be collaborative**: Work together to resolve conflicts and assume good intentions.
- **Be inclusive**: Welcome and support people of all backgrounds and identities.
- **Be professional**: Maintain professionalism in all interactions.

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally
3. Create a new branch for your contribution
4. Make your changes
5. Push to your fork and submit a pull request

## How to Contribute

### Types of Contributions

We welcome various types of contributions:

- **Bug fixes**: Fix reported issues or bugs you discover
- **Features**: Implement new features or enhance existing ones
- **Documentation**: Improve documentation, add examples, or fix typos
- **Tests**: Add missing tests or improve test coverage
- **Performance**: Optimize code for better performance
- **Examples**: Add new example implementations

### Before You Start

1. Check if an issue already exists for your contribution
2. For significant changes, open an issue first to discuss your proposal
3. Ensure your contribution aligns with the project's goals and architecture

## Development Setup

### Prerequisites

- Go 1.24 or later
- golangci-lint (for linting)
- Make (for build automation)

### Setup Steps

1. **Clone the repository**
   ```bash
   git clone https://github.com/njayp/ophis.git
   cd ophis
   ```

2. **Install dependencies**
   ```bash
   make up
   ```

3. **Run tests to ensure everything works**
   ```bash
   make test
   ```

4. **Run the linter**
   ```bash
   make lint
   ```

### Project Structure

```
ophis/
├── bridge/           # Core MCP server bridge logic
├── tools/            # Command-to-tool conversion
├── mcp/              # Built-in MCP commands
│   └── claude/       # Claude Desktop integration
├── examples/         # Example implementations
└── tests/            # Test files
```

## Code Style

We follow standard Go conventions and use automated tooling to maintain consistency:

### Go Code Style

1. **Format your code**: Use `gofmt` and `gofumpt`
   ```bash
   make lint
   ```

2. **Follow Go conventions**:
   - Use meaningful variable and function names
   - Keep functions small and focused
   - Document exported types and functions
   - Handle errors explicitly
   - Use table-driven tests where appropriate

3. **Package organization**:
   - Keep packages focused on a single responsibility
   - Minimize dependencies between packages
   - Export only what's necessary

### Documentation

- Add godoc comments to all exported types, functions, and packages
- Include examples in documentation where helpful
- Keep comments up-to-date with code changes

Example:
```go
// CommandFactory provides a factory pattern for creating Cobra commands.
// It ensures fresh command instances for each MCP tool execution,
// preventing state pollution between calls.
type CommandFactory interface {
    // Tools returns all available MCP tools from your command tree.
    Tools() []tools.Tool
    
    // New creates a fresh command instance and execution function.
    New() (*cobra.Command, CommandExecFunc)
}
```

## Testing

### Writing Tests

1. **Test files**: Place tests in `*_test.go` files in the same package
2. **Test coverage**: Aim for high test coverage, especially for critical paths
3. **Test types**:
   - Unit tests for individual functions
   - Integration tests for component interactions
   - Table-driven tests for multiple scenarios

### Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./bridge/...

# Run tests with verbose output
go test -v ./...
```

### Test Guidelines

- Test both success and error cases
- Use meaningful test names that describe what's being tested
- Mock external dependencies when appropriate
- Keep tests deterministic and independent

Example:
```go
func TestCommandFactory_New(t *testing.T) {
    tests := []struct {
        name      string
        factory   CommandFactory
        wantError bool
    }{
        {
            name:      "creates fresh command instance",
            factory:   &MockFactory{},
            wantError: false,
        },
        // Add more test cases
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## Submitting Changes

### Pull Request Process

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clean, documented code
   - Add tests for new functionality
   - Update documentation as needed

3. **Commit your changes**
   - Use clear, descriptive commit messages
   - Follow conventional commit format if possible:
     ```
     type(scope): description
     
     Longer explanation if needed
     ```
   - Types: feat, fix, docs, style, refactor, test, chore

4. **Run checks locally**
   ```bash
   make all  # Runs format, lint, test, and build
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/your-feature-name
   ```

6. **Create a Pull Request**
   - Use the PR template (if available)
   - Provide a clear description of changes
   - Reference any related issues
   - Ensure CI checks pass

### Pull Request Guidelines

- **Title**: Clear and descriptive
- **Description**: Explain what, why, and how
- **Size**: Keep PRs focused and reasonably sized
- **Tests**: Include tests for new code
- **Documentation**: Update docs if needed
- **Breaking changes**: Clearly indicate if PR includes breaking changes

## Reporting Issues

### Before Creating an Issue

1. Search existing issues to avoid duplicates
2. Check if the issue is already fixed in the main branch
3. Collect relevant information about the problem

### Creating an Issue

Include the following information:

1. **Clear title**: Summarize the issue concisely
2. **Description**: Detailed explanation of the problem
3. **Steps to reproduce**: How to recreate the issue
4. **Expected behavior**: What should happen
5. **Actual behavior**: What actually happens
6. **Environment**:
   - Go version
   - Operating system
   - Ophis version
7. **Additional context**: Logs, screenshots, or code samples

## Feature Requests

We welcome feature requests! To propose a new feature:

1. **Check existing issues**: See if someone already requested it
2. **Create a feature request issue** with:
   - Clear description of the feature
   - Use cases and benefits
   - Potential implementation approach
   - Any alternatives considered
3. **Be patient**: Features are prioritized based on project goals

## Questions

For questions about using or contributing to Ophis:

1. Check the documentation and README
2. Search existing issues and discussions
3. Create a new issue with the "question" label
4. Be specific about what you're trying to achieve

## Recognition

Contributors will be recognized in the following ways:

- Inclusion in release notes
- Credit in the commit history
- Potential addition to a CONTRIBUTORS file

## License

By contributing to Ophis, you agree that your contributions will be licensed under the Apache License 2.0.

---

Thank you for contributing to Ophis! Your efforts help make this project better for everyone.
