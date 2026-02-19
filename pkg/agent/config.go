// Package agent provides the agent framework for building AI agents.
package agent

import (
	"os"
	"time"

	"github.com/AdrianWangs/ai-nexus/pkg/llm"
	"gopkg.in/yaml.v3"
)

// Config holds the configuration for an agent.
type Config struct {
	// Agent identity
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`

	// Network configuration
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	HTTPPort int    `yaml:"http_port"`

	// Agent type
	Type string `yaml:"type"` // "normal" or "orchestrator"

	// LLM configuration
	LLM LLMConfig `yaml:"llm"`

	// Registry configuration
	Registry RegistryConfig `yaml:"registry"`

	// Capabilities and skills
	Capabilities []string      `yaml:"capabilities"`
	Skills       []SkillConfig `yaml:"skills"`
}

// LLMConfig holds LLM configuration.
type LLMConfig struct {
	Provider    string        `yaml:"provider"` // "openai"
	APIKey      string        `yaml:"api_key"`
	BaseURL     string        `yaml:"base_url"`
	Model       string        `yaml:"model"`
	Temperature float64       `yaml:"temperature"`
	MaxTokens   int           `yaml:"max_tokens"`
	Timeout     time.Duration `yaml:"timeout"`
}

// RegistryConfig holds registry configuration.
type RegistryConfig struct {
	Type      string   `yaml:"type"` // "etcd"
	Endpoints []string `yaml:"endpoints"`
	Prefix    string   `yaml:"prefix"`
	TTL       int64    `yaml:"ttl"`
}

// SkillConfig holds skill configuration.
type SkillConfig struct {
	Name        string   `yaml:"name"`
	Description string   `yaml:"description"`
	Keywords    []string `yaml:"keywords"`
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
	if config.Version == "" {
		config.Version = "1.0.0"
	}
	if config.Type == "" {
		config.Type = "normal"
	}

	return &config, nil
}

// NewLLMClient creates an LLM client from config.
func NewLLMClient(config LLMConfig) llm.Client {
	switch config.Provider {
	case "openai", "":
		return llm.NewOpenAIClient(llm.OpenAIConfig{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
			Timeout: config.Timeout,
		})
	default:
		return llm.NewOpenAIClient(llm.OpenAIConfig{
			APIKey:  config.APIKey,
			BaseURL: config.BaseURL,
			Model:   config.Model,
			Timeout: config.Timeout,
		})
	}
}
