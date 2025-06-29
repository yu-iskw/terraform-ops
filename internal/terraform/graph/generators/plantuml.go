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

			// Get shape for the node
			shape := getNodeShape(node.Type, node.Type)

			// Use different notation for each node type in PlantUML based on shape
			switch shape {
			case "box":
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Rectangle
			case "house":
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // House (using rectangle as approximation)
			case "diamond":
				builder.WriteString(fmt.Sprintf("  <%s> as %s\n", label, node.ID)) // Rhombus/Diamond
			case "invhouse":
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Inverted house (using rectangle as approximation)
			case "ellipse":
				builder.WriteString(fmt.Sprintf("  (%s) as %s\n", label, node.ID)) // Circle
			case "cylinder":
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Cylinder (using rectangle as approximation)
			case "parallelogram":
				builder.WriteString(fmt.Sprintf("  \"%s\" as %s\n", label, node.ID)) // Parallelogram
			case "hexagon":
				builder.WriteString(fmt.Sprintf("  {%s} as %s\n", label, node.ID)) // Hexagon
			case "octagon":
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Octagon (using rectangle as approximation)
			default:
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Default rectangle
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
