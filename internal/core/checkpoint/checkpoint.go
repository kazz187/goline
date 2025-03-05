package checkpoint

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kazz187/goline/internal/core/ignore"
	pb "github.com/kazz187/goline/proto/gen/go/goline/v1"
)

// Manager handles checkpoint operations for a task
type Manager struct {
	taskID           string
	workingDir       string
	ignoreController *ignore.Controller
	shadowGitPath    string
}

// NewManager creates a new checkpoint manager for a task
func NewManager(taskID, workingDir string) (*Manager, error) {
	ignoreController := ignore.NewController(workingDir)
	if err := ignoreController.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize ignore controller: %w", err)
	}

	return &Manager{
		taskID:           taskID,
		workingDir:       workingDir,
		ignoreController: ignoreController,
	}, nil
}

// Initialize initializes the checkpoint manager
func (m *Manager) Initialize() error {
	// Check if git is installed
	if err := m.checkGitInstalled(); err != nil {
		return err
	}

	// Initialize shadow git repository
	gitPath, err := m.initShadowGit()
	if err != nil {
		return err
	}
	m.shadowGitPath = gitPath

	return nil
}

// checkGitInstalled checks if git is installed on the system
func (m *Manager) checkGitInstalled() error {
	cmd := exec.Command("git", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git must be installed to use checkpoints: %w", err)
	}
	return nil
}

// getShadowGitPath returns the path to the shadow git repository
func (m *Manager) getShadowGitPath() (string, error) {
	// Get the goline data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	// Use .goline/tasks/[taskID]/checkpoints/.git
	checkpointsDir := filepath.Join(homeDir, ".goline", "tasks", m.taskID, "checkpoints")
	if err := os.MkdirAll(checkpointsDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create checkpoints directory: %w", err)
	}

	gitPath := filepath.Join(checkpointsDir, ".git")
	return gitPath, nil
}

// initShadowGit initializes the shadow git repository
func (m *Manager) initShadowGit() (string, error) {
	gitPath, err := m.getShadowGitPath()
	if err != nil {
		return "", err
	}

	// Check if git repository already exists
	if _, err := os.Stat(gitPath); err == nil {
		// Verify worktree configuration
		worktree, err := m.getShadowGitConfigWorkTree()
		if err != nil {
			return "", err
		}
		if worktree != m.workingDir {
			return "", fmt.Errorf("checkpoints can only be used in the original workspace: %s", worktree)
		}
		return gitPath, nil
	}

	// Initialize new git repository
	checkpointsDir := filepath.Dir(gitPath)
	cmd := exec.Command("git", "init")
	cmd.Dir = checkpointsDir
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to initialize git repository: %w", err)
	}

	// Configure git repository
	configs := []struct {
		key, value string
	}{
		{"core.worktree", m.workingDir},
		{"commit.gpgSign", "false"},
		{"user.name", "Goline Checkpoint"},
		{"user.email", "checkpoint@goline.bot"},
		{"core.quotePath", "false"},
		{"core.precomposeunicode", "true"},
	}

	for _, config := range configs {
		cmd := exec.Command("git", "config", config.key, config.value)
		cmd.Dir = checkpointsDir
		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to configure git repository: %w", err)
		}
	}

	// Set up excludes
	if err := m.writeExcludesFile(gitPath); err != nil {
		return "", err
	}

	// Initial commit
	cmd = exec.Command("git", "commit", "--allow-empty", "-m", "initial commit")
	cmd.Dir = checkpointsDir
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create initial commit: %w", err)
	}

	return gitPath, nil
}

