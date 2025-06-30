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

package commands

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/terraform/config"
)

// ShowTerraformCommand represents the show-terraform command with dependency injection
type ShowTerraformCommand struct {
	configParser core.ConfigParser
}

// NewShowTerraformCommand creates a new show-terraform command with injected dependencies
func NewShowTerraformCommand(configParser core.ConfigParser) *ShowTerraformCommand {
	return &ShowTerraformCommand{
		configParser: configParser,
	}
}

// LegacyOutput represents the expected JSON structure for backward compatibility
type LegacyOutput struct {
	Path      string `json:"path"`
	Terraform struct {
		RequiredVersion   string            `json:"required_version"`
		Backend           *LegacyBackend    `json:"backend,omitempty"`
		RequiredProviders map[string]string `json:"required_providers"`
	} `json:"terraform"`
}

// LegacyBackend represents the backend structure for backward compatibility
type LegacyBackend struct {
	Type   string                 `json:"type"`
	Config map[string]interface{} `json:"config"`
}

// Command returns the cobra command for show-terraform
func (c *ShowTerraformCommand) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-terraform <path...>",
		Short: "Displays information from the terraform block in workspaces",
		Long:  `The show-terraform command inspects Terraform configuration files (*.tf) in the specified paths and outputs information contained in the terraform block (required_version, backend, and required_providers) in JSON format. It does not recurse into subdirectories.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runShowTerraform(args)
		},
	}

	return cmd
}

// runShowTerraform executes the show-terraform command
func (c *ShowTerraformCommand) runShowTerraform(paths []string) error {
	allInfo, err := c.configParser.ParseConfigFiles(paths)
	if err != nil {
		return fmt.Errorf("failed to parse config files: %w", err)
	}

	// Transform to legacy format for backward compatibility
	var legacyOutputs []LegacyOutput
	for _, info := range allInfo {
		legacy := LegacyOutput{
			Path: info.Path,
		}
		legacy.Terraform.RequiredVersion = info.RequiredVersion
		legacy.Terraform.RequiredProviders = info.RequiredProviders

		if info.Backend != nil {
			legacy.Terraform.Backend = &LegacyBackend{
				Type:   info.Backend.Type,
				Config: make(map[string]interface{}),
			}
			for k, v := range info.Backend.Config {
				legacy.Terraform.Backend.Config[k] = v
			}
		}

		legacyOutputs = append(legacyOutputs, legacy)
	}

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(legacyOutputs); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	fmt.Println(out.String())
	return nil
}

// DefaultShowTerraformCommand creates a show-terraform command with default dependencies
func DefaultShowTerraformCommand() *ShowTerraformCommand {
	return NewShowTerraformCommand(config.NewParser())
}
