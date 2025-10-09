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

// Vscode runs './bin/<cli> mcp cursor enable' for every binary in bin/
func Cursor() error {
	return enableMCP("cursor")
}

// runMakeInExamples runs a make target in all example directories in parallel
func runMakeInExamples(target string) error {
	dirs, err := findExampleDirs()
	if err != nil {
		return err
	}

	return runParallel(dirs, func(dir string) error {
		cmd := exec.Command("make", target)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	})
}

// enableMCP runs './bin/<cli> mcp <platform> enable' for every binary in bin/
func enableMCP(platform string) error {
	binaries, err := findExecutables("bin")
	if err != nil {
		return err
	}

	if err := runParallel(binaries, func(binary string) error {
		cmd := exec.Command(binary, "mcp", platform, "enable")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}); err != nil {
		return err
	}

	fmt.Printf("\n🎉 All binaries are now available in %s!\n", platform)
	fmt.Printf("⚠️  Please restart %s for changes to take effect.\n", platform)
	return nil
}

// findExampleDirs returns all example directories with makefiles
func findExampleDirs() ([]string, error) {
	entries, err := os.ReadDir("examples")
	if err != nil {
		return nil, fmt.Errorf("failed to read examples directory: %w", err)
	}

	var dirs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		dir := filepath.Join("examples", entry.Name())
		if _, err := os.Stat(filepath.Join(dir, "makefile")); err == nil {
			dirs = append(dirs, dir)
		}
	}
	return dirs, nil
}

// findExecutables returns all executable files in a directory
func findExecutables(dir string) ([]string, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%s directory does not exist - run 'mage build' first", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s directory: %w", dir, err)
	}

	var executables []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.Mode()&0o111 != 0 {
			executables = append(executables, filepath.Join(dir, entry.Name()))
		}
	}
	return executables, nil
}

// runParallel runs a function on multiple items in parallel
func runParallel(items []string, fn func(string) error) error {
	if len(items) == 0 {
		return fmt.Errorf("no items to process")
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(items))

	for _, item := range items {
		wg.Add(1)
		go func(item string) {
			defer wg.Done()
			if err := fn(item); err != nil {
				errChan <- fmt.Errorf("%s: %w", item, err)
			}
		}(item)
	}

	wg.Wait()
	close(errChan)

	var errors []string
	for err := range errChan {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}
