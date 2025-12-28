package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ConfigKey struct{}

type Config struct {
	ProjectManagementRepoName string   `json:"project_managment_repo_name"`
	Org                       string   `json:"org"`
	ProjectID                 string   `json:"project_id"`
	ReposToTransferTasks      []string `json:"repos_to_transfer_tasks"`
}

func Load() (*Config, error) {
	configFileName := ".gh-project-managment.json"

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current directory: %w", err)
	}

	configPath := filepath.Join(currentDir, configFileName)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration file %s: %w", configPath, err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("error parsing configuration file: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func (c *Config) Validate() error {
	if c.ProjectManagementRepoName == "" {
		return fmt.Errorf("field 'project_managment_repo_name' is required")
	}
	if c.Org == "" {
		return fmt.Errorf("field 'org' is required")
	}
	if c.ProjectID == "" {
		return fmt.Errorf("field 'project_id' is required")
	}
	if len(c.ReposToTransferTasks) == 0 {
		return fmt.Errorf("field 'repos_to_transfer_tasks' must contain at least one repository")
	}
	return nil
}
