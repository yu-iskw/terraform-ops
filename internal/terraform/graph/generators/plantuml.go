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

// PlantUMLGenerator implements the core.GraphGenerator interface for PlantUML format
type PlantUMLGenerator struct{}

// NewPlantUMLGenerator creates a new PlantUML generator
func NewPlantUMLGenerator() *PlantUMLGenerator {
	return &PlantUMLGenerator{}
}

// Generate generates a PlantUML format graph
func (g *PlantUMLGenerator) Generate(graphData *core.GraphData, opts core.GraphOptions) (string, error) {
	var builder strings.Builder

	builder.WriteString("@startuml\n")
	builder.WriteString("!theme plain\n")
	builder.WriteString("skinparam backgroundColor white\n")
	builder.WriteString("skinparam defaultFontName Arial\n\n")

	// Define colors for different action types
	builder.WriteString("!define CREATE_COLOR #d4edda\n")
	builder.WriteString("!define UPDATE_COLOR #fff3cd\n")
	builder.WriteString("!define DELETE_COLOR #f8d7da\n")
	builder.WriteString("!define REPLACE_COLOR #fde2e2\n")
	builder.WriteString("!define NOOP_COLOR #e9ecef\n")
	builder.WriteString("!define RESOURCE_COLOR #d4edda\n")
	builder.WriteString("!define DATASOURCE_COLOR #d1ecf1\n")
	builder.WriteString("!define OUTPUT_COLOR #cce5ff\n")
	builder.WriteString("!define VARIABLE_COLOR #fff3cd\n")
	builder.WriteString("!define LOCAL_COLOR #f8d7da\n\n")

	// Group nodes by module
	moduleGroups := groupNodesByModule(graphData.Nodes)

	for moduleName, nodes := range moduleGroups {
		if moduleName == "" {
			moduleName = "Root Module"
		}

		builder.WriteString(fmt.Sprintf("package \"%s\" {\n", moduleName))

		for _, node := range nodes {
			actionType := getActionType(node.Actions)
			label := fmt.Sprintf("%s\\n[%s]", node.Address, actionType)

			// Get color based on action type for resources, or node type for others
			var color string
			if isResourceType(node.Type) {
				color = getPlantUMLActionColor(actionType)
			} else {
				color = getPlantUMLNodeTypeColor(node.Type)
			}

			// Get shape for the node
			shape := getNodeShape(node.Type, node.Type)

			// Use different notation for each node type in PlantUML based on shape
			switch shape {
			case "box":
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // Rectangle
			case "house":
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // House (using rectangle as approximation)
			case "diamond":
				builder.WriteString(fmt.Sprintf("  <%s> as %s #%s\n", label, node.ID, color)) // Rhombus/Diamond
			case "invhouse":
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // Inverted house (using rectangle as approximation)
			case "ellipse":
				builder.WriteString(fmt.Sprintf("  (%s) as %s #%s\n", label, node.ID, color)) // Circle
			case "cylinder":
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // Cylinder (using rectangle as approximation)
			case "parallelogram":
				builder.WriteString(fmt.Sprintf("  \"%s\" as %s #%s\n", label, node.ID, color)) // Parallelogram
			case "hexagon":
				builder.WriteString(fmt.Sprintf("  {%s} as %s #%s\n", label, node.ID, color)) // Hexagon
			case "octagon":
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // Octagon (using rectangle as approximation)
			default:
				builder.WriteString(fmt.Sprintf("  [%s] as %s #%s\n", label, node.ID, color)) // Default rectangle
			}
		}

		builder.WriteString("}\n\n")
	}

	// Add edges
	for _, edge := range graphData.Edges {
		builder.WriteString(fmt.Sprintf("%s --> %s\n", edge.From, edge.To))
	}

	builder.WriteString("@enduml\n")
	return builder.String(), nil
}

// Helper functions
func getPlantUMLActionColor(actionType core.ActionType) string {
	switch actionType {
	case core.ActionCreate:
		return "CREATE_COLOR"
	case core.ActionUpdate:
		return "UPDATE_COLOR"
	case core.ActionDelete:
		return "DELETE_COLOR"
	case core.ActionReplace:
		return "REPLACE_COLOR"
	case core.ActionNoOp:
		return "NOOP_COLOR"
	default:
		return "NOOP_COLOR"
	}
}

func getPlantUMLNodeTypeColor(nodeType string) string {
	switch nodeType {
	case string(core.NodeTypeResource):
		return "RESOURCE_COLOR"
	case string(core.NodeTypeData):
		return "DATASOURCE_COLOR"
	case string(core.NodeTypeOutput):
		return "OUTPUT_COLOR"
	case string(core.NodeTypeVariable):
		return "VARIABLE_COLOR"
	case string(core.NodeTypeLocal):
		return "LOCAL_COLOR"
	default:
		return "NOOP_COLOR"
	}
}
