package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreController(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "goline-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test .golineignore file
	ignoreContent := []byte(`.env
*.secret
private/
# This is a comment

temp.*
file-with-space-at-end.* 
**/.git/**`)
	err = os.WriteFile(filepath.Join(tempDir, ".golineignore"), ignoreContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write .golineignore file: %v", err)
	}

	// Create the controller
	controller := NewController(tempDir)
	err = controller.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize controller: %v", err)
	}

	// Test default patterns
	t.Run("DefaultPatterns", func(t *testing.T) {
		// Test allowed files
		allowedFiles := []string{
			"src/index.go",
			"README.md",
			"go.mod",
		}
		for _, file := range allowedFiles {
			if !controller.ValidateAccess(file) {
				t.Errorf("Expected file %s to be allowed, but it was blocked", file)
			}
		}

		// Test .golineignore file is blocked
		if controller.ValidateAccess(".golineignore") {
			t.Errorf("Expected .golineignore to be blocked, but it was allowed")
		}
	})

	// Test custom patterns
	t.Run("CustomPatterns", func(t *testing.T) {
		// Test blocked files
		blockedFiles := []string{
			"config.secret",
			"private/data.txt",
			"temp.json",
			"nested/deep/file.secret",
			"private/nested/deep/file.txt",
		}
		for _, file := range blockedFiles {
			if controller.ValidateAccess(file) {
				t.Errorf("Expected file %s to be blocked, but it was allowed", file)
			}
		}

		// Test allowed files
		allowedFiles := []string{
			"public/data.txt",
			"config.json",
			"src/temp/file.go",
			"nested/deep/file.txt",
			"not-private/data.txt",
		}
		for _, file := range allowedFiles {
			if !controller.ValidateAccess(file) {
				t.Errorf("Expected file %s to be allowed, but it was blocked", file)
			}
		}
	})

	// Test path handling
	t.Run("PathHandling", func(t *testing.T) {
		// Test absolute paths
		allowedPath := filepath.Join(tempDir, "src/file.go")
		if !controller.ValidateAccess(allowedPath) {
			t.Errorf("Expected absolute path %s to be allowed, but it was blocked", allowedPath)
		}

		ignoredPath := filepath.Join(tempDir, "config.secret")
		if controller.ValidateAccess(ignoredPath) {
			t.Errorf("Expected absolute path %s to be blocked, but it was allowed", ignoredPath)
		}

		// Test relative paths
		if !controller.ValidateAccess("./src/file.go") {
			t.Errorf("Expected relative path ./src/file.go to be allowed, but it was blocked")
		}

		if controller.ValidateAccess("./config.secret") {
			t.Errorf("Expected relative path ./config.secret to be blocked, but it was allowed")
		}
	})

	// Test batch filtering
	t.Run("BatchFiltering", func(t *testing.T) {
		paths := []string{"src/index.go", ".env", "lib/utils.go", ".git/config", "dist/bundle.js"}
		expected := []string{"src/index.go", "lib/utils.go", "dist/bundle.js"}

		filtered := controller.FilterPaths(paths)
		if len(filtered) != len(expected) {
			t.Errorf("Expected %d paths, got %d", len(expected), len(filtered))
		}

		// Check each path is in the expected list
		for i, path := range filtered {
			if path != expected[i] {
				t.Errorf("Expected path %s at index %d, got %s", expected[i], i, path)
			}
		}
	})

	// Test command validation
	t.Run("CommandValidation", func(t *testing.T) {
		// Test allowed commands
		allowedCommands := []string{
			"ls -la",
			"cat README.md",
			"grep pattern src/main.go",
			"head -n 10 go.mod",
		}
		for _, cmd := range allowedCommands {
			if result := controller.ValidateCommand(cmd); result != "" {
				t.Errorf("Expected command %s to be allowed, but it was blocked due to %s", cmd, result)
			}
		}

		// Test blocked commands
		blockedCommands := []string{
			"cat .env",
			"grep pattern config.secret",
			"head -n 10 private/data.txt",
		}
		for _, cmd := range blockedCommands {
			if result := controller.ValidateCommand(cmd); result == "" {
				t.Errorf("Expected command %s to be blocked, but it was allowed", cmd)
			}
		}
	})

	// Test error handling
	t.Run("ErrorHandling", func(t *testing.T) {
		// Test with missing .golineignore
		emptyDir, err := os.MkdirTemp("", "goline-empty-*")
		if err != nil {
			t.Fatalf("Failed to create empty temp directory: %v", err)
		}
		defer os.RemoveAll(emptyDir)

		emptyController := NewController(emptyDir)
		err = emptyController.Initialize()
		if err != nil {
			t.Fatalf("Failed to initialize controller with missing .golineignore: %v", err)
		}

		if !emptyController.ValidateAccess("file.txt") {
			t.Errorf("Expected file to be allowed when .golineignore is missing, but it was blocked")
		}

		// Test with empty .golineignore
		err = os.WriteFile(filepath.Join(tempDir, ".golineignore"), []byte{}, 0644)
		if err != nil {
			t.Fatalf("Failed to write empty .golineignore file: %v", err)
		}

		err = controller.Reload()
		if err != nil {
			t.Fatalf("Failed to reload controller with empty .golineignore: %v", err)
		}

		if !controller.ValidateAccess("regular-file.txt") {
			t.Errorf("Expected file to be allowed with empty .golineignore, but it was blocked")
		}
	})
}
