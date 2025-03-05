package subcmd

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/kazz187/goline/internal/config"
)

// Command variables for config commands
var (
	// Provider command variables
	providerGetName     *string
	providerSetName     *string
	providerSetAPIKey   *string
	providerSetEndpoint *string
	providerSetModel    *string
	providerRemoveName  *string

	// Default provider command variables
	defaultProviderSetName *string

	// Repository provider command variables
	repoProviderSetName *string

	// Repository model command variables
	repoModelSetName *string
)

// RegisterConfigCommands registers the config commands with the application
func RegisterConfigCommands(app *kingpin.Application) {
	// Config command
	configCmd := app.Command("config", "Manage Goline configuration")
	configCmd.Help("Manage Goline configuration, including provider settings.")

	// Provider subcommands
	providerCmd := configCmd.Command("provider", "Manage provider configurations")

	_ = providerCmd.Command("list", "List all configured providers")

	providerGetCmd := providerCmd.Command("get", "Get a provider configuration")
	providerGetName = providerGetCmd.Arg("name", "Provider name").Required().String()

	providerSetCmd := providerCmd.Command("set", "Set a provider configuration")
	providerSetName = providerSetCmd.Arg("name", "Provider name").Required().String()
	providerSetAPIKey = providerSetCmd.Flag("api-key", "API key for the provider").String()
	providerSetEndpoint = providerSetCmd.Flag("endpoint", "API endpoint for the provider").String()
	providerSetModel = providerSetCmd.Flag("model", "Default model name for the provider").String()

	providerRemoveCmd := providerCmd.Command("remove", "Remove a provider configuration")
	providerRemoveName = providerRemoveCmd.Arg("name", "Provider name").Required().String()

	// Default provider subcommands
	defaultProviderCmd := configCmd.Command("default-provider", "Manage default provider")
	_ = defaultProviderCmd.Command("get", "Get the default provider")

	defaultProviderSetCmd := defaultProviderCmd.Command("set", "Set the default provider")
	defaultProviderSetName = defaultProviderSetCmd.Arg("name", "Provider name").Required().String()

	// Repository provider subcommands
	repoProviderCmd := configCmd.Command("repo-provider", "Manage repository provider")
	_ = repoProviderCmd.Command("get", "Get the repository provider")

	repoProviderSetCmd := repoProviderCmd.Command("set", "Set the repository provider")
	repoProviderSetName = repoProviderSetCmd.Arg("name", "Provider name").Required().String()

	// Repository model subcommands
	repoModelCmd := configCmd.Command("repo-model", "Manage repository model")
	_ = repoModelCmd.Command("get", "Get the repository model")

	repoModelSetCmd := repoModelCmd.Command("set", "Set the repository model")
	repoModelSetName = repoModelSetCmd.Arg("name", "Model name").Required().String()
}

// HandleConfigCommand handles the config command
func HandleConfigCommand(cmd string) error {
	// Create a new config manager
	manager, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to create config manager: %w", err)
	}

	// Load existing configuration
	if err := manager.Load(); err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Handle the appropriate subcommand
	switch cmd {
	case "config provider list":
		return handleProviderList(manager)
	case "config provider get":
		return handleProviderGet(manager, *providerGetName)
	case "config provider set":
		return handleProviderSet(manager, *providerSetName, *providerSetAPIKey, *providerSetEndpoint, *providerSetModel)
	case "config provider remove":
		return handleProviderRemove(manager, *providerRemoveName)
	case "config default-provider get":
		return handleDefaultProviderGet(manager)
	case "config default-provider set":
		return handleDefaultProviderSet(manager, *defaultProviderSetName)
	case "config repo-provider get":
		return handleRepoProviderGet(manager)
	case "config repo-provider set":
		return handleRepoProviderSet(manager, *repoProviderSetName)
	case "config repo-model get":
		return handleRepoModelGet(manager)
	case "config repo-model set":
		return handleRepoModelSet(manager, *repoModelSetName)
	default:
		return fmt.Errorf("unknown config command: %s", cmd)
	}
}

// handleProviderList lists all configured providers
func handleProviderList(manager *config.Manager) error {
	globalConfig := manager.GetGlobalConfig()
	if globalConfig == nil || len(globalConfig.Providers) == 0 {
		fmt.Println("No providers configured")
		return nil
	}

	fmt.Println("Configured providers:")
	for name, provider := range globalConfig.Providers {
		fmt.Printf("  %s:\n", name)
		fmt.Printf("    API Key: %s\n", maskAPIKey(provider.APIKey))
		if provider.Endpoint != "" {
			fmt.Printf("    Endpoint: %s\n", provider.Endpoint)
		}
		if provider.ModelName != "" {
			fmt.Printf("    Model: %s\n", provider.ModelName)
		}
	}

	defaultProvider := manager.GetDefaultProvider()
	if defaultProvider != "" {
		fmt.Printf("\nDefault provider: %s\n", defaultProvider)
	}

	return nil
}

