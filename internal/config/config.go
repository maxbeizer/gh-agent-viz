package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Repos           []string `yaml:"repos"`
	RefreshInterval int      `yaml:"refreshInterval"`
	DefaultFilter   string   `yaml:"defaultFilter"`
	Animations      *bool    `yaml:"animations,omitempty"`
	AsciiHeader     *bool    `yaml:"asciiHeader,omitempty"`
	Theme           string   `yaml:"theme,omitempty"`
}

// AnimationsEnabled returns whether animations are enabled (default: true).
func (c *Config) AnimationsEnabled() bool {
	if c.Animations == nil {
		return true
	}
	return *c.Animations
}

// AsciiHeaderEnabled returns whether the ASCII art header is enabled (default: true).
func (c *Config) AsciiHeaderEnabled() bool {
	if c.AsciiHeader == nil {
		return true
	}
	return *c.AsciiHeader
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Repos:           []string{},
		RefreshInterval: 30,
		DefaultFilter:   "",
	}
}

// Load loads configuration from a file
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// If no path provided, try default location
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return cfg, nil // Return default config if can't get home dir
		}
		path = filepath.Join(homeDir, ".gh-agent-viz.yml")
	}

	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return cfg, nil // Return default config if file doesn't exist
	}

	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Save saves configuration to a file
func Save(cfg *Config, path string) error {
	// If no path provided, use default location
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		path = filepath.Join(homeDir, ".gh-agent-viz.yml")
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	// Write file
	return os.WriteFile(path, data, 0600)
}
