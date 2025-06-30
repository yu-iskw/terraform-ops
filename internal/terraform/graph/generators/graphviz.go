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

			// Use action color for resources, node type color for others
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