// getShadowGitConfigWorkTree returns the worktree path from the shadow git configuration
func (m *Manager) getShadowGitConfigWorkTree() (string, error) {
	gitPath, err := m.getShadowGitPath()
	if err != nil {
		return "", err
	}

	cmd := exec.Command("git", "config", "core.worktree")
	cmd.Dir = filepath.Dir(gitPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree configuration: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// writeExcludesFile writes the excludes file for the shadow git repository
func (m *Manager) writeExcludesFile(gitPath string) error {
	excludesDir := filepath.Join(gitPath, "info")
	if err := os.MkdirAll(excludesDir, 0755); err != nil {
		return fmt.Errorf("failed to create excludes directory: %w", err)
	}

	excludesPath := filepath.Join(excludesDir, "exclude")
	excludes := []string{
		// Git directories
		".git/",
		".git_disabled/",

		// Build and dependency directories
		"node_modules/",
		"__pycache__/",
		"env/",
		"venv/",
		"target/dependency/",
		"build/dependencies/",
		"dist/",
		"out/",
		"bundle/",
		"vendor/",
		"tmp/",
		"temp/",
		"deps/",
		"pkg/",
		"Pods/",

		// Media files
		"*.jpg",
		"*.jpeg",
		"*.png",
		"*.gif",
		"*.bmp",
		"*.ico",
		"*.mp3",
		"*.mp4",
		"*.wav",
		"*.avi",
		"*.mov",
		"*.wmv",
		"*.webm",
		"*.webp",
		"*.m4a",
		"*.flac",

		// Build and dependency directories
		"build/",
		"bin/",
		"obj/",
		".gradle/",
		".idea/",
		".vscode/",
		".vs/",
		"coverage/",
		".next/",
		".nuxt/",

		// Cache and temporary files
		"*.cache",
		"*.tmp",
		"*.temp",
		"*.swp",
		"*.swo",
		"*.pyc",
		"*.pyo",
		".pytest_cache/",
		".eslintcache",

		// Environment and config files
		".env*",
		"*.local",
		"*.development",
		"*.production",

		// Large data files
		"*.zip",
		"*.tar",
		"*.gz",
		"*.rar",
		"*.7z",
		"*.iso",
		"*.bin",
		"*.exe",
		"*.dll",
		"*.so",
		"*.dylib",

		// Database files
		"*.sqlite",
		"*.db",
		"*.sql",

		// Log files
		"*.logs",
		"*.error",
		"npm-debug.log*",
		"yarn-debug.log*",
		"yarn-error.log*",

		// System files
		".DS_Store",
	}

	// Add LFS patterns from .gitattributes if it exists
	lfsPatterns, err := m.getLFSPatterns()
	if err != nil {
		return err
	}
	excludes = append(excludes, lfsPatterns...)

	return os.WriteFile(excludesPath, []byte(strings.Join(excludes, "\n")), 0644)
}

// getLFSPatterns returns LFS patterns from .gitattributes
func (m *Manager) getLFSPatterns() ([]string, error) {
	attributesPath := filepath.Join(m.workingDir, ".gitattributes")
	content, err := os.ReadFile(attributesPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var patterns []string
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if strings.Contains(line, "filter=lfs") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				patterns = append(patterns, parts[0])
			}
		}
	}

	return patterns, nil
}

// renameNestedGitRepos renames nested .git directories to avoid conflicts
func (m *Manager) renameNestedGitRepos(disable bool) error {
	suffix := "_disabled"

	var gitPaths []string
	err := filepath.WalkDir(m.workingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && (d.Name() == ".git" || d.Name() == ".git"+suffix) {
			// Skip the root .git directory
			if filepath.Dir(path) == m.workingDir {
				return filepath.SkipDir
			}
			gitPaths = append(gitPaths, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	for _, gitPath := range gitPaths {
		var newPath string
		if disable {
			newPath = gitPath + suffix
		} else {
			newPath = strings.TrimSuffix(gitPath, suffix)
		}

		if err := os.Rename(gitPath, newPath); err != nil {
			return err
		}
	}

	return nil
}

// addAllFiles adds all files to the shadow git repository
func (m *Manager) addAllFiles() error {
	// Disable nested git repositories
	if err := m.renameNestedGitRepos(true); err != nil {
		return err
	}
	defer m.renameNestedGitRepos(false)

	// Add all files
	cmd := exec.Command("git", "add", ".")
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add files to git: %w", err)
	}

	return nil
}

// CreateCheckpoint creates a new checkpoint
func (m *Manager) CreateCheckpoint(name, description string) (string, error) {
	// Add all files to git
	if err := m.addAllFiles(); err != nil {
		return "", err
	}

	// Create commit
	cmd := exec.Command("git", "commit", "--allow-empty", "-m", fmt.Sprintf("checkpoint: %s", name))
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to create checkpoint: %w", err)
	}

	// Get commit hash
	cmd = exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit hash: %w", err)
	}

	commitHash := strings.TrimSpace(string(output))
	return commitHash, nil
}

// RestoreCheckpoint restores a checkpoint
func (m *Manager) RestoreCheckpoint(commitHash string) error {
	// Clean working directory and force reset
	cmd := exec.Command("git", "clean", "-f", "-d")
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to clean working directory: %w", err)
	}

	// Reset to commit
	cmd = exec.Command("git", "reset", "--hard", commitHash)
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to reset to checkpoint: %w", err)
	}

	return nil
}

// GetDiff returns the diff between two checkpoints
func (m *Manager) GetDiff(fromHash, toHash string) ([]FileDiff, error) {
	// If toHash is empty, compare to working directory
	diffArg := fmt.Sprintf("%s..%s", fromHash, toHash)
	if toHash == "" {
		diffArg = fromHash
	}

	// Get diff summary
	cmd := exec.Command("git", "diff", "--name-only", diffArg)
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get diff summary: %w", err)
	}

	changedFiles := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(changedFiles) == 1 && changedFiles[0] == "" {
		return nil, nil
	}

	var diffs []FileDiff
	for _, file := range changedFiles {
		// Skip empty lines
		if file == "" {
			continue
		}

		// Get file content before
		var beforeContent string
		beforeCmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", fromHash, file))
		beforeCmd.Dir = filepath.Dir(m.shadowGitPath)
		beforeOutput, err := beforeCmd.Output()
		if err != nil {
			// File didn't exist in older commit
			beforeContent = ""
		} else {
			beforeContent = string(beforeOutput)
		}

		// Get file content after
		var afterContent string
		if toHash == "" {
			// Read from disk
			afterPath := filepath.Join(m.workingDir, file)
			afterBytes, err := os.ReadFile(afterPath)
			if err != nil {
				// File might be deleted
				afterContent = ""
			} else {
				afterContent = string(afterBytes)
			}
		} else {
			// Get from git
			afterCmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", toHash, file))
			afterCmd.Dir = filepath.Dir(m.shadowGitPath)
			afterOutput, err := afterCmd.Output()
			if err != nil {
				// File didn't exist in newer commit
				afterContent = ""
			} else {
				afterContent = string(afterOutput)
			}
		}

		// Create diff
		diff := FileDiff{
			RelativePath: file,
			AbsolutePath: filepath.Join(m.workingDir, file),
			Before:       beforeContent,
			After:        afterContent,
		}
		diffs = append(diffs, diff)
	}

	return diffs, nil
}

