package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveConfig writes the given NetworkConfig to a JSON file at the specified path
func SaveConfig(cfg *NetworkConfig, path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadConfig reads a NetworkConfig from a JSON file at the specified path
func LoadConfig(path string) (*NetworkConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil // Return nil config if it doesn't exist
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg NetworkConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
