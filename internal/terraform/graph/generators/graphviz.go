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

package generators

import (
	"fmt"
	"sort"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// GraphvizGenerator implements the core.GraphGenerator interface for Graphviz format
type GraphvizGenerator struct{}

// NewGraphvizGenerator creates a new Graphviz generator
func NewGraphvizGenerator() *GraphvizGenerator {
	return &GraphvizGenerator{}
}

// Generate generates a Graphviz DOT format graph
func (g *GraphvizGenerator) Generate(graphData *core.GraphData, opts core.GraphOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("digraph terraform_plan {\n")
	builder.WriteString("  rankdir=TB;\n")
	builder.WriteString("  node [shape=box, style=filled, fontname=\"Arial\"];\n")
	builder.WriteString("  edge [fontname=\"Arial\"];\n\n")

	// Group nodes by module
	moduleGroups := groupNodesByModule(graphData.Nodes)

	for moduleName, nodes := range moduleGroups {
		if moduleName == "" {
			moduleName = "Root Module"
		}

		builder.WriteString(fmt.Sprintf("  subgraph cluster_%s {\n", sanitizeID(moduleName)))
		builder.WriteString(fmt.Sprintf("    label=\"%s\";\n", moduleName))
		builder.WriteString("    style=filled;\n")
		builder.WriteString("    color=lightgrey;\n\n")

		for _, node := range nodes {
			actionType := getActionType(node.Actions)

			// Use node type color and shape, fall back to action color for resources
			var color string
			if isResourceType(node.Type) {
				color = getActionColor(actionType)
			} else {
				color = getNodeTypeColor(node.Type)
			}

			shape := getNodeShape(node.Type, node.Type)
			label := fmt.Sprintf("%s\\n[%s]", node.Address, actionType)

			builder.WriteString(fmt.Sprintf("    %s [label=\"%s\", fillcolor=%s, shape=%s];\n",
				node.ID, label, color, shape))
		}

		builder.WriteString("  }\n\n")
	}

	// Add edges
	for _, edge := range graphData.Edges {
		builder.WriteString(fmt.Sprintf("  %s -> %s;\n", edge.From, edge.To))
	}

	builder.WriteString("}\n")
	return builder.String(), nil
}

// Helper functions
func sanitizeID(id string) string {
	// Replace special characters that might cause issues in graph formats
	replacements := map[string]string{
		".": "_",
		"-": "_",
		"[": "_",
		"]": "_",
		"(": "_",
		")": "_",
		" ": "_",
	}

	result := id
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

func groupNodesByModule(nodes []core.GraphNode) map[string][]core.GraphNode {
	groups := make(map[string][]core.GraphNode)
	for _, node := range nodes {
		module := node.Module
		if module == "" {
			module = "root"
		}
		groups[module] = append(groups[module], node)
	}
	return groups
}

func getActionType(actions []string) core.ActionType {
	if len(actions) == 0 {
		return core.ActionNoOp
	}

	// Sort actions for consistent comparison
	sortedActions := make([]string, len(actions))
	copy(sortedActions, actions)
	sort.Strings(sortedActions)

	actionStr := strings.Join(sortedActions, ",")

	switch actionStr {
	case "create":
		return core.ActionCreate
	case "update":
		return core.ActionUpdate
	case "delete":
		return core.ActionDelete
	case "create,delete", "delete,create":
		return core.ActionReplace
	case "no-op":
		return core.ActionNoOp
	default:
		return core.ActionNoOp
	}
}

func getActionColor(actionType core.ActionType) string {
	switch actionType {
	case core.ActionCreate:
		return "lightgreen"
	case core.ActionUpdate:
		return "lightyellow"
	case core.ActionDelete:
		return "lightcoral"
	case core.ActionReplace:
		return "orange"
	case core.ActionNoOp:
		return "lightgrey"
	default:
		return "white"
	}
}

func getNodeTypeColor(nodeType string) string {
	switch nodeType {
	case string(core.NodeTypeResource):
		return "lightblue"
	case string(core.NodeTypeData):
		return "lightcyan"
	case string(core.NodeTypeOutput):
		return "lightsteelblue"
	case string(core.NodeTypeVariable):
		return "lightyellow"
	case string(core.NodeTypeLocal):
		return "lightpink"
	default:
		return "white"
	}
}

func getNodeShape(nodeType string, resourceType string) string {
	switch nodeType {
	case string(core.NodeTypeResource):
		return "house" // House shape for infrastructure resources
	case string(core.NodeTypeData):
		return "diamond" // Diamond for data sources (keep existing)
	case string(core.NodeTypeOutput):
		return "invhouse" // Inverted house for outputs/exports
	case string(core.NodeTypeVariable):
		return "cylinder" // Cylinder for input variables
	case string(core.NodeTypeLocal):
		return "octagon" // Octagon for computed locals
	default:
		// Check if this looks like a resource type (has underscore, like "aws_instance")
		if strings.Contains(nodeType, "_") {
			return "house" // House shape for infrastructure resources
		}
		return "box" // Default fallback
	}
}

func isResourceType(s string) bool {
	// Terraform resource types follow the pattern: provider_resource_type
	// Since we're parsing from a valid Terraform plan, any resource type
	// that follows this pattern should be considered valid
	parts := strings.Split(s, "_")
	if len(parts) < 2 {
		return false
	}

	// Check if it looks like a valid resource type pattern
	// This is a more flexible approach that accepts any provider
	return true
}
