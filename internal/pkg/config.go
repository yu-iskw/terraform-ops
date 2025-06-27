package pkg

import (
	"os"
	"path/filepath"
)

// Config holds the application configuration
type Config struct {
	Verbose      bool   `yaml:"verbose"`
	ConfigDir    string `yaml:"config_dir"`
	LogLevel     string `yaml:"log_level"`
	TerraformBin string `yaml:"terraform_bin"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		Verbose:      false,
		ConfigDir:    filepath.Join(homeDir, ".terraform-ops"),
		LogLevel:     "info",
		TerraformBin: "terraform",
	}
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// TODO: Implement config file loading (YAML/JSON)
	return DefaultConfig(), nil
}
