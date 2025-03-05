package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kazz187/goline/internal/core/ignore"
)

func main() {
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current working directory: %v", err)
	}

	// Create a .golineignore file for demonstration
	ignoreFilePath := filepath.Join(cwd, ".golineignore")
	ignoreContent := []byte(`# This is a comment
*.secret
private/
temp.*
**/.git/**
`)
	err = os.WriteFile(ignoreFilePath, ignoreContent, 0644)
	if err != nil {
		log.Fatalf("Failed to write .golineignore file: %v", err)
	}
	defer os.Remove(ignoreFilePath) // Clean up after demo

	// Create the ignore controller
	controller := ignore.NewController(cwd)
	err = controller.Initialize()
	if err != nil {
		log.Fatalf("Failed to initialize controller: %v", err)
	}

	// Create and start the watcher
	watcher := ignore.NewWatcher(controller, cwd)
	watcher.Start()
	defer watcher.Stop()

	// Test some file paths
	testPaths := []string{
		"README.md",
		"config.secret",
		"private/data.txt",
		"public/data.txt",
		"temp.json",
		".git/config",
	}

	fmt.Println("Initial ignore patterns:")
	for _, path := range testPaths {
		allowed := controller.ValidateAccess(path)
		fmt.Printf("  %s: %s\n", path, formatAccess(allowed))
	}

	// Test command validation
	testCommands := []string{
		"ls -la",
		"cat README.md",
		"cat config.secret",
		"grep pattern private/data.txt",
	}

	fmt.Println("\nCommand validation:")
	for _, cmd := range testCommands {
		result := controller.ValidateCommand(cmd)
		if result == "" {
			fmt.Printf("  %s: Allowed\n", cmd)
		} else {
			fmt.Printf("  %s: Blocked (accessing %s)\n", cmd, result)
		}
	}

	// Demonstrate batch filtering
	batchPaths := []string{
		"src/main.go",
		"config.secret",
		"README.md",
		"private/data.txt",
		"public/data.txt",
	}
	fmt.Println("\nBatch filtering:")
	fmt.Printf("  Original paths: %v\n", batchPaths)
	fmt.Printf("  Filtered paths: %v\n", controller.FilterPaths(batchPaths))

	// Demonstrate file watcher by modifying .golineignore
	fmt.Println("\nModifying .golineignore file...")
	newIgnoreContent := []byte(`# Updated ignore patterns
*.md
private/
`)
	err = os.WriteFile(ignoreFilePath, newIgnoreContent, 0644)
	if err != nil {
		log.Fatalf("Failed to update .golineignore file: %v", err)
	}

	// Wait for the watcher to detect the change
	time.Sleep(3 * time.Second)

	fmt.Println("Updated ignore patterns:")
	for _, path := range testPaths {
		allowed := controller.ValidateAccess(path)
		fmt.Printf("  %s: %s\n", path, formatAccess(allowed))
	}
}

func formatAccess(allowed bool) string {
	if allowed {
		return "Allowed"
	}
	return fmt.Sprintf("Blocked %s", ignore.LockTextSymbol)
}
