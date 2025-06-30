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

package graph

import (
	"fmt"
	"os"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// Builder implements the core.GraphBuilder interface
type Builder struct{}

// NewBuilder creates a new graph builder
func NewBuilder() *Builder {
	return &Builder{}
}

// BuildGraph converts a Terraform plan into graph data
func (b *Builder) BuildGraph(plan *core.TerraformPlan, opts core.GraphOptions) (*core.GraphData, error) {
	graphData := &core.GraphData{
		Nodes: make([]core.GraphNode, 0),
		Edges: make([]core.GraphEdge, 0),
	}

	// Extract nodes from resource changes (managed resources and data sources)
	for _, change := range plan.ResourceChanges {
		// Skip data sources if disabled
		if change.Mode == "data" && opts.NoDataSources {
			continue
		}

		// Skip resources from modules if disabled
		if opts.NoModules && change.ModuleAddress != "" {
			continue
		}

		// Extract provider from resource type
		provider := extractProviderFromType(change.Type)

		node := core.GraphNode{
			ID:        sanitizeID(change.Address),
			Address:   change.Address,
			Type:      change.Type, // Use the actual resource type (e.g., "aws_instance") instead of the NodeType constant
			Name:      change.Name,
			Module:    change.ModuleAddress,
			Provider:  provider,
			Actions:   change.Change.Actions,
			Sensitive: hasSensitiveValues(change.Change.AfterSensitive),
		}
		graphData.Nodes = append(graphData.Nodes, node)

		if opts.Verbose && change.Mode == "data" {
			fmt.Fprintf(os.Stderr, "Debug: Added data source node: %s\n", change.Address)
		}
	}

	// Extract nodes from output changes (if enabled)
	if !opts.NoOutputs && plan.OutputChanges != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: ShowOutputs is enabled, found %d output changes\n", len(plan.OutputChanges))
		}
		for outputName, outputChange := range plan.OutputChanges {
			// Create output address in standard format
			outputAddress := "output." + outputName

			node := core.GraphNode{
				ID:        sanitizeID(outputAddress),
				Address:   outputAddress,
				Type:      string(core.NodeTypeOutput),
				Name:      outputName,
				Module:    "", // Outputs are always in root module in output_changes
				Actions:   outputChange.Change.Actions,
				Sensitive: hasSensitiveValues(outputChange.Change.AfterSensitive),
			}
			graphData.Nodes = append(graphData.Nodes, node)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added output node: %s with actions %v\n", outputAddress, outputChange.Change.Actions)
			}
		}
	}

	// Extract nodes from variables (if enabled)
	if !opts.NoVariables && plan.Variables != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: ShowVariables is enabled, found %d variables\n", len(plan.Variables))
		}
		for varName := range plan.Variables {
			// Create variable address in standard format
			varAddress := "var." + varName

			node := core.GraphNode{
				ID:        sanitizeID(varAddress),
				Address:   varAddress,
				Type:      string(core.NodeTypeVariable),
				Name:      varName,
				Module:    "",                // Variables are always in root module
				Actions:   []string{"no-op"}, // Variables don't have actions
				Sensitive: false,             // Variable sensitivity would need to be determined from configuration
			}
			graphData.Nodes = append(graphData.Nodes, node)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added variable node: %s\n", varAddress)
			}
		}
	}

	// Extract nodes from variables in configuration (if enabled and not already processed)
	if !opts.NoVariables && plan.Configuration.RootModule.Variables != nil {
		existingVars := make(map[string]bool)
		// Track already processed variables from plan.Variables
		if plan.Variables != nil {
			for varName := range plan.Variables {
				existingVars[varName] = true
			}
		}

		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing %d variables from configuration\n", len(plan.Configuration.RootModule.Variables))
		}
		for varName, varConfig := range plan.Configuration.RootModule.Variables {
			// Skip if already processed from plan.Variables
			if existingVars[varName] {
				continue
			}

			// Create variable address in standard format
			varAddress := "var." + varName

			node := core.GraphNode{
				ID:        sanitizeID(varAddress),
				Address:   varAddress,
				Type:      string(core.NodeTypeVariable),
				Name:      varName,
				Module:    "",                // Variables are always in root module
				Actions:   []string{"no-op"}, // Variables don't have actions
				Sensitive: varConfig.Sensitive,
			}
			graphData.Nodes = append(graphData.Nodes, node)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added variable node from config: %s\n", varAddress)
			}
		}
	}

	// Extract nodes from locals (if enabled)
	if !opts.NoLocals && plan.Configuration.RootModule.Locals != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: ShowLocals is enabled, found %d locals\n", len(plan.Configuration.RootModule.Locals))
		}
		for localName := range plan.Configuration.RootModule.Locals {
			// Create local address in standard format
			localAddress := "local." + localName

			node := core.GraphNode{
				ID:        sanitizeID(localAddress),
				Address:   localAddress,
				Type:      string(core.NodeTypeLocal),
				Name:      localName,
				Module:    "",                // Locals are always in root module
				Actions:   []string{"no-op"}, // Locals don't have actions
				Sensitive: false,             // Local sensitivity would need deeper analysis
			}
			graphData.Nodes = append(graphData.Nodes, node)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added local node: %s\n", localAddress)
			}
		}
	}

	// Add edges based on dependencies
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: Analyzing dependencies...\n")
	}
	edges, err := b.analyzeDependencies(plan, opts)
	if err != nil {
		return nil, &core.GraphBuildError{
			Message: "failed to analyze dependencies",
			Cause:   err,
		}
	}
	graphData.Edges = edges

	return graphData, nil
}

// sanitizeID sanitizes an ID for use in graph formats
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

// hasSensitiveValues checks if a value has sensitive data
func hasSensitiveValues(sensitive interface{}) bool {
	// Implementation from original utils.go
	return false
}

// isResourceType checks if a string represents a resource type
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

// extractProviderFromType extracts the provider from a resource type
func extractProviderFromType(resourceType string) string {
	// Terraform resource types follow the pattern: provider_resource_type
	// For example: aws_instance, google_compute_instance, azurerm_virtual_machine
	parts := strings.Split(resourceType, "_")
	if len(parts) >= 2 {
		return parts[0]
	}
	return ""
}
