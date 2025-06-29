// Copyright 2025 yu-iskw
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

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
