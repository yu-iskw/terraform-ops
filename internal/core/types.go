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

import "github.com/yu/terraform-ops/internal/ir"

// ConfigParser defines the interface for parsing Terraform configuration files.
// Configuration inspection is separate from plan/change intelligence and keeps
// its HCL-specific representation here.
type ConfigParser interface {
	ParseConfigFiles(paths []string) ([]TerraformConfig, error)
}

// GraphBuilder projects the normalized ChangeSet graph into the renderer-facing
// graph model. Dependency discovery itself belongs to the source normalizer.
type GraphBuilder interface {
	BuildGraph(changeSet *ir.ChangeSet, opts GraphOptions) (*GraphData, error)
}

// GraphGenerator defines the interface for generating graphs in different formats.
type GraphGenerator interface {
	Generate(graphData *GraphData, opts GraphOptions) (string, error)
}

// PlanSummarizer projects a normalized ChangeSet into the compatibility summary
// model consumed by the existing text/JSON/Markdown/table/plan renderers.
type PlanSummarizer interface {
	SummarizePlan(changeSet *ir.ChangeSet, opts SummaryOptions) (*PlanSummary, error)
}

// SummaryOptions holds the options for plan summarization.
type SummaryOptions struct {
	Format      SummaryFormat
	Output      string
	GroupBy     SummaryGrouping
	NoSensitive bool
	Compact     bool
	Verbose     bool
	ShowDetails bool
	Color       ColorMode
}

// PlanSummary is a renderer-facing projection of an ir.ChangeSet. It is not a
// second parsed Terraform plan representation.
type PlanSummary struct {
	PlanInfo   PlanInfo        `json:"plan_info"`
	Statistics Statistics      `json:"statistics"`
	Changes    Changes         `json:"changes"`
	Outputs    []OutputSummary `json:"outputs,omitempty"`
}

// PlanInfo contains basic information about the normalized plan.
type PlanInfo struct {
	FormatVersion string `json:"format_version"`
	Applicable    bool   `json:"applicable"`
	Complete      bool   `json:"complete"`
	Errored       bool   `json:"errored"`
}

// Statistics contains counts and breakdowns.
type Statistics struct {
	TotalChanges      int            `json:"total_changes"`
	ActionBreakdown   map[string]int `json:"action_breakdown"`
	ProviderBreakdown map[string]int `json:"provider_breakdown"`
	ResourceBreakdown map[string]int `json:"resource_breakdown"`
	ModuleBreakdown   map[string]int `json:"module_breakdown"`
}

// Changes represents grouped resource changes.
type Changes struct {
	Create  []ResourceSummary `json:"create,omitempty"`
	Update  []ResourceSummary `json:"update,omitempty"`
	Delete  []ResourceSummary `json:"delete,omitempty"`
	Replace []ResourceSummary `json:"replace,omitempty"`
	NoOp    []ResourceSummary `json:"no_op,omitempty"`
}

// ResourceSummary is a renderer-facing resource projection.
type ResourceSummary struct {
	Address       string                 `json:"address"`
	ModuleAddress string                 `json:"module_address"`
	Type          string                 `json:"type"`
	Name          string                 `json:"name"`
	Provider      string                 `json:"provider"`
	Actions       []string               `json:"actions"`
	Sensitive     bool                   `json:"sensitive"`
	KeyChanges    map[string]interface{} `json:"key_changes,omitempty"`
}

// OutputSummary is a renderer-facing output projection.
type OutputSummary struct {
	Name      string      `json:"name"`
	Actions   []string    `json:"actions"`
	Sensitive bool        `json:"sensitive"`
	Value     interface{} `json:"value,omitempty"`
}

// SummaryFormat represents the output format for the summary.
type SummaryFormat string

const (
	FormatText     SummaryFormat = "text"
	FormatJSON     SummaryFormat = "json"
	FormatMarkdown SummaryFormat = "markdown"
	FormatTable    SummaryFormat = "table"
	FormatPlan     SummaryFormat = "plan"
)

// SummaryGrouping represents the strategy for grouping resources in the summary.
type SummaryGrouping = GroupingStrategy

// ColorMode represents the color output mode.
type ColorMode string

const (
	ColorAuto   ColorMode = "auto"
	ColorAlways ColorMode = "always"
	ColorNever  ColorMode = "never"
)

// TerraformConfig represents terraform block configuration details used by the
// show-terraform command. It is intentionally independent of plan ChangeSet IR.
type TerraformConfig struct {
	Path              string            `json:"path"`
	RequiredVersion   string            `json:"required_version,omitempty"`
	Backend           *Backend          `json:"backend,omitempty"`
	RequiredProviders map[string]string `json:"required_providers"`
}

// Backend represents a backend configuration.
type Backend struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config,omitempty"`
}

// GraphOptions holds options for graph rendering. NoLocals remains for CLI
// compatibility; Terraform's machine-readable plan configuration does not expose
// local declarations, so normalized ChangeSet graphs do not synthesize them.
type GraphOptions struct {
	Format        GraphFormat
	Output        string
	GroupBy       GroupingStrategy
	NoDataSources bool
	NoOutputs     bool
	NoVariables   bool
	NoLocals      bool
	NoModules     bool
	Compact       bool
	Verbose       bool
}

// GraphData is a renderer-facing projection of ir.DependencyGraph.
type GraphData struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// GraphNode represents a rendered graph node.
type GraphNode struct {
	ID        string
	Address   string
	Type      string
	Name      string
	Module    string
	Provider  string
	Actions   []string
	Sensitive bool
}

// GraphEdge represents a rendered graph edge.
type GraphEdge struct {
	From string
	To   string
}

// GraphFormat represents the output format for the graph.
type GraphFormat string

const (
	FormatGraphviz GraphFormat = "graphviz"
	FormatMermaid  GraphFormat = "mermaid"
	FormatPlantUML GraphFormat = "plantuml"
)

// GroupingStrategy represents the strategy for grouping nodes in the graph.
type GroupingStrategy string

const (
	GroupByModule       GroupingStrategy = "module"
	GroupByAction       GroupingStrategy = "action"
	GroupByResourceType GroupingStrategy = "resource_type"
	GroupByProvider     GroupingStrategy = "provider"
)

// ActionType represents the type of action to be performed on a resource.
type ActionType string

const (
	ActionCreate  ActionType = "create"
	ActionUpdate  ActionType = "update"
	ActionDelete  ActionType = "delete"
	ActionReplace ActionType = "replace"
	ActionNoOp    ActionType = "no-op"
)

// NodeType represents the type of a rendered graph node.
type NodeType string

const (
	NodeTypeResource NodeType = "resource"
	NodeTypeData     NodeType = "data"
	NodeTypeOutput   NodeType = "output"
	NodeTypeVariable NodeType = "variable"
	NodeTypeLocal    NodeType = "local"
)
