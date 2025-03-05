# Goline Ignore Controller

The Ignore Controller is a component that controls AI access to files by enforcing ignore patterns. It is similar to how `.gitignore` works, but specifically for the Goline AI assistant.

## Features

- Uses standard `.gitignore` syntax in `.golineignore` files
- Validates file access based on ignore patterns
- Validates terminal commands to prevent access to ignored files
- Filters arrays of paths, removing those that should be ignored
- Watches for changes to the `.golineignore` file and automatically reloads

## Usage

### Basic Usage

```go
import (
    "github.com/kazz187/goline/internal/core/ignore"
)

// Create a new controller for the current working directory
controller := ignore.NewController("/path/to/workspace")

// Initialize the controller
err := controller.Initialize()
if err != nil {
    log.Fatalf("Failed to initialize controller: %v", err)
}

// Check if a file should be accessible
if controller.ValidateAccess("config.secret") {
    // File is accessible
} else {
    // File is blocked
}

// Validate a terminal command
if result := controller.ValidateCommand("cat config.secret"); result != "" {
    fmt.Printf("Command blocked: accessing %s\n", result)
} else {
    fmt.Println("Command allowed")
}

// Filter an array of paths
paths := []string{"src/main.go", ".env", "README.md"}
allowedPaths := controller.FilterPaths(paths)
```

### Using the File Watcher

The file watcher monitors changes to the `.golineignore` file and automatically reloads the ignore patterns when the file is modified.

```go
// Create a new watcher for the controller
watcher := ignore.NewWatcher(controller, "/path/to/workspace")

// Start the watcher
watcher.Start()

// Stop the watcher when done
defer watcher.Stop()
```

## `.golineignore` File Format

The `.golineignore` file uses the same syntax as `.gitignore`. Here's an example:

```
# This is a comment
*.secret
private/
temp.*
**/.git/**
```

This will ignore:
- All files with the `.secret` extension
- All files in the `private/` directory and its subdirectories
- All files starting with `temp.`
- All files in any `.git` directory

## Implementation Details

The Ignore Controller uses the `github.com/sabhiram/go-gitignore` package to parse and match ignore patterns. It normalizes paths to ensure consistent matching regardless of whether absolute or relative paths are used.

The file watcher uses a polling approach to detect changes to the `.golineignore` file. It checks the file's modification time at regular intervals and reloads the ignore patterns when the file is modified.
