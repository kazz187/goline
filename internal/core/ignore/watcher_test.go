package ignore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWatcherManualReload(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "goline-watcher-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test .golineignore file
	ignoreContent := []byte("*.secret\nprivate/")
	ignoreFilePath := filepath.Join(tempDir, ".golineignore")
	err = os.WriteFile(ignoreFilePath, ignoreContent, 0644)
	if err != nil {
		t.Fatalf("Failed to write .golineignore file: %v", err)
	}

	// Create the controller and initialize it
	controller := NewController(tempDir)
	err = controller.Initialize()
	if err != nil {
		t.Fatalf("Failed to initialize controller: %v", err)
	}

	// Verify initial state
	if controller.ValidateAccess("test.secret") {
		t.Errorf("Expected test.secret to be blocked initially, but it was allowed")
	}
	if !controller.ValidateAccess("test.txt") {
		t.Errorf("Expected test.txt to be allowed initially, but it was blocked")
	}

	// We're not testing the watcher directly, just the reload functionality

	// Modify the .golineignore file
	newIgnoreContent := []byte("*.txt\nprivate/")
	err = os.WriteFile(ignoreFilePath, newIgnoreContent, 0644)
	if err != nil {
		t.Fatalf("Failed to update .golineignore file: %v", err)
	}

	// Manually reload the controller
	err = controller.Reload()
	if err != nil {
		t.Fatalf("Failed to reload controller: %v", err)
	}

	// Verify the controller was reloaded with new patterns
	if !controller.ValidateAccess("test.secret") {
		t.Errorf("Expected test.secret to be allowed after update, but it was blocked")
	}
	if controller.ValidateAccess("test.txt") {
		t.Errorf("Expected test.txt to be blocked after update, but it was allowed")
	}

	// Delete the .golineignore file
	err = os.Remove(ignoreFilePath)
	if err != nil {
		t.Fatalf("Failed to delete .golineignore file: %v", err)
	}

	// Manually reload the controller
	err = controller.Reload()
	if err != nil {
		t.Fatalf("Failed to reload controller: %v", err)
	}

	// Verify the controller allows all files when .golineignore is deleted
	if !controller.ValidateAccess("test.secret") {
		t.Errorf("Expected test.secret to be allowed after deletion, but it was blocked")
	}
	if !controller.ValidateAccess("test.txt") {
		t.Errorf("Expected test.txt to be allowed after deletion, but it was blocked")
	}
}
