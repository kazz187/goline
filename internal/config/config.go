package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Provider represents an AI provider configuration
type Provider struct {
	APIKey    string `yaml:"api_key"`
	Endpoint  string `yaml:"endpoint,omitempty"`
	ModelName string `yaml:"model_name,omitempty"`
}

// Config represents the Goline configuration
type Config struct {
	// Providers is a map of provider name to provider configuration
	Providers map[string]Provider `yaml:"providers"`
	// DefaultProvider is the name of the default provider to use
	DefaultProvider string `yaml:"default_provider,omitempty"`
	// TasksDir is the directory where tasks are stored
	TasksDir string `yaml:"tasks_dir,omitempty"`
}

// RepoConfig represents repository-specific configuration
type RepoConfig struct {
	// Provider is the name of the provider to use for this repository
	Provider string `yaml:"provider,omitempty"`
	// ModelName is the name of the model to use for this repository
	ModelName string `yaml:"model_name,omitempty"`
	// TasksDir is the directory where tasks are stored for this repository
	TasksDir string `yaml:"tasks_dir,omitempty"`
}

// Manager handles configuration file operations
type Manager struct {
	globalConfig *Config
	repoConfig   *RepoConfig
	globalPath   string
	repoPath     string
}

// NewManager creates a new configuration manager
func NewManager() (*Manager, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	globalPath := filepath.Join(homeDir, ".goline", "config.yaml")

	// Find repository root (where .git directory exists)
	repoRoot, err := findRepoRoot()
	if err != nil {
		// Not in a git repository, use current directory
		repoRoot, err = os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	repoPath := filepath.Join(repoRoot, ".goline", "config.yaml")

	return &Manager{
		globalPath: globalPath,
		repoPath:   repoPath,
	}, nil
}

// findRepoRoot finds the repository root by looking for a .git directory
func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("not in a git repository")
		}
		dir = parent
	}
}

// Load loads both global and repository-specific configurations
func (m *Manager) Load() error {
	// Load global config
	globalConfig, err := m.loadGlobalConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load global config: %w", err)
	}
	m.globalConfig = globalConfig

	// Load repo config if it exists
	repoConfig, err := m.loadRepoConfig()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("failed to load repo config: %w", err)
	}
	m.repoConfig = repoConfig

	return nil
}

// loadGlobalConfig loads the global configuration file
func (m *Manager) loadGlobalConfig() (*Config, error) {
	data, err := os.ReadFile(m.globalPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return default config if file doesn't exist
			return &Config{
				Providers: make(map[string]Provider),
			}, os.ErrNotExist
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse global config: %w", err)
	}

	// Initialize providers map if nil
	if config.Providers == nil {
		config.Providers = make(map[string]Provider)
	}

	return &config, nil
}

// loadRepoConfig loads the repository-specific configuration file
func (m *Manager) loadRepoConfig() (*RepoConfig, error) {
	data, err := os.ReadFile(m.repoPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Return default config if file doesn't exist
			return &RepoConfig{}, os.ErrNotExist
		}
		return nil, err
	}

	var config RepoConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse repo config: %w", err)
	}

	return &config, nil
}

// SaveGlobalConfig saves the global configuration
func (m *Manager) SaveGlobalConfig() error {
	if m.globalConfig == nil {
		return errors.New("global config not loaded")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(m.globalPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(m.globalConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal global config: %w", err)
	}

	if err := os.WriteFile(m.globalPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write global config: %w", err)
	}

	return nil
}

// SaveRepoConfig saves the repository-specific configuration
func (m *Manager) SaveRepoConfig() error {
	if m.repoConfig == nil {
		return errors.New("repo config not loaded")
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(m.repoPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(m.repoConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal repo config: %w", err)
	}

	if err := os.WriteFile(m.repoPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write repo config: %w", err)
	}

	return nil
}

// GetGlobalConfig returns the global configuration
func (m *Manager) GetGlobalConfig() *Config {
	return m.globalConfig
}

// GetRepoConfig returns the repository-specific configuration
func (m *Manager) GetRepoConfig() *RepoConfig {
	return m.repoConfig
}

// SetProvider sets a provider configuration in the global config
func (m *Manager) SetProvider(name string, provider Provider) {
	if m.globalConfig == nil {
		m.globalConfig = &Config{
			Providers: make(map[string]Provider),
		}
	}
	m.globalConfig.Providers[name] = provider
}

// GetProvider returns a provider configuration from the global config
func (m *Manager) GetProvider(name string) (Provider, bool) {
	if m.globalConfig == nil || m.globalConfig.Providers == nil {
		return Provider{}, false
	}
	provider, ok := m.globalConfig.Providers[name]
	return provider, ok
}

// SetDefaultProvider sets the default provider in the global config
func (m *Manager) SetDefaultProvider(name string) {
	if m.globalConfig == nil {
		m.globalConfig = &Config{
			Providers: make(map[string]Provider),
		}
	}
	m.globalConfig.DefaultProvider = name
}

// GetDefaultProvider returns the default provider name
func (m *Manager) GetDefaultProvider() string {
	if m.globalConfig == nil {
		return ""
	}
	return m.globalConfig.DefaultProvider
}

// SetRepoProvider sets the provider for the repository config
func (m *Manager) SetRepoProvider(name string) {
	if m.repoConfig == nil {
		m.repoConfig = &RepoConfig{}
	}
	m.repoConfig.Provider = name
}

// GetRepoProvider returns the provider for the repository
func (m *Manager) GetRepoProvider() string {
	if m.repoConfig == nil {
		return ""
	}
	return m.repoConfig.Provider
}

// SetRepoModelName sets the model name for the repository config
func (m *Manager) SetRepoModelName(modelName string) {
	if m.repoConfig == nil {
		m.repoConfig = &RepoConfig{}
	}
	m.repoConfig.ModelName = modelName
}

// GetRepoModelName returns the model name for the repository
func (m *Manager) GetRepoModelName() string {
	if m.repoConfig == nil {
		return ""
	}
	return m.repoConfig.ModelName
}

// GetEffectiveProvider returns the effective provider to use
// It first checks the repo config, then falls back to the global default
func (m *Manager) GetEffectiveProvider() string {
	if m.repoConfig != nil && m.repoConfig.Provider != "" {
		return m.repoConfig.Provider
	}
	if m.globalConfig != nil {
		return m.globalConfig.DefaultProvider
	}
	return ""
}

// GetEffectiveModelName returns the effective model name to use
// It first checks the repo config, then falls back to the provider's default
func (m *Manager) GetEffectiveModelName() string {
	// First check repo config
	if m.repoConfig != nil && m.repoConfig.ModelName != "" {
		return m.repoConfig.ModelName
	}

	// Then check provider's default model
	providerName := m.GetEffectiveProvider()
	if providerName != "" {
		if provider, ok := m.GetProvider(providerName); ok && provider.ModelName != "" {
			return provider.ModelName
		}
	}

	return ""
}

// GetEffectiveTasksDir returns the effective tasks directory to use
// It first checks the repo config, then falls back to the global config
func (m *Manager) GetEffectiveTasksDir() string {
	if m.repoConfig != nil && m.repoConfig.TasksDir != "" {
		return m.repoConfig.TasksDir
	}
	if m.globalConfig != nil && m.globalConfig.TasksDir != "" {
		return m.globalConfig.TasksDir
	}

	// Default to .goline/tasks in the repo root
	repoRoot := filepath.Dir(filepath.Dir(m.repoPath))
	return filepath.Join(repoRoot, ".goline", "tasks")
}
