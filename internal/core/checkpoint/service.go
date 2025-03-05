package checkpoint

import (
	"fmt"
	"time"

	pb "github.com/kazz187/goline/proto/gen/go/goline/v1"
)

// Service provides checkpoint functionality for tasks
type Service struct {
	managers map[string]*Manager
}

// NewService creates a new checkpoint service
func NewService() *Service {
	return &Service{
		managers: make(map[string]*Manager),
	}
}

// GetManager returns a checkpoint manager for a task
func (s *Service) GetManager(taskID, workingDir string) (*Manager, error) {
	// Check if manager already exists
	if manager, ok := s.managers[taskID]; ok {
		return manager, nil
	}

	// Create new manager
	manager, err := NewManager(taskID, workingDir)
	if err != nil {
		return nil, err
	}

	// Initialize manager
	if err := manager.Initialize(); err != nil {
		return nil, err
	}

	// Store manager
	s.managers[taskID] = manager
	return manager, nil
}

// SaveCheckpoint saves a checkpoint for a task
func (s *Service) SaveCheckpoint(taskID, workingDir, name, description string) (*pb.CheckpointEvent, error) {
	// Get manager
	manager, err := s.GetManager(taskID, workingDir)
	if err != nil {
		return nil, err
	}

	// Create checkpoint
	checkpointID, err := manager.CreateCheckpoint(name, description)
	if err != nil {
		return nil, err
	}

	// Create checkpoint event
	checkpointEvent := &pb.CheckpointEvent{
		OperationType: pb.CheckpointOperationType_CHECKPOINT_OPERATION_TYPE_SAVE,
		CheckpointId:  checkpointID,
		Name:          name,
		Description:   description,
	}

	return checkpointEvent, nil
}

// RestoreCheckpoint restores a checkpoint for a task
func (s *Service) RestoreCheckpoint(taskID, workingDir, checkpointID string) (*pb.CheckpointEvent, error) {
	// Get manager
	manager, err := s.GetManager(taskID, workingDir)
	if err != nil {
		return nil, err
	}

	// Get checkpoint info
	checkpoints, err := manager.GetCheckpoints()
	if err != nil {
		return nil, err
	}

	// Find checkpoint
	var checkpoint *CheckpointInfo
	for _, cp := range checkpoints {
		if cp.ID == checkpointID {
			checkpoint = &cp
			break
		}
	}
	if checkpoint == nil {
		return nil, fmt.Errorf("checkpoint not found: %s", checkpointID)
	}

	// Restore checkpoint
	if err := manager.RestoreCheckpoint(checkpointID); err != nil {
		return nil, err
	}

	// Create checkpoint event
	checkpointEvent := &pb.CheckpointEvent{
		OperationType: pb.CheckpointOperationType_CHECKPOINT_OPERATION_TYPE_RESTORE,
		CheckpointId:  checkpointID,
		Name:          checkpoint.Name,
	}

	return checkpointEvent, nil
}

// GetCheckpoints returns all checkpoints for a task
func (s *Service) GetCheckpoints(taskID, workingDir string) ([]CheckpointInfo, error) {
	// Get manager
	manager, err := s.GetManager(taskID, workingDir)
	if err != nil {
		return nil, err
	}

	// Get checkpoints
	return manager.GetCheckpoints()
}

// GetDiff returns the diff between two checkpoints
func (s *Service) GetDiff(taskID, workingDir, fromCheckpointID, toCheckpointID string) ([]FileDiff, error) {
	// Get manager
	manager, err := s.GetManager(taskID, workingDir)
	if err != nil {
		return nil, err
	}

	// Get diff
	return manager.GetDiff(fromCheckpointID, toCheckpointID)
}

// FormatDiff formats a diff for display
func (s *Service) FormatDiff(diffs []FileDiff) string {
	if len(diffs) == 0 {
		return "No changes"
	}

	var result string
	for _, diff := range diffs {
		result += fmt.Sprintf("File: %s\n", diff.RelativePath)
		if diff.Before == "" && diff.After != "" {
			result += "  (New file)\n"
		} else if diff.Before != "" && diff.After == "" {
			result += "  (Deleted)\n"
		} else {
			result += "  (Modified)\n"
		}
		result += "\n"
	}

	return result
}

// FormatCheckpointList formats a list of checkpoints for display
func (s *Service) FormatCheckpointList(checkpoints []CheckpointInfo) string {
	if len(checkpoints) == 0 {
		return "No checkpoints"
	}

	var result string
	result += "Checkpoints:\n"
	for _, cp := range checkpoints {
		result += fmt.Sprintf("  %s: %s (%s)\n", cp.ID[:8], cp.Name, cp.Timestamp.Format(time.RFC3339))
	}

	return result
}
