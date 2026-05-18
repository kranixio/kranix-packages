// Package clilib provides shared CLI utilities for Kranix CLI tools.
package clilib

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the CLI configuration
type Config struct {
	CurrentContext string             `yaml:"currentContext"`
	Contexts       map[string]Context `yaml:"contexts"`
	Defaults       Defaults           `yaml:"defaults"`
}

// Context represents a Kranix API context
type Context struct {
	Name        string `yaml:"name"`
	ServerURL   string `yaml:"serverURL"`
	APIKey      string `yaml:"apiKey"`
	Namespace   string `yaml:"namespace"`
	Timeout     int    `yaml:"timeout"`
	InsecureTLS bool   `yaml:"insecureSkipTLSVerify"`
}

// Defaults represents default configuration values
type Defaults struct {
	Namespace string `yaml:"namespace"`
	Output    string `yaml:"output"`
	Timeout   int    `yaml:"timeout"`
}

// LoadConfig loads configuration from a file
func LoadConfig(configFile string) (*Config, error) {
	if configFile == "" {
		// Try default locations
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}

		defaultLocations := []string{
			filepath.Join(homeDir, ".kranix", "config.yaml"),
			filepath.Join(homeDir, ".config", "kranix", "config.yaml"),
			".kranix.yaml",
		}

		for _, loc := range defaultLocations {
			if _, err := os.Stat(loc); err == nil {
				configFile = loc
				break
			}
		}

		if configFile == "" {
			// Return default config if no file found
			return DefaultConfig(), nil
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to a file
func SaveConfig(config *Config, configFile string) error {
	if configFile == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		configDir := filepath.Join(homeDir, ".kranix")
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		configFile = filepath.Join(configDir, "config.yaml")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configFile, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		CurrentContext: "default",
		Contexts: map[string]Context{
			"default": {
				Name:        "default",
				ServerURL:   "http://localhost:8080",
				Namespace:   "default",
				Timeout:     30,
				InsecureTLS: false,
			},
		},
		Defaults: Defaults{
			Namespace: "default",
			Output:    "table",
			Timeout:   30,
		},
	}
}

// GetCurrentContext returns the current context
func (c *Config) GetCurrentContext() (*Context, error) {
	if c.CurrentContext == "" {
		return nil, fmt.Errorf("no current context set")
	}

	ctx, ok := c.Contexts[c.CurrentContext]
	if !ok {
		return nil, fmt.Errorf("context not found: %s", c.CurrentContext)
	}

	return &ctx, nil
}

// AddContext adds a new context
func (c *Config) AddContext(ctx Context) error {
	if c.Contexts == nil {
		c.Contexts = make(map[string]Context)
	}
	c.Contexts[ctx.Name] = ctx
	return nil
}

// RemoveContext removes a context
func (c *Config) RemoveContext(name string) error {
	if name == c.CurrentContext {
		return fmt.Errorf("cannot remove current context")
	}
	delete(c.Contexts, name)
	return nil
}

// SetCurrentContext sets the current context
func (c *Config) SetCurrentContext(name string) error {
	if _, ok := c.Contexts[name]; !ok {
		return fmt.Errorf("context not found: %s", name)
	}
	c.CurrentContext = name
	return nil
}
