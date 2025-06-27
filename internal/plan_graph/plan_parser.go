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

package plan_graph

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// TerraformPlan represents the structure of a Terraform plan JSON file
// Based on https://developer.hashicorp.com/terraform/internals/json-format
type TerraformPlan struct {
	FormatVersion   string                  `json:"format_version"`
	PriorState      interface{}             `json:"prior_state"`
	PlannedValues   interface{}             `json:"planned_values"`
	ResourceChanges []ResourceChange        `json:"resource_changes"`
	OutputChanges   map[string]OutputChange `json:"output_changes"`
	Configuration   Configuration           `json:"configuration"`
	Variables       map[string]Variable     `json:"variables"`
	// Additional fields from the official specification
	Applicable       bool        `json:"applyable"`
	Complete         bool        `json:"complete"`
	Errored          bool        `json:"errored"`
	ProposedUnknown  interface{} `json:"proposed_unknown"`
	Checks           []Check     `json:"checks,omitempty"`
	Timestamp        string      `json:"timestamp,omitempty"`
	TerraformVersion string      `json:"terraform_version,omitempty"`
}

// Configuration represents the configuration section of the plan
type Configuration struct {
	ProviderConfig map[string]interface{} `json:"provider_config"`
	RootModule     RootModule             `json:"root_module"`
}

// RootModule represents the root module configuration
type RootModule struct {
	Resources   []ConfigurationResource   `json:"resources"`
	ModuleCalls map[string]ModuleCall     `json:"module_calls"`
	Outputs     map[string]OutputConfig   `json:"outputs"`
	Variables   map[string]VariableConfig `json:"variables"`
	Locals      map[string]LocalConfig    `json:"locals"`
}

// ConfigurationResource represents a resource in the configuration section
type ConfigurationResource struct {
	Address     string                 `json:"address"`
	Mode        string                 `json:"mode"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	ProviderKey string                 `json:"provider_config_key"`
	Expressions map[string]interface{} `json:"expressions"`
	SchemaVer   int                    `json:"schema_version"`
	DependsOn   []string               `json:"depends_on"`
}

// ResourceChange represents a resource change in the plan
type ResourceChange struct {
	Address       string                 `json:"address"`
	ModuleAddress string                 `json:"module_address"`
	Mode          string                 `json:"mode"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Index         interface{}            `json:"index"`
	Deposed       interface{}            `json:"deposed"`
	Change        Change                 `json:"change"`
	ReplacePaths  [][]interface{}        `json:"replace_paths"`
	Importing     map[string]interface{} `json:"importing"`
}

// Change represents the change details for a resource
type Change struct {
	Actions         []string    `json:"actions"`
	Before          interface{} `json:"before"`
	After           interface{} `json:"after"`
	AfterUnknown    interface{} `json:"after_unknown"`
	BeforeSensitive interface{} `json:"before_sensitive"`
	AfterSensitive  interface{} `json:"after_sensitive"`
}

// OutputChange represents a change to an output value
type OutputChange struct {
	Change Change `json:"change"`
}

// OutputConfig represents an output configuration
type OutputConfig struct {
	Expression map[string]interface{} `json:"expression"`
	Sensitive  bool                   `json:"sensitive"`
}

// ModuleCall represents a module call in the configuration
type ModuleCall struct {
	Source      string                 `json:"source"`
	Expressions map[string]interface{} `json:"expressions"`
	Module      *ModuleConfig          `json:"module"`
}

// ModuleConfig represents a module's configuration
type ModuleConfig struct {
	Resources   []ConfigurationResource `json:"resources"`
	ModuleCalls map[string]ModuleCall   `json:"module_calls"`
	Outputs     map[string]interface{}  `json:"outputs"`
	Variables   map[string]interface{}  `json:"variables"`
}

// Variable represents a variable value at the top level of the plan
type Variable struct {
	Value interface{} `json:"value"`
}