// maskAPIKey masks an API key for display
func maskAPIKey(apiKey string) string {
	if len(apiKey) <= 8 {
		return "********"
	}
	return apiKey[:4] + "..." + apiKey[len(apiKey)-4:]
}

// handleProviderGet gets a provider configuration
func handleProviderGet(manager *config.Manager, name string) error {
	provider, ok := manager.GetProvider(name)
	if !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	fmt.Printf("Provider: %s\n", name)
	fmt.Printf("  API Key: %s\n", maskAPIKey(provider.APIKey))
	if provider.Endpoint != "" {
		fmt.Printf("  Endpoint: %s\n", provider.Endpoint)
	}
	if provider.ModelName != "" {
		fmt.Printf("  Model: %s\n", provider.ModelName)
	}

	return nil
}

// handleProviderSet sets a provider configuration
func handleProviderSet(manager *config.Manager, name, apiKey, endpoint, modelName string) error {
	// Get existing provider if it exists
	provider, ok := manager.GetProvider(name)
	if !ok {
		provider = config.Provider{}
	}

	// Update provider with new values
	if apiKey != "" {
		provider.APIKey = apiKey
	}
	if endpoint != "" {
		provider.Endpoint = endpoint
	}
	if modelName != "" {
		provider.ModelName = modelName
	}

	// Set the provider
	manager.SetProvider(name, provider)

	// Save the configuration
	if err := manager.SaveGlobalConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Provider %s updated\n", name)
	return nil
}

// handleProviderRemove removes a provider configuration
func handleProviderRemove(manager *config.Manager, name string) error {
	globalConfig := manager.GetGlobalConfig()
	if globalConfig == nil || globalConfig.Providers == nil {
		return fmt.Errorf("provider %s not found", name)
	}

	if _, ok := globalConfig.Providers[name]; !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	// Remove the provider
	delete(globalConfig.Providers, name)

	// If this was the default provider, clear it
	if globalConfig.DefaultProvider == name {
		globalConfig.DefaultProvider = ""
	}

	// Save the configuration
	if err := manager.SaveGlobalConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Provider %s removed\n", name)
	return nil
}

// handleDefaultProviderGet gets the default provider
func handleDefaultProviderGet(manager *config.Manager) error {
	defaultProvider := manager.GetDefaultProvider()
	if defaultProvider == "" {
		fmt.Println("No default provider configured")
		return nil
	}

	fmt.Printf("Default provider: %s\n", defaultProvider)
	return nil
}

// handleDefaultProviderSet sets the default provider
func handleDefaultProviderSet(manager *config.Manager, name string) error {
	// Check if the provider exists
	if _, ok := manager.GetProvider(name); !ok {
		return fmt.Errorf("provider %s not found", name)
	}

	// Set the default provider
	manager.SetDefaultProvider(name)

	// Save the configuration
	if err := manager.SaveGlobalConfig(); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	fmt.Printf("Default provider set to %s\n", name)
	return nil
}

// handleRepoProviderGet gets the repository provider
func handleRepoProviderGet(manager *config.Manager) error {
	repoProvider := manager.GetRepoProvider()
	if repoProvider == "" {
		fmt.Println("No repository provider configured")
		return nil
	}

	fmt.Printf("Repository provider: %s\n", repoProvider)
	return nil
}

// handleRepoProviderSet sets the repository provider
func handleRepoProviderSet(manager *config.Manager, name string) error {
	// Check if the provider exists in global config
	if _, ok := manager.GetProvider(name); !ok {
		fmt.Fprintf(os.Stderr, "Warning: provider %s not found in global configuration\n", name)
	}

	// Set the repository provider
	manager.SetRepoProvider(name)

	// Save the configuration
	if err := manager.SaveRepoConfig(); err != nil {
		return fmt.Errorf("failed to save repository configuration: %w", err)
	}

	fmt.Printf("Repository provider set to %s\n", name)
	return nil
}

// handleRepoModelGet gets the repository model
func handleRepoModelGet(manager *config.Manager) error {
	repoModel := manager.GetRepoModelName()
	if repoModel == "" {
		fmt.Println("No repository model configured")
		return nil
	}

	fmt.Printf("Repository model: %s\n", repoModel)
	return nil
}

// handleRepoModelSet sets the repository model
func handleRepoModelSet(manager *config.Manager, name string) error {
	// Set the repository model
	manager.SetRepoModelName(name)

	// Save the configuration
	if err := manager.SaveRepoConfig(); err != nil {
		return fmt.Errorf("failed to save repository configuration: %w", err)
	}

	fmt.Printf("Repository model set to %s\n", name)
	return nil
}
