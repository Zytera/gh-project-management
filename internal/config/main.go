package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type ConfigKey struct{}

// GlobalConfig holds all contexts and the current active context
type GlobalConfig struct {
	CurrentContext string             `yaml:"current-context"`
	Contexts       map[string]Context `yaml:"contexts"`
}

// Context represents a project configuration
type Context struct {
	Org         string            `yaml:"org"`
	ProjectID   string            `yaml:"project_id"`
	ProjectName string            `yaml:"project_name"`
	DefaultRepo string            `yaml:"default_repo"`
	TeamRepos   map[string]string `yaml:"team_repos"` // Team name -> Repo name
}

// Config is the active context configuration (for backwards compatibility in code)
type Config struct {
	Org         string
	ProjectID   string
	ProjectName string
	DefaultRepo string
	TeamRepos   map[string]string
}

// GetConfigPath returns the path to the global config file
func GetConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config", "gh-project-management")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", fmt.Errorf("error creating config directory: %w", err)
	}

	return filepath.Join(configDir, "config.yaml"), nil
}

// Load reads the global config and returns the active context configuration
func Load() (*Config, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("no configuration found. Please run 'gh project-management init' to set up your first project")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file %s: %w", configPath, err)
	}

	var globalConfig GlobalConfig
	if err := yaml.Unmarshal(data, &globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	if globalConfig.CurrentContext == "" {
		return nil, fmt.Errorf("no current context set. Use 'gh project-management context use <name>'")
	}

	ctx, exists := globalConfig.Contexts[globalConfig.CurrentContext]
	if !exists {
		return nil, fmt.Errorf("current context '%s' not found in configuration", globalConfig.CurrentContext)
	}

	config := &Config{
		Org:         ctx.Org,
		ProjectID:   ctx.ProjectID,
		ProjectName: ctx.ProjectName,
		DefaultRepo: ctx.DefaultRepo,
		TeamRepos:   ctx.TeamRepos,
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration for context '%s': %w", globalConfig.CurrentContext, err)
	}

	return config, nil
}

// LoadGlobal loads the entire global configuration
func LoadGlobal() (*GlobalConfig, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return nil, err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &GlobalConfig{
			Contexts: make(map[string]Context),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file: %w", err)
	}

	var globalConfig GlobalConfig
	if err := yaml.Unmarshal(data, &globalConfig); err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	if globalConfig.Contexts == nil {
		globalConfig.Contexts = make(map[string]Context)
	}

	return &globalConfig, nil
}

// Save writes the global configuration to disk
func Save(globalConfig *GlobalConfig) error {
	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	data, err := yaml.Marshal(globalConfig)
	if err != nil {
		return fmt.Errorf("error marshaling configuration: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing configuration file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Org == "" {
		return fmt.Errorf("field 'org' is required")
	}
	if c.ProjectID == "" {
		return fmt.Errorf("field 'project_id' is required")
	}
	if c.DefaultRepo == "" {
		return fmt.Errorf("field 'default_repo' is required")
	}
	if len(c.TeamRepos) == 0 {
		return fmt.Errorf("at least one team repository is required")
	}
	return nil
}

// Validate checks if a context is valid
func (c *Context) Validate() error {
	if c.Org == "" {
		return fmt.Errorf("field 'org' is required")
	}
	if c.ProjectID == "" {
		return fmt.Errorf("field 'project_id' is required")
	}
	if c.DefaultRepo == "" {
		return fmt.Errorf("field 'default_repo' is required")
	}
	if len(c.TeamRepos) == 0 {
		return fmt.Errorf("at least one team repository is required")
	}
	return nil
}
