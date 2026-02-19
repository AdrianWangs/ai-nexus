package main

// WebUI config.

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the WebUI.
type Config struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`

	Host string `yaml:"host"`
	Port int    `yaml:"port"`

	Registry RegistryConfig `yaml:"registry"`
	Session  SessionConfig  `yaml:"session"`
	Server   ServerConfig   `yaml:"server"`
}

// RegistryConfig holds registry configuration.
type RegistryConfig struct {
	Type      string   `yaml:"type"`
	Endpoints []string `yaml:"endpoints"`
	Prefix    string   `yaml:"prefix"`
	TTL       int64    `yaml:"ttl"`
}

// SessionConfig holds session storage configuration.
type SessionConfig struct {
	Type     string        `yaml:"type"`
	Addr     string        `yaml:"addr"`
	Password string        `yaml:"password"`
	DB       int           `yaml:"db"`
	TTL      time.Duration `yaml:"ttl"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	CORSOrigins  []string      `yaml:"cors_origins"`
}

// LoadConfig loads configuration from a YAML file.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, err
	}

	// Set defaults
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 8080
	}
	if config.Server.ReadTimeout == 0 {
		config.Server.ReadTimeout = 30 * time.Second
	}
	if config.Server.WriteTimeout == 0 {
		config.Server.WriteTimeout = 120 * time.Second
	}

	return &config, nil
}
