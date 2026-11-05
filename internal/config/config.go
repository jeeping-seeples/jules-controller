package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"
)

// Config represents application configuration loaded from file and environment.
type Config struct {
	APIKey  string `toml:"api_key"`
	BaseURL string `toml:"base_url"`
}

// DefaultBaseURL is the Google Jules v1alpha endpoint.
const DefaultBaseURL = "https://jules.googleapis.com"

// DefaultConfigDir returns the standard configuration directory.
func DefaultConfigDir() (string, error) {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "jules-tui"), nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("unable to determine user home directory: %w", err)
	}
	return filepath.Join(home, ".config", "jules-tui"), nil
}

// DefaultConfigFile returns the standard configuration file path.
func DefaultConfigFile() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

// Load loads configuration with precedence:
// flagOverride > JULES_API_KEY env var > config.toml
// If flagOverride is non-empty, it takes top precedence.
func Load(explicitConfigFile string, flagAPIKey string) (*Config, error) {
	cfg := &Config{
		BaseURL: DefaultBaseURL,
	}

	configFile := explicitConfigFile
	if configFile == "" {
		if defaultPath, err := DefaultConfigFile(); err == nil {
			configFile = defaultPath
		}
	}

	if configFile != "" {
		data, err := os.ReadFile(configFile)
		if err == nil {
			if err := toml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
			}
		} else if !errors.Is(err, os.ErrNotExist) && explicitConfigFile != "" {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	}

	if envKey := os.Getenv("JULES_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	if envURL := os.Getenv("JULES_BASE_URL"); envURL != "" {
		cfg.BaseURL = envURL
	}

	if flagAPIKey != "" {
		cfg.APIKey = flagAPIKey
	}

	return cfg, nil
}
