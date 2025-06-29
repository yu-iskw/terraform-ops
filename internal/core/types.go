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

package core

// PlanParser defines the interface for parsing Terraform plan files
type PlanParser interface {
	ParsePlanFile(filename string) (*TerraformPlan, error)
}

// ConfigParser defines the interface for parsing Terraform configuration files
type ConfigParser interface {
	ParseConfigFiles(paths []string) ([]TerraformConfig, error)
}

// GraphBuilder defines the interface for building graph data from Terraform plans
type GraphBuilder interface {
	BuildGraph(plan *TerraformPlan, opts GraphOptions) (*GraphData, error)
}

// GraphGenerator defines the interface for generating graphs in different formats
type GraphGenerator interface {
	Generate(graphData *GraphData, opts GraphOptions) (string, error)
}

// TerraformPlan represents a parsed Terraform plan
type TerraformPlan struct {
	FormatVersion   string                  `json:"format_version"`
	ResourceChanges []ResourceChange        `json:"resource_changes"`
	OutputChanges   map[string]OutputChange `json:"output_changes"`
	Configuration   Configuration           `json:"configuration"`
	Variables       map[string]Variable     `json:"variables"`
	Applicable      bool                    `json:"applicable"`
	Complete        bool                    `json:"complete"`
	Errored         bool                    `json:"errored"`
}

// ResourceChange represents a resource change in the plan
type ResourceChange struct {
	Address       string `json:"address"`
	ModuleAddress string `json:"module_address"`
	Mode          string `json:"mode"`
	Type          string `json:"type"`
	Name          string `json:"name"`
	Change        Change `json:"change"`
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

// OutputConfig represents an output configuration
type OutputConfig struct {
	Expression map[string]interface{} `json:"expression"`
	Sensitive  bool                   `json:"sensitive"`
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

// TerraformConfig represents the terraform block configuration details
type TerraformConfig struct {
	Path              string            `json:"path"`
	RequiredVersion   string            `json:"required_version,omitempty"`
	Backend           *Backend          `json:"backend,omitempty"`
	RequiredProviders map[string]string `json:"required_providers"`
}

// Backend represents the backend configuration
type Backend struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config,omitempty"`
}

// GraphOptions holds the options for graph generation
type GraphOptions struct {
	Format        GraphFormat
	Output        string
	GroupBy       GroupingStrategy
	NoDataSources bool
	NoOutputs     bool
	NoVariables   bool
	NoLocals      bool
	Compact       bool
	Verbose       bool
}

// GraphData holds all the nodes and edges of the graph
type GraphData struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// GraphNode represents a node in the graph
type GraphNode struct {
	ID        string
	Address   string
	Type      string
	Name      string
	Module    string
	Provider  string // Provider name (e.g., "aws", "google", "azurerm")
	Actions   []string
	Sensitive bool
}

// GraphEdge represents an edge between two nodes in the graph
type GraphEdge struct {
	From string
	To   string
}

// GraphFormat represents the output format for the graph
type GraphFormat string

const (
	FormatGraphviz GraphFormat = "graphviz"
	FormatMermaid  GraphFormat = "mermaid"
	FormatPlantUML GraphFormat = "plantuml"
)

// GroupingStrategy represents the strategy for grouping nodes in the graph
type GroupingStrategy string

const (
	GroupByModule       GroupingStrategy = "module"
	GroupByAction       GroupingStrategy = "action"
	GroupByResourceType GroupingStrategy = "resource_type"
)

// ActionType represents the type of action to be performed on a resource
type ActionType string

const (
	ActionCreate  ActionType = "create"
	ActionUpdate  ActionType = "update"
	ActionDelete  ActionType = "delete"
	ActionReplace ActionType = "replace"
	ActionNoOp    ActionType = "no-op"
)

// NodeType represents the type of a node in the graph
type NodeType string

const (
	NodeTypeResource NodeType = "resource"
	NodeTypeData     NodeType = "data"
	NodeTypeOutput   NodeType = "output"
	NodeTypeVariable NodeType = "variable"
	NodeTypeLocal    NodeType = "local"
)
