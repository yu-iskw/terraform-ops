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

// MermaidGenerator implements the core.GraphGenerator interface for Mermaid format
type MermaidGenerator struct{}

// NewMermaidGenerator creates a new Mermaid generator
func NewMermaidGenerator() *MermaidGenerator {
	return &MermaidGenerator{}
}

// Generate generates a Mermaid format graph
func (g *MermaidGenerator) Generate(graphData *core.GraphData, opts core.GraphOptions) (string, error) {
	var builder strings.Builder

	// Add Mermaid theme configuration for Terraform colors
	builder.WriteString("---\n")
	builder.WriteString("theme: base\n")
	builder.WriteString("themeVariables:\n")
	builder.WriteString("  primaryColor: '#e8f5e8'\n")         // Light green for resources
	builder.WriteString("  primaryTextColor: '#2d5016'\n")     // Dark green text
	builder.WriteString("  primaryBorderColor: '#4caf50'\n")   // Green border
	builder.WriteString("  secondaryColor: '#fff3cd'\n")       // Light yellow for updates
	builder.WriteString("  secondaryTextColor: '#856404'\n")   // Dark yellow text
	builder.WriteString("  secondaryBorderColor: '#ffc107'\n") // Yellow border
	builder.WriteString("  tertiaryColor: '#f8d7da'\n")        // Light red for deletes
	builder.WriteString("  tertiaryTextColor: '#721c24'\n")    // Dark red text
	builder.WriteString("  tertiaryBorderColor: '#dc3545'\n")  // Red border
	builder.WriteString("  noteBkgColor: '#fff5ad'\n")         // Light yellow for notes
	builder.WriteString("  noteTextColor: '#333'\n")           // Dark text for notes
	builder.WriteString("  lineColor: '#666'\n")               // Gray lines
	builder.WriteString("  textColor: '#333'\n")               // Dark text
	builder.WriteString("  mainBkg: '#f8f9fa'\n")              // Light background
	builder.WriteString("---\n\n")

	// Start the graph
	builder.WriteString("graph TB\n")

	// Collect used CSS classes
	usedClasses := make(map[string]bool)

	// Group nodes by module
	moduleGroups := groupNodesByModule(graphData.Nodes)

	for moduleName, nodes := range moduleGroups {
		if moduleName == "" {
			moduleName = "root"
		}

		// Define subgraph
		builder.WriteString(fmt.Sprintf("  subgraph %s[\"%s\"]\n",
			sanitizeID(moduleName), moduleName))

		for _, node := range nodes {
			actionType := getActionType(node.Actions)

			// Use simple single-line labels to avoid parsing issues
			label := fmt.Sprintf("%s [%s]", node.Address, actionType)

			// Get color based on action type for resources, or node type for others
			var color string
			if isResourceType(node.Type) {
				color = getMermaidActionColor(actionType)
			} else {
				color = getMermaidNodeTypeColor(node.Type)
			}

			// Track used classes
			usedClasses[color] = true

			// Get shape for the node
			shape := getNodeShape(node.Type, node.Type)
			mermaidShape := getMermaidShape(shape)

			// Define nodes with proper Mermaid syntax using resource type-specific shapes
			builder.WriteString(fmt.Sprintf("    %s"+mermaidShape+"\n", node.ID, label))
		}

		builder.WriteString("  end\n\n")
	}

	// Add edges
	for _, edge := range graphData.Edges {
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
	}

	// Add CSS class definitions
	if len(usedClasses) > 0 {
		builder.WriteString("\n")
		for class := range usedClasses {
			switch class {
			case "create":
				builder.WriteString("classDef create fill:#d4edda,stroke:#c3e6cb,stroke-width:2px,color:#155724\n")
			case "update":
				builder.WriteString("classDef update fill:#fff3cd,stroke:#ffeaa7,stroke-width:2px,color:#856404\n")
			case "delete":
				builder.WriteString("classDef delete fill:#f8d7da,stroke:#f5c6cb,stroke-width:2px,color:#721c24\n")
			case "replace":
				builder.WriteString("classDef replace fill:#fde2e2,stroke:#fecaca,stroke-width:2px,color:#991b1b\n")
			case "noop":
				builder.WriteString("classDef noop fill:#e9ecef,stroke:#dee2e6,stroke-width:2px,color:#495057\n")
			case "default":
				builder.WriteString("classDef default fill:#f8f9fa,stroke:#dee2e6,stroke-width:2px,color:#495057\n")
			case "resource":
				builder.WriteString("classDef resource fill:#d4edda,stroke:#c3e6cb,stroke-width:2px,color:#155724\n")
			case "datasource":
				builder.WriteString("classDef datasource fill:#d1ecf1,stroke:#bee5eb,stroke-width:2px,color:#0c5460\n")
			case "output":
				builder.WriteString("classDef output fill:#cce5ff,stroke:#b3d9ff,stroke-width:2px,color:#004085\n")
			case "variable":
				builder.WriteString("classDef variable fill:#fff3cd,stroke:#ffeaa7,stroke-width:2px,color:#856404\n")
			case "local":
				builder.WriteString("classDef local fill:#f8d7da,stroke:#f5c6cb,stroke-width:2px,color:#721c24\n")
			}
		}

		// Apply CSS classes to nodes
		builder.WriteString("\n")
		for _, node := range graphData.Nodes {
			actionType := getActionType(node.Actions)
			var color string
			if isResourceType(node.Type) {
				color = getMermaidActionColor(actionType)
			} else {
				color = getMermaidNodeTypeColor(node.Type)
			}
			builder.WriteString(fmt.Sprintf("class %s %s\n", node.ID, color))
		}
	}

	return builder.String(), nil
}

// Helper functions
func getMermaidActionColor(actionType core.ActionType) string {
	switch actionType {
	case core.ActionCreate:
		return "create"
	case core.ActionUpdate:
		return "update"
	case core.ActionDelete:
		return "delete"
	case core.ActionReplace:
		return "replace"
	case core.ActionNoOp:
		return "noop"
	default:
		return "default"
	}
}

func getMermaidNodeTypeColor(nodeType string) string {
	switch nodeType {
	case string(core.NodeTypeResource):
		return "resource"
	case string(core.NodeTypeData):
		return "datasource"
	case string(core.NodeTypeOutput):
		return "output"
	case string(core.NodeTypeVariable):
		return "variable"
	case string(core.NodeTypeLocal):
		return "local"
	default:
		return "default"
	}
}

func getMermaidShape(shape string) string {
	switch shape {
	case "box":
		return "[\"%s\"]" // Rectangle
	case "house":
		return "[\"%s\"]" // House (using rectangle as approximation in Mermaid)
	case "diamond":
		return "{\"%s\"}" // Rhombus/Diamond
	case "invhouse":
		return "[\"%s\"]" // Inverted house (using rectangle as approximation)
	case "ellipse":
		return "((\"%s\"))" // Circle
	case "cylinder":
		return "[\"%s\"]" // Cylinder (using rectangle as approximation)
	case "parallelogram":
		return "[/\"%s\"/]" // Parallelogram
	case "hexagon":
		return "{{\"%s\"}}" // Hexagon
	case "octagon":
		return "{{{\"%s\"}}}" // Octagon (using triple braces as approximation)
	default:
		return "[\"%s\"]" // Default rectangle
	}
}