// VariableConfig represents a variable configuration
type VariableConfig struct {
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
	Type        string      `json:"type"`
	Sensitive   bool        `json:"sensitive"`
}

// LocalConfig represents a local value configuration
type LocalConfig struct {
	Expression map[string]interface{} `json:"expression"`
}

// Check represents a checkable object in the configuration
// Based on the experimental checks representation from the official spec
type Check struct {
	Address   CheckAddress    `json:"address"`
	Status    string          `json:"status"`
	Instances []CheckInstance `json:"instances,omitempty"`
}

// CheckAddress represents the address of a checkable object
type CheckAddress struct {
	Kind        string      `json:"kind"`
	ToDisplay   string      `json:"to_display"`
	Mode        string      `json:"mode,omitempty"`
	Type        string      `json:"type,omitempty"`
	Name        string      `json:"name"`
	Module      string      `json:"module,omitempty"`
	InstanceKey interface{} `json:"instance_key,omitempty"`
}

// CheckInstance represents an instance of a checkable object
type CheckInstance struct {
	Address  CheckAddress   `json:"address"`
	Status   string         `json:"status"`
	Problems []CheckProblem `json:"problems,omitempty"`
}

// CheckProblem represents a problem with a check
type CheckProblem struct {
	Message string `json:"message"`
}

// ParsePlanFile reads and parses a Terraform plan JSON file
func ParsePlanFile(filename string) (*TerraformPlan, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open plan file: %w", err)
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close plan file: %w", closeErr)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read plan file: %w", err)
	}

	var plan TerraformPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	// Validate the plan structure
	if err := plan.Validate(); err != nil {
		return nil, fmt.Errorf("invalid plan structure: %w", err)
	}

	return &plan, nil
}

// Validate checks if the plan has a valid structure
func (p *TerraformPlan) Validate() error {
	if p.FormatVersion == "" {
		return fmt.Errorf("missing format_version")
	}

	// Check if this is a valid format version we support
	// According to the official spec, minor version increments (1.1, 1.2, etc.)
	// are backward-compatible changes, so we support any 1.x version
	if !strings.HasPrefix(p.FormatVersion, "1.") {
		return fmt.Errorf("unsupported format version: %s (only 1.x versions are supported)", p.FormatVersion)
	}

	return nil
}

// IsApplyable returns whether the plan can be applied
func (p *TerraformPlan) IsApplyable() bool {
	return p.Applicable
}

// IsComplete returns whether the plan is complete
func (p *TerraformPlan) IsComplete() bool {
	return p.Complete
}

// HasErrors returns whether the plan has errors
func (p *TerraformPlan) HasErrors() bool {
	return p.Errored
}

// GetResourceCount returns the total number of resource changes
func (p *TerraformPlan) GetResourceCount() int {
	return len(p.ResourceChanges)
}

// GetOutputCount returns the total number of output changes
func (p *TerraformPlan) GetOutputCount() int {
	return len(p.OutputChanges)
}

// GetVariableCount returns the total number of variables
func (p *TerraformPlan) GetVariableCount() int {
	return len(p.Variables)
}

// GetDataResourceCount returns the number of data source changes
func (p *TerraformPlan) GetDataResourceCount() int {
	count := 0
	for _, change := range p.ResourceChanges {
		if change.Mode == "data" {
			count++
		}
	}
	return count
}

// GetManagedResourceCount returns the number of managed resource changes
func (p *TerraformPlan) GetManagedResourceCount() int {
	count := 0
	for _, change := range p.ResourceChanges {
		if change.Mode == "managed" {
			count++
		}
	}
	return count
}

// GetChecksCount returns the number of checks
func (p *TerraformPlan) GetChecksCount() int {
	return len(p.Checks)
}

// GetFailedChecksCount returns the number of failed checks
func (p *TerraformPlan) GetFailedChecksCount() int {
	count := 0
	for _, check := range p.Checks {
		if check.Status == "fail" || check.Status == "error" {
			count++
		}
	}
	return count
}