// GetCheckpoints returns all checkpoints for the task
func (m *Manager) GetCheckpoints() ([]CheckpointInfo, error) {
	// Get all commits
	cmd := exec.Command("git", "log", "--pretty=format:%H|%s|%at")
	cmd.Dir = filepath.Dir(m.shadowGitPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commits: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var checkpoints []CheckpointInfo
	for _, line := range lines {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) != 3 {
			continue
		}

		hash := parts[0]
		message := parts[1]
		timestamp, err := strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			continue
		}

		// Skip initial commit
		if message == "initial commit" {
			continue
		}

		// Extract name from message
		name := message
		if strings.HasPrefix(message, "checkpoint: ") {
			name = strings.TrimPrefix(message, "checkpoint: ")
		}

		checkpoint := CheckpointInfo{
			ID:        hash,
			Name:      name,
			Timestamp: time.Unix(timestamp, 0),
		}
		checkpoints = append(checkpoints, checkpoint)
	}

	return checkpoints, nil
}

// CreateCheckpointProto creates a checkpoint proto message
func (m *Manager) CreateCheckpointProto(id, name, description string) (*pb.Checkpoint, error) {
	// Get file snapshots
	var fileSnapshots []*pb.FileSnapshot
	err := filepath.WalkDir(m.workingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		// Skip files that should be ignored
		relPath, err := filepath.Rel(m.workingDir, path)
		if err != nil {
			return err
		}
		if !m.ignoreController.ValidateAccess(relPath) {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		// Calculate content hash
		hash := sha256.Sum256(content)
		contentHash := hex.EncodeToString(hash[:])

		fileSnapshot := &pb.FileSnapshot{
			FilePath:    relPath,
			Content:     string(content),
			ContentHash: contentHash,
		}
		fileSnapshots = append(fileSnapshots, fileSnapshot)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Get git status
	gitStatus, err := m.getGitStatus()
	if err != nil {
		// Git status is optional, so we can continue without it
		gitStatus = nil
	}

	// Create checkpoint proto
	checkpoint := &pb.Checkpoint{
		Id:          id,
		Name:        name,
		Description: description,
		Timestamp:   time.Now().Format(time.RFC3339),
		Files:       fileSnapshots,
		GitStatus:   gitStatus,
	}

	return checkpoint, nil
}

// getGitStatus returns the git status of the workspace
func (m *Manager) getGitStatus() (*pb.GitStatus, error) {
	// Check if workspace is a git repository
	gitDir := filepath.Join(m.workingDir, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil, nil
	}

	// Get current branch
	branchCmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	branchCmd.Dir = m.workingDir
	branchOutput, err := branchCmd.Output()
	if err != nil {
		return nil, err
	}
	branch := strings.TrimSpace(string(branchOutput))

	// Get current commit hash
	hashCmd := exec.Command("git", "rev-parse", "HEAD")
	hashCmd.Dir = m.workingDir
	hashOutput, err := hashCmd.Output()
	if err != nil {
		return nil, err
	}
	commitHash := strings.TrimSpace(string(hashOutput))

	// Check if there are uncommitted changes
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = m.workingDir
	statusOutput, err := statusCmd.Output()
	if err != nil {
		return nil, err
	}
	hasUncommittedChanges := len(statusOutput) > 0

	// Get modified files
	var modifiedFiles []string
	if hasUncommittedChanges {
		lines := strings.Split(strings.TrimSpace(string(statusOutput)), "\n")
		for _, line := range lines {
			if line == "" {
				continue
			}
			// Extract file path from status line
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				modifiedFiles = append(modifiedFiles, parts[1])
			}
		}
	}

	gitStatus := &pb.GitStatus{
		Branch:                branch,
		CommitHash:            commitHash,
		HasUncommittedChanges: hasUncommittedChanges,
		ModifiedFiles:         modifiedFiles,
	}

	return gitStatus, nil
}

// FileDiff represents a diff between two versions of a file
type FileDiff struct {
	RelativePath string
	AbsolutePath string
	Before       string
	After        string
}

// CheckpointInfo represents information about a checkpoint
type CheckpointInfo struct {
	ID        string
	Name      string
	Timestamp time.Time
}
