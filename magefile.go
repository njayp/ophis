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

// runMakeInExamples runs a make target in all example directories in parallel
func runMakeInExamples(target string) error {
	// Find all example directories
	examplesDir := "examples"
	entries, err := os.ReadDir(examplesDir)
	if err != nil {
		return fmt.Errorf("failed to read examples directory: %w", err)
	}

	type result struct {
		dir string
		err error
	}

	var wg sync.WaitGroup
	resultChan := make(chan result, len(entries))

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
				fmt.Printf("✗ Failed: %s\n", dir)
				resultChan <- result{dir: dir, err: err}
			} else {
				fmt.Printf("✓ Success: %s\n", dir)
				resultChan <- result{dir: dir, err: nil}
			}
		}(dir)
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(resultChan)

	// Collect all results and errors
	var failures []string
	var successCount int
	for res := range resultChan {
		if res.err != nil {
			failures = append(failures, fmt.Sprintf("  - %s: %v", res.dir, res.err))
		} else {
			successCount++
		}
	}

	// Report final status
	if len(failures) > 0 {
		fmt.Printf("\n✗ 'make %s' failed in %d example(s):\n", target, len(failures))
		fmt.Println(strings.Join(failures, "\n"))
		fmt.Printf("\nSummary: %d succeeded, %d failed\n", successCount, len(failures))
		return fmt.Errorf("'make %s' failed in %d example(s)", target, len(failures))
	}

	fmt.Printf("\n✓ Successfully ran 'make %s' in all %d examples\n", target, successCount)
	return nil
}
