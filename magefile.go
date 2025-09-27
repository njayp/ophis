//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

// runMakeInExamples runs a make target in all example directories in parallel
func runMakeInExamples(target string) error {
	// Find all example directories
	examplesDir := "examples"
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		return fmt.Errorf("failed to read examples directory: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(entries))

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		dir := filepath.Join(examplesDir, entry.Name())
		makefilePath := filepath.Join(dir, "Makefile")

		// Check if Makefile exists
		if _, err := os.Stat(makefilePath); os.IsNotExist(err) {
			continue
		}

		wg.Add(1)
		go func(dir string) {
			defer wg.Done()

			fmt.Printf("Running 'make %s' in %s\n", target, dir)
			cmd := exec.Command("make", target)
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				errChan <- fmt.Errorf("%s: %w", dir, err)
				fmt.Printf("✗ Failed: %s\n", dir)
			} else {
				fmt.Printf("✓ Success: %s\n", dir)
			}
		}(dir)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(errChan)

	// Collect any errors
	var failed bool
	for err := range errChan {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			failed = true
		}
	}

	if failed {
		return fmt.Errorf("'make %s' failed in some examples", target)
	}

	fmt.Printf("\n✓ Successfully ran 'make %s' in all examples\n", target)
	return nil
}

// All runs 'make all' for each example directory in parallel
func All() error {
	return runMakeInExamples("all")
}

// Build runs 'make build' for each example directory in parallel
func Build() error {
	return runMakeInExamples("build")
}

// Test runs 'make test' for the main project
func Test() error {
	cmd := exec.Command("make", "test")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Lint runs 'make lint' for the main project
func Lint() error {
	cmd := exec.Command("make", "lint")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
