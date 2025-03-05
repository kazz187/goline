package ignore

import (
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// LockTextSymbol represents a lock emoji used to indicate locked/ignored files
const LockTextSymbol = "🔒"

// Controller controls AI access to files by enforcing ignore patterns.
// Uses the 'go-gitignore' library to support standard .gitignore syntax in .golineignore files.
type Controller struct {
	cwd                 string
	ignoreInstance      *ignore.GitIgnore
	golineIgnoreContent string
}

// NewController creates a new ignore controller for the given working directory
func NewController(cwd string) *Controller {
	return &Controller{
		cwd:                 cwd,
		ignoreInstance:      nil,
		golineIgnoreContent: "",
	}
}

// Initialize initializes the controller by loading custom patterns
// Must be called after construction and before using the controller
func (c *Controller) Initialize() error {
	return c.loadGolineIgnore()
}

// loadGolineIgnore loads custom patterns from .golineignore if it exists
func (c *Controller) loadGolineIgnore() error {
	ignorePath := filepath.Join(c.cwd, ".golineignore")

	// Check if .golineignore exists
	content, err := os.ReadFile(ignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, that's fine
			c.golineIgnoreContent = ""
			c.ignoreInstance = nil
			return nil
		}
		// Other error reading file
		return err
	}

	// File exists, parse it
	c.golineIgnoreContent = string(content)

	// Add .golineignore to the patterns
	contentWithSelf := c.golineIgnoreContent
	if !strings.Contains(contentWithSelf, ".golineignore") {
		contentWithSelf += "\n.golineignore"
	}

	// Create ignore instance
	ignoreInstance := ignore.CompileIgnoreLines(strings.Split(contentWithSelf, "\n")...)

	c.ignoreInstance = ignoreInstance
	return nil
}

// ValidateAccess checks if a file should be accessible to the AI
// filePath can be absolute or relative to cwd
func (c *Controller) ValidateAccess(filePath string) bool {
	// Always allow access if .golineignore does not exist
	if c.ignoreInstance == nil {
		return true
	}

	// Normalize path to be relative to cwd
	absolutePath := filePath
	if !filepath.IsAbs(filePath) {
		absolutePath = filepath.Join(c.cwd, filePath)
	}

	relativePath, err := filepath.Rel(c.cwd, absolutePath)
	if err != nil {
		// Path is outside cwd, allow access
		return true
	}

	// Convert to forward slashes for consistency
	relativePath = filepath.ToSlash(relativePath)

	// Check if the file is ignored
	return !c.ignoreInstance.MatchesPath(relativePath)
}

// ValidateCommand checks if a terminal command should be allowed to execute based on file access patterns
// Returns path of file that is being accessed if it is being accessed, nil if command is allowed
func (c *Controller) ValidateCommand(command string) string {
	// Always allow if no .golineignore exists
	if c.ignoreInstance == nil {
		return ""
	}

	// Split command into parts and get the base command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return ""
	}

	baseCommand := strings.ToLower(parts[0])

	// Commands that read file contents
	fileReadingCommands := map[string]bool{
		// Unix commands
		"cat":  true,
		"less": true,
		"more": true,
		"head": true,
		"tail": true,
		"grep": true,
		"awk":  true,
		"sed":  true,
		// PowerShell commands and aliases
		"get-content":   true,
		"gc":            true,
		"type":          true,
		"select-string": true,
		"sls":           true,
	}

	if _, ok := fileReadingCommands[baseCommand]; ok {
		// Check each argument that could be a file path
		for i := 1; i < len(parts); i++ {
			arg := parts[i]
			// Skip command flags/options (both Unix and PowerShell style)
			if strings.HasPrefix(arg, "-") || strings.HasPrefix(arg, "/") {
				continue
			}
			// Ignore PowerShell parameter names
			if strings.Contains(arg, ":") {
				continue
			}
			// Validate file access
			if !c.ValidateAccess(arg) {
				return arg
			}
		}
	}

	return ""
}

// FilterPaths filters an array of paths, removing those that should be ignored
func (c *Controller) FilterPaths(paths []string) []string {
	var allowedPaths []string

	for _, p := range paths {
		if c.ValidateAccess(p) {
			allowedPaths = append(allowedPaths, p)
		}
	}

	return allowedPaths
}

// Reload reloads the ignore patterns from the .golineignore file
func (c *Controller) Reload() error {
	return c.loadGolineIgnore()
}
