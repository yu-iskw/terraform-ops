// Copyright 2026 yu-iskw
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

package ir

import (
	"fmt"
	"sort"
	"strings"
)

const SchemaVersion = "1.0"

type Engine string

const (
	EngineTerraform Engine = "terraform"
	EngineOpenTofu  Engine = "opentofu"
	EngineUnknown   Engine = "unknown-compatible"
)

type SourceMetadata struct {
	Engine            Engine `json:"engine"`
	EngineVersion     string `json:"engine_version,omitempty"`
	PlanFormatVersion string `json:"plan_format_version"`
}

type PlanMetadata struct {
	Applyable bool `json:"applyable"`
	Complete  bool `json:"complete"`
	Errored   bool `json:"errored"`
}

type Address string

type ResourceMode string

const (
	ResourceModeManaged ResourceMode = "managed"
	ResourceModeData    ResourceMode = "data"
)

type ActionKind string

const (
	ActionNoOp                 ActionKind = "no_op"
	ActionCreate               ActionKind = "create"
	ActionUpdate               ActionKind = "update"
	ActionDelete               ActionKind = "delete"
	ActionReplaceDestroyCreate ActionKind = "replace_destroy_create"
	ActionReplaceCreateDestroy ActionKind = "replace_create_destroy"
	ActionRead                 ActionKind = "read"
	ActionUnknown              ActionKind = "unknown"
)

type Action struct {
	Raw      []string   `json:"raw"`
	Semantic ActionKind `json:"semantic"`
}

func NormalizeAction(actions []string) Action {
	raw := append([]string(nil), actions...)
	semantic := ActionUnknown
	switch {
	case len(actions) == 0 || (len(actions) == 1 && actions[0] == "no-op"):
		semantic = ActionNoOp
	case len(actions) == 1 && actions[0] == "create":
		semantic = ActionCreate
	case len(actions) == 1 && actions[0] == "update":
		semantic = ActionUpdate
	case len(actions) == 1 && actions[0] == "delete":
		semantic = ActionDelete
	case len(actions) == 1 && actions[0] == "read":
		semantic = ActionRead
	case len(actions) == 2 && actions[0] == "delete" && actions[1] == "create":
		semantic = ActionReplaceDestroyCreate
	case len(actions) == 2 && actions[0] == "create" && actions[1] == "delete":
		semantic = ActionReplaceCreateDestroy
	}
	return Action{Raw: raw, Semantic: semantic}
}

func (a Action) IsReplace() bool {
	return a.Semantic == ActionReplaceDestroyCreate || a.Semantic == ActionReplaceCreateDestroy
}

type PathStep struct {
	Attribute *string `json:"attribute,omitempty"`
	Index     *string `json:"index,omitempty"`
}

type AttributePath []PathStep

func Attribute(name string) PathStep { return PathStep{Attribute: &name} }
func Index(key string) PathStep      { return PathStep{Index: &key} }

func (p AttributePath) String() string {
	var b strings.Builder
	for _, step := range p {
		switch {
		case step.Attribute != nil:
			if b.Len() > 0 {
				b.WriteByte('.')
			}
			b.WriteString(*step.Attribute)
		case step.Index != nil:
			b.WriteByte('[')
			b.WriteString(*step.Index)
			b.WriteByte(']')
		}
	}
	return b.String()
}

type SafeValue struct {
	Value    any  `json:"value,omitempty"`
	Redacted bool `json:"redacted,omitempty"`
}

type ImportInfo struct {
	ID      string `json:"id,omitempty"`
	Unknown bool   `json:"unknown,omitempty"`
}

type ResourceChange struct {
	Address         Address         `json:"address"`
	PreviousAddress *Address        `json:"previous_address,omitempty"`
	ModuleAddress   *Address        `json:"module_address,omitempty"`
	Mode            ResourceMode    `json:"mode"`
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	Index           string          `json:"index,omitempty"`
	DeposedKey      string          `json:"deposed_key,omitempty"`
	Action          Action          `json:"action"`
	ActionReason    string          `json:"action_reason,omitempty"`
	ReplacePaths    []AttributePath `json:"replace_paths,omitempty"`
	Import          *ImportInfo     `json:"import,omitempty"`
	Before          SafeValue       `json:"before"`
	After           SafeValue       `json:"after"`
	UnknownPaths    []AttributePath `json:"unknown_paths,omitempty"`
	SensitivePaths  []AttributePath `json:"sensitive_paths,omitempty"`
}

type OutputChange struct {
	Name           string          `json:"name"`
	Action         Action          `json:"action"`
	Before         SafeValue       `json:"before"`
	After          SafeValue       `json:"after"`
	UnknownPaths   []AttributePath `json:"unknown_paths,omitempty"`
	SensitivePaths []AttributePath `json:"sensitive_paths,omitempty"`
}

type CheckStatus string

const (
	CheckPass    CheckStatus = "pass"
	CheckFail    CheckStatus = "fail"
	CheckError   CheckStatus = "error"
	CheckUnknown CheckStatus = "unknown"
)

type CheckProblem struct {
	Message string `json:"message,omitempty"`
}

type CheckResult struct {
	Address  string         `json:"address"`
	Kind     string         `json:"kind,omitempty"`
	Status   CheckStatus    `json:"status"`
	Problems []CheckProblem `json:"problems,omitempty"`
}

type RelevantAttribute struct {
	Resource Address       `json:"resource"`
	Path     AttributePath `json:"path"`
}

