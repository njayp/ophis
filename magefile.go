//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

// All runs 'make all' for each example directory in parallel
func All() error {
	return runMakeInExamples("all")
}

// Build runs 'make build' for each example directory in parallel
func Build() error {
	return runMakeInExamples("build")
}

func Test() error {
	return runMakeInExamples("test")
}

// Claude runs './bin/<cli> mcp claude enable' for every binary in bin/
func Claude() error {
	return enableMCP("claude")
}

// Vscode runs './bin/<cli> mcp vscode enable' for every binary in bin/
func Vscode() error {
	return enableMCP("vscode")
}

// taskConfig defines a task to run in parallel
type taskConfig struct {
	name        string
	description string
	cmd         func(item string) *exec.Cmd
}

// runParallelTasks runs commands in parallel and reports results
func runParallelTasks(items []string, cfg taskConfig) error {
	if len(items) == 0 {
		return fmt.Errorf("no items to process")
	}

	type result struct {
		item string
		err  error
	}

	var wg sync.WaitGroup
	resultChan := make(chan result, len(items))

	for _, item := range items {
		wg.Add(1)
		go func(item string) {
			defer wg.Done()

			fmt.Printf("%s: %s\n", cfg.description, item)
			cmd := cfg.cmd(item)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				fmt.Printf("✗ Failed: %s\n", item)
				resultChan <- result{item: item, err: err}
			} else {
				fmt.Printf("✓ Success: %s\n", item)
				resultChan <- result{item: item, err: nil}
			}
		}(item)
	}

	wg.Wait()
	close(resultChan)

	// Collect results
	var failures []string
	var successCount int
	for res := range resultChan {
		if res.err != nil {
			failures = append(failures, fmt.Sprintf("  - %s: %v", res.item, res.err))
		} else {
			successCount++
		}
	}

	// Report final status
	if len(failures) > 0 {
		fmt.Printf("\n✗ %s failed for %d item(s):\n", cfg.name, len(failures))
		fmt.Println(strings.Join(failures, "\n"))
		fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, len(failures))
		return fmt.Errorf("%s failed for %d item(s)", cfg.name, len(failures))
	}

	fmt.Printf("\n✓ Successfully ran %s for all %d items\n", cfg.name, successCount)
	return nil
}

// runMakeInExamples runs a make target in all example directories in parallel
func runMakeInExamples(target string) error {
	examplesDir := "examples"
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		return fmt.Errorf("failed to read examples directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(examplesDir, entry.Name())
		makefilePath := filepath.Join(dir, "makefile")

		if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
			return fmt.Errorf("failed to find makefile at %s: %w", makefilePath, err)
		}

		dirs = append(dirs, dir)
	}

	return runParallelTasks(dirs, taskConfig{
		name:        fmt.Sprintf("make %s", target),
		description: fmt.Sprintf("Running 'make %s' in", target),
		cmd: func(dir string) *exec.Cmd {
			cmd := exec.Command("make", target)
			cmd.Dir = dir
			return cmd
		},
	})
}

// enableMCP runs './bin/<cli> mcp <platform> enable' for every binary in bin/
func enableMCP(platform string) error {
	binDir := "bin"

	if _, err := os.Stat(binDir); os.IsNotExist(err) {
		return fmt.Errorf("bin directory does not exist - run 'mage build' first")
	}

	entries, err := os.ReadDir(binDir)
	if err != nil {
		return fmt.Errorf("failed to read bin directory: %w", err)
	}

	var binaries []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Check if file is executable
		if info.Mode()&0o111 == 0 {
			continue
		}

		binaries = append(binaries, filepath.Join(binDir, entry.Name()))
	}

	err = runParallelTasks(binaries, taskConfig{
		name:        fmt.Sprintf("%s MCP enable", platform),
		description: fmt.Sprintf("Enabling %s MCP for", platform),
		cmd: func(binary string) *exec.Cmd {
			return exec.Command(binary, "mcp", platform, "enable")
		},
	})

	if err == nil {
		fmt.Printf("\n🎉 All binaries are now available in %s!\n", platform)
		fmt.Printf("⚠️  Please restart %s for changes to take effect.\n", platform)
	}

	return err
}
