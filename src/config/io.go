package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveConfig writes the given NetworkConfig to a JSON file at the specified path atomically with secure permissions (0600).
func SaveConfig(cfg *NetworkConfig, path string) error {
	if cfg == nil {
		return fmt.Errorf("cannot save nil network config")
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid network config: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return WriteFileAtomic(path, data, 0600)
}

// LoadConfig reads a NetworkConfig from a JSON file at the specified path and validates it.
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

	if err := cfg.Validate(); err != nil {
		return &cfg, fmt.Errorf("loaded configuration failed validation: %w", err)
	}

	return &cfg, nil
}

// Verbose controls whether per-file log output is printed during generation.
var Verbose bool

// SkipCleanup controls whether old artifact cleanup is skipped during network generation.
var SkipCleanup bool

// WriteFileAtomic writes data to a temporary file in the target directory and atomically renames it to destPath.
func WriteFileAtomic(destPath string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", dir, err)
	}

	tmpFile, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file in %s: %w", dir, err)
	}
	tmpName := tmpFile.Name()

	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpName) // remove if rename failed or didn't execute
	}()

	if err := tmpFile.Chmod(perm); err != nil {
		return fmt.Errorf("failed to set permissions on %s: %w", tmpName, err)
	}

	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed writing to temp file %s: %w", tmpName, err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed syncing temp file %s: %w", tmpName, err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed closing temp file %s: %w", tmpName, err)
	}

	if err := os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("failed atomic rename %s -> %s: %w", tmpName, destPath, err)
	}

	if Verbose {
		fmt.Printf("Generated %s\n", destPath)
	}

	return nil
}
