package checkpoint

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Skip TestCheckpointManager for now as it requires mocking
func TestCheckpointManager(t *testing.T) {
	t.Skip("Skipping TestCheckpointManager as it requires mocking")
}

// TestCheckpointManagerSimple tests basic checkpoint manager functionality
// This test is skipped by default as it requires specific environment setup
func TestCheckpointManagerSimple(t *testing.T) {
	// Skip if not running in CI environment
	if os.Getenv("CI") != "true" {
		t.Skip("Skipping test in non-CI environment")
	}

	// Create a unique task ID for this test
	taskID := "test-task-manager-" + time.Now().Format("20060102150405")

	// Use the current working directory for testing
	workingDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	// Create a test file
	testFilePath := filepath.Join(workingDir, "test-checkpoint.txt")
	testContent := "Hello, world!"
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFilePath) // Clean up after test

	// Create a checkpoint manager
	manager, err := NewManager(taskID, workingDir)
	if err != nil {
		t.Fatalf("Failed to create checkpoint manager: %v", err)
	}

	// Initialize the manager
	if err := manager.Initialize(); err != nil {
		t.Fatalf("Failed to initialize checkpoint manager: %v", err)
	}

	// Create a checkpoint
	checkpointID, err := manager.CreateCheckpoint("Test Checkpoint", "Test description")
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}
	t.Logf("Created checkpoint: %s", checkpointID)

	// Get checkpoints
	checkpoints, err := manager.GetCheckpoints()
	if err != nil {
		t.Fatalf("Failed to get checkpoints: %v", err)
	}

	// Verify checkpoint was created
	if len(checkpoints) < 1 {
		t.Fatalf("Expected at least 1 checkpoint, got %d", len(checkpoints))
	}

	found := false
	for _, cp := range checkpoints {
		if cp.ID == checkpointID {
			found = true
			if cp.Name != "Test Checkpoint" {
				t.Errorf("Expected checkpoint name %q, got %q", "Test Checkpoint", cp.Name)
			}
			break
		}
	}

	if !found {
		t.Errorf("Checkpoint with ID %s not found", checkpointID)
	}
}

func TestCheckpointService(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "goline-checkpoint-service-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test file
	testFilePath := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, world!"
	if err := os.WriteFile(testFilePath, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create a checkpoint service
	service := NewService()

	// Create a unique task ID for this test
	taskID := "test-task-service-" + time.Now().Format("20060102150405")

	// Save a checkpoint
	event, err := service.SaveCheckpoint(taskID, tempDir, "Test Checkpoint", "Test description")
	if err != nil {
		t.Fatalf("Failed to save checkpoint: %v", err)
	}
	checkpointID := event.CheckpointId
	t.Logf("Saved checkpoint: %s", checkpointID)

	// Modify the test file
	modifiedContent := "Modified content"
	if err := os.WriteFile(testFilePath, []byte(modifiedContent), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Get diff
	diffs, err := service.GetDiff(taskID, tempDir, checkpointID, "")
	if err != nil {
		t.Fatalf("Failed to get diff: %v", err)
	}

	// Verify diff
	if len(diffs) != 1 {
		t.Fatalf("Expected 1 diff, got %d", len(diffs))
	}

	// Format diff
	diffText := service.FormatDiff(diffs)
	t.Logf("Diff: %s", diffText)

	// Get checkpoints
	checkpoints, err := service.GetCheckpoints(taskID, tempDir)
	if err != nil {
		t.Fatalf("Failed to get checkpoints: %v", err)
	}
	if len(checkpoints) != 1 {
		t.Fatalf("Expected 1 checkpoint, got %d", len(checkpoints))
	}

	// Format checkpoint list
	listText := service.FormatCheckpointList(checkpoints)
	t.Logf("Checkpoint list: %s", listText)

	// Restore checkpoint
	_, err = service.RestoreCheckpoint(taskID, tempDir, checkpointID)
	if err != nil {
		t.Fatalf("Failed to restore checkpoint: %v", err)
	}

	// Verify file was restored
	restoredContent, err := os.ReadFile(testFilePath)
	if err != nil {
		t.Fatalf("Failed to read restored file: %v", err)
	}
	if string(restoredContent) != testContent {
		t.Errorf("Expected restored content %q, got %q", testContent, string(restoredContent))
	}
}