type DriftChange struct {
	Resource ResourceChange      `json:"resource"`
	Relevant []RelevantAttribute `json:"relevant,omitempty"`
}

type EvidenceConfidence string

const (
	ConfidenceExact     EvidenceConfidence = "exact"
	ConfidenceStrong    EvidenceConfidence = "strong"
	ConfidenceHeuristic EvidenceConfidence = "heuristic"
	ConfidenceUnknown   EvidenceConfidence = "unknown"
)

type NodeID string

type NodeKind string

const (
	NodeKindResource NodeKind = "resource"
	NodeKindData     NodeKind = "data"
	NodeKindOutput   NodeKind = "output"
	NodeKindVariable NodeKind = "variable"
)

type Node struct {
	ID      NodeID   `json:"id"`
	Address Address  `json:"address"`
	Kind    NodeKind `json:"kind,omitempty"`
}

type EdgeKind string

const (
	EdgeExplicitDependsOn EdgeKind = "explicit_depends_on"
	EdgeExpressionRef     EdgeKind = "expression_reference"
	EdgeVariableReference EdgeKind = "variable_reference"
	EdgeModuleInput       EdgeKind = "module_input"
	EdgeModuleOutput      EdgeKind = "module_output"
	EdgeOutputReference   EdgeKind = "output_reference"
	EdgeConservative      EdgeKind = "conservative_module_propagation"
)

type Edge struct {
	From       NodeID             `json:"from"`
	To         NodeID             `json:"to"`
	Kind       EdgeKind           `json:"kind"`
	Confidence EvidenceConfidence `json:"confidence"`
}

type DependencyGraph struct {
	Nodes []Node `json:"nodes,omitempty"`
	Edges []Edge `json:"edges,omitempty"`
}

// DirectDependents preserves the v1 change-intelligence meaning of blast radius:
// changed resources/data sources count as dependents; renderer-only variable and
// output nodes remain available through Edges but do not inflate risk metrics.
func (g DependencyGraph) DirectDependents(node NodeID) []NodeID {
	seen := map[NodeID]struct{}{}
	for _, edge := range g.Edges {
		if edge.From == node && g.isChangeNode(edge.To) {
			seen[edge.To] = struct{}{}
		}
	}
	return sortedNodeIDs(seen)
}

func (g DependencyGraph) TransitiveDependents(node NodeID) []NodeID {
	seen := map[NodeID]struct{}{}
	queue := []NodeID{node}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, next := range g.DirectDependents(current) {
			if next == node {
				continue
			}
			if _, ok := seen[next]; ok {
				continue
			}
			seen[next] = struct{}{}
			queue = append(queue, next)
		}
	}
	return sortedNodeIDs(seen)
}

func (g DependencyGraph) NodeKind(id NodeID) NodeKind {
	for _, node := range g.Nodes {
		if node.ID == id {
			return node.Kind
		}
	}
	return ""
}

func (g DependencyGraph) isChangeNode(id NodeID) bool {
	switch g.NodeKind(id) {
	case NodeKindOutput, NodeKindVariable:
		return false
	default:
		// Empty kind preserves compatibility with ChangeSet fixtures/reports from
		// before typed graph nodes were introduced.
		return true
	}
}

func sortedNodeIDs(set map[NodeID]struct{}) []NodeID {
	out := make([]NodeID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

type RedactionMode string

const (
	RedactionStandard RedactionMode = "standard"
	RedactionStrict   RedactionMode = "strict"
)

type RedactionSummary struct {
	Mode                    RedactionMode `json:"mode"`
	TerraformSensitivePaths int           `json:"terraform_sensitive_paths"`
	VariableValuesRemoved   int           `json:"variable_values_removed"`
	StrictValuesRemoved     int           `json:"strict_values_removed,omitempty"`
}

type ChangeSet struct {
	SchemaVersion string              `json:"schema_version"`
	Source        SourceMetadata      `json:"source"`
	Plan          PlanMetadata        `json:"plan"`
	Resources     []ResourceChange    `json:"resources,omitempty"`
	Outputs       []OutputChange      `json:"outputs,omitempty"`
	Checks        []CheckResult       `json:"checks,omitempty"`
	Drift         []DriftChange       `json:"drift,omitempty"`
	Relevant      []RelevantAttribute `json:"relevant,omitempty"`
	Graph         DependencyGraph     `json:"graph"`
	Redaction     RedactionSummary    `json:"redaction"`
}

func (c *ChangeSet) Sort() {
	sort.Slice(c.Resources, func(i, j int) bool { return c.Resources[i].Address < c.Resources[j].Address })
	sort.Slice(c.Outputs, func(i, j int) bool { return c.Outputs[i].Name < c.Outputs[j].Name })
	sort.Slice(c.Checks, func(i, j int) bool { return c.Checks[i].Address < c.Checks[j].Address })
	sort.Slice(c.Drift, func(i, j int) bool { return c.Drift[i].Resource.Address < c.Drift[j].Resource.Address })
	sort.Slice(c.Graph.Nodes, func(i, j int) bool { return c.Graph.Nodes[i].ID < c.Graph.Nodes[j].ID })
	sort.Slice(c.Graph.Edges, func(i, j int) bool {
		a, b := c.Graph.Edges[i], c.Graph.Edges[j]
		return fmt.Sprintf("%s:%s:%s", a.From, a.To, a.Kind) < fmt.Sprintf("%s:%s:%s", b.From, b.To, b.Kind)
	})
}
