package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration.
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database,omitempty"`
	Logging  LoggingConfig  `yaml:"logging,omitempty"`
	Auth     AuthConfig     `yaml:"auth,omitempty"`
}

// ServerConfig defines server configuration.
type ServerConfig struct {
	Port         int    `yaml:"port"`
	Host         string `yaml:"host,omitempty"`
	ReadTimeout  int    `yaml:"readTimeout,omitempty"`
	WriteTimeout int    `yaml:"writeTimeout,omitempty"`
}

// DatabaseConfig defines database configuration.
type DatabaseConfig struct {
	Type     string `yaml:"type"` // postgres, mysql, sqlite
	Host     string `yaml:"host,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	Name     string `yaml:"name"`
	User     string `yaml:"user,omitempty"`
	Password string `yaml:"password,omitempty"`
	DSN      string `yaml:"dsn,omitempty"`
}

// LoggingConfig defines logging configuration.
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, console
	Output string `yaml:"output"` // stdout, stderr, file
}

// AuthConfig defines authentication configuration.
type AuthConfig struct {
	JWTSecret            string `yaml:"jwtSecret"`
	TokenExpirationHours int    `yaml:"tokenExpirationHours"`
	APIKeyHeader         string `yaml:"apiKeyHeader,omitempty"`
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvOverrides(&cfg)

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides to the config.
func applyEnvOverrides(cfg *Config) {
	if port := os.Getenv("SERVER_PORT"); port != "" {
		// Parse port from env
		cfg.Server.Port = parseInt(port)
	}

	if dbDSN := os.Getenv("DATABASE_DSN"); dbDSN != "" {
		cfg.Database.DSN = dbDSN
	}

	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		cfg.Auth.JWTSecret = jwtSecret
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = strings.ToLower(logLevel)
	}
}

// parseInt parses a string to int with a default of 0.
func parseInt(s string) int {
	var i int
	if _, err := fmt.Sscanf(s, "%d", &i); err != nil {
		return 0
	}
	return i
}

// MustLoad loads configuration and panics on error.
func MustLoad(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		panic(err)
	}
	return cfg
}
