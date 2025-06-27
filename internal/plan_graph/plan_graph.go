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
	"fmt"
	"os"
	"sort"
	"strings"
)

// GraphNode represents a node in the graph
type GraphNode struct {
	ID        string
	Address   string
	Type      string
	Name      string
	Module    string
	Actions   []string
	Sensitive bool
}

// GraphEdge represents an edge in the graph
type GraphEdge struct {
	From string
	To   string
}

// GraphData represents the complete graph structure
type GraphData struct {
	Nodes []GraphNode
	Edges []GraphEdge
}

// NodeType represents the type of a node in the graph
type NodeType string

const (
	NodeTypeResource NodeType = "resource"
	NodeTypeData     NodeType = "data"
	NodeTypeOutput   NodeType = "output"
	NodeTypeVariable NodeType = "variable"
	NodeTypeLocal    NodeType = "local"
)

// ActionType represents the type of action for a resource
type ActionType string

const (
	ActionCreate  ActionType = "CREATE"
	ActionUpdate  ActionType = "UPDATE"
	ActionDelete  ActionType = "DELETE"
	ActionReplace ActionType = "REPLACE"
	ActionNoOp    ActionType = "NO-OP"
)

// GraphFormat represents the output format for the graph
type GraphFormat string

const (
	FormatGraphviz GraphFormat = "graphviz"
	FormatMermaid  GraphFormat = "mermaid"
	FormatPlantUML GraphFormat = "plantuml"
)

// GroupingStrategy represents how resources should be grouped
type GroupingStrategy string

const (
	GroupByModule       GroupingStrategy = "module"
	GroupByAction       GroupingStrategy = "action"
	GroupByResourceType GroupingStrategy = "resource_type"
)

// Options represents the command line options for plan-graph
type Options struct {
	Format        GraphFormat
	Output        string
	GroupBy       GroupingStrategy
	NoDataSources bool
	NoLocals      bool
	NoOutputs     bool
	NoVariables   bool
	Compact       bool
	Verbose       bool
}

// BuildGraphData converts a Terraform plan into graph data
func BuildGraphData(plan *TerraformPlan, opts Options) (*GraphData, error) {
	graphData := &GraphData{
		Nodes: make([]GraphNode, 0),
		Edges: make([]GraphEdge, 0),
	}

	// Extract nodes from resource changes (managed resources and data sources)
	for _, change := range plan.ResourceChanges {
		// Skip data sources if disabled
		if change.Mode == "data" && opts.NoDataSources {
			continue
		}

		node := GraphNode{
			ID:        sanitizeID(change.Address),
			Address:   change.Address,
			Type:      change.Type,
			Name:      change.Name,
			Module:    change.ModuleAddress,
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

			node := GraphNode{
				ID:        sanitizeID(outputAddress),
				Address:   outputAddress,
				Type:      string(NodeTypeOutput),
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

			node := GraphNode{
				ID:        sanitizeID(varAddress),
				Address:   varAddress,
				Type:      string(NodeTypeVariable),
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

			node := GraphNode{
				ID:        sanitizeID(varAddress),
				Address:   varAddress,
				Type:      string(NodeTypeVariable),
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

			node := GraphNode{
				ID:        sanitizeID(localAddress),
				Address:   localAddress,
				Type:      string(NodeTypeLocal),
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

	// Add edges based on dependencies (if enabled)
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: ShowDependencies is enabled, analyzing dependencies...\n")
	}
	edges, err := analyzeDependencies(plan, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze dependencies: %w", err)
	}
	graphData.Edges = edges

	return graphData, nil
}

// analyzeDependencies extracts dependency relationships from the plan
func analyzeDependencies(plan *TerraformPlan, opts Options) ([]GraphEdge, error) {
	var edges []GraphEdge

	// Create a map of resource addresses to their IDs for quick lookup
	resourceMap := make(map[string]string)
	for _, change := range plan.ResourceChanges {
		// Include both managed resources and data sources if enabled
		if change.Mode == "data" && opts.NoDataSources {
			continue
		}
		resourceMap[change.Address] = sanitizeID(change.Address)
	}

	// Add outputs to the resource map if enabled
	if !opts.NoOutputs && plan.OutputChanges != nil {
		for outputName := range plan.OutputChanges {
			outputAddress := "output." + outputName
			resourceMap[outputAddress] = sanitizeID(outputAddress)
		}
	}

	// Add variables to the resource map if enabled
	if !opts.NoVariables {
		if plan.Variables != nil {
			for varName := range plan.Variables {
				varAddress := "var." + varName
				resourceMap[varAddress] = sanitizeID(varAddress)
			}
		}
		if plan.Configuration.RootModule.Variables != nil {
			for varName := range plan.Configuration.RootModule.Variables {
				varAddress := "var." + varName
				resourceMap[varAddress] = sanitizeID(varAddress)
			}
		}
	}

	// Add locals to the resource map if enabled
	if !opts.NoLocals && plan.Configuration.RootModule.Locals != nil {
		for localName := range plan.Configuration.RootModule.Locals {
			localAddress := "local." + localName
			resourceMap[localAddress] = sanitizeID(localAddress)
		}
	}

	// Debug logging
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: Found %d resource changes\n", len(plan.ResourceChanges))
		fmt.Fprintf(os.Stderr, "Debug: Root module has %d resources\n", len(plan.Configuration.RootModule.Resources))
		fmt.Fprintf(os.Stderr, "Debug: Root module has %d module calls\n", len(plan.Configuration.RootModule.ModuleCalls))
	}

	// Extract dependencies from the configuration section
	if len(plan.Configuration.RootModule.Resources) > 0 || len(plan.Configuration.RootModule.ModuleCalls) > 0 {
		// Extract dependencies from root module resources
		for _, resource := range plan.Configuration.RootModule.Resources {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing root resource: %s\n", resource.Address)
			}
			resourceEdges := extractResourceDependencies(resource.Address, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from module calls
		for moduleName, moduleCall := range plan.Configuration.RootModule.ModuleCalls {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing module: %s\n", moduleName)
			}
			moduleEdges := extractModuleDependencies(moduleName, moduleCall, resourceMap, opts)
			edges = append(edges, moduleEdges...)
		}

		// Extract dependencies from output configurations (if enabled)
		if !opts.NoOutputs && plan.Configuration.RootModule.Outputs != nil {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing %d output configurations\n", len(plan.Configuration.RootModule.Outputs))
			}
			for outputName, outputConfig := range plan.Configuration.RootModule.Outputs {
				outputAddress := "output." + outputName
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Processing output: %s\n", outputAddress)
				}
				outputEdges := extractOutputDependencies(outputAddress, outputConfig, resourceMap, opts)
				edges = append(edges, outputEdges...)
			}
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: Generated %d edges\n", len(edges))
	}

	return edges, nil
}

// extractModuleDependencies extracts dependencies from a module call
func extractModuleDependencies(moduleName string, moduleCall ModuleCall, resourceMap map[string]string, opts Options) []GraphEdge {
	var edges []GraphEdge

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractModuleDependencies for module: %s\n", moduleName)
		fmt.Fprintf(os.Stderr, "Debug: Module has expressions: %v\n", moduleCall.Expressions != nil)
		fmt.Fprintf(os.Stderr, "Debug: Module has nested module: %v\n", moduleCall.Module != nil)
	}

	// Extract dependencies from module expressions
	if moduleCall.Expressions != nil {
		modulePrefix := "module." + moduleName
		implicitDeps := extractResourceReferences(moduleCall.Expressions, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d implicit dependencies in module expressions\n", len(implicitDeps))
		}
		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := resolveDependencyAddress(ref, "", resourceMap, opts)
			if resolvedDep == "" {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Could not resolve module dependency %s\n", ref)
				}
				continue
			}

			// Create edges from all resources in this module to the dependency
			for resourceAddr, resourceID := range resourceMap {
				if strings.HasPrefix(resourceAddr, modulePrefix) {
					edge := GraphEdge{
						From: resourceID,
						To:   sanitizeID(resolvedDep),
					}
					edges = append(edges, edge)
					if opts.Verbose {
						fmt.Fprintf(os.Stderr, "Debug: Added cross-module dependency edge: %s -> %s\n", resourceID, sanitizeID(resolvedDep))
					}
				}
			}
		}
	}

	// Recursively extract dependencies from nested modules
	if moduleCall.Module != nil {
		modulePrefix := "module." + moduleName
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing nested module resources for %s\n", moduleName)
			fmt.Fprintf(os.Stderr, "Debug: Module has %d resources\n", len(moduleCall.Module.Resources))
			fmt.Fprintf(os.Stderr, "Debug: Module has %d nested module calls\n", len(moduleCall.Module.ModuleCalls))
		}

		// Extract dependencies from nested module resources
		for _, resource := range moduleCall.Module.Resources {
			fullAddress := modulePrefix + "." + resource.Address
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing nested resource: %s\n", fullAddress)
			}
			resourceEdges := extractResourceDependencies(fullAddress, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from nested module calls
		for nestedModuleName, nestedModuleCall := range moduleCall.Module.ModuleCalls {
			nestedModulePrefix := modulePrefix + ".module." + nestedModuleName
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing nested module call: %s\n", nestedModulePrefix)
			}
			nestedEdges := extractNestedModuleDependencies(nestedModulePrefix, nestedModuleCall, resourceMap, opts)
			edges = append(edges, nestedEdges...)
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractModuleDependencies for %s generated %d edges\n", moduleName, len(edges))
	}

	return edges
}

// extractNestedModuleDependencies extracts dependencies from nested modules
func extractNestedModuleDependencies(modulePrefix string, moduleCall ModuleCall, resourceMap map[string]string, opts Options) []GraphEdge {
	var edges []GraphEdge

	// Extract dependencies from module expressions
	if moduleCall.Expressions != nil {
		implicitDeps := extractResourceReferences(moduleCall.Expressions, opts)
		for _, ref := range implicitDeps {
			// Check if the dependency resource exists in our resource changes
			if _, exists := resourceMap[ref]; !exists {
				continue
			}

			// Create edges from all resources in this module to the dependency
			for resourceAddr, resourceID := range resourceMap {
				if strings.HasPrefix(resourceAddr, modulePrefix) {
					edge := GraphEdge{
						From: resourceID,
						To:   sanitizeID(ref),
					}
					edges = append(edges, edge)
				}
			}
		}
	}

	// Recursively extract dependencies from nested modules
	if moduleCall.Module != nil {
		// Extract dependencies from nested module resources
		for _, resource := range moduleCall.Module.Resources {
			fullAddress := modulePrefix + "." + resource.Address
			resourceEdges := extractResourceDependencies(fullAddress, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from nested module calls
		for nestedModuleName, nestedModuleCall := range moduleCall.Module.ModuleCalls {
			nestedModulePrefix := modulePrefix + ".module." + nestedModuleName
			nestedEdges := extractNestedModuleDependencies(nestedModulePrefix, nestedModuleCall, resourceMap, opts)
			edges = append(edges, nestedEdges...)
		}
	}

	return edges
}

// extractResourceDependencies extracts dependencies for a specific resource
func extractResourceDependencies(resourceAddress string, resource ConfigurationResource, resourceMap map[string]string, opts Options) []GraphEdge {
	var edges []GraphEdge
	fromID := sanitizeID(resourceAddress)

	// Use a map to track unique edges and avoid duplicates
	edgeMap := make(map[string]bool)

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractResourceDependencies for resource: %s\n", resourceAddress)
		fmt.Fprintf(os.Stderr, "Debug: Resource exists in changes: %v\n", resourceMap[resourceAddress] != "")
		fmt.Fprintf(os.Stderr, "Debug: Resource has expressions: %v\n", resource.Expressions != nil)
		fmt.Fprintf(os.Stderr, "Debug: Resource has depends_on: %v\n", resource.DependsOn != nil)
	}

	// Skip if this resource is not in the changes
	if _, exists := resourceMap[resourceAddress]; !exists {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Skipping resource %s (not in resource changes)\n", resourceAddress)
		}
		return edges // Skip if this resource is not in the changes
	}

	// Get the module prefix for this resource to resolve local references
	modulePrefix := ""
	if strings.Contains(resourceAddress, "module.") {
		// Extract module prefix from resource address
		// e.g., "module.app.module.database.google_sql_database.app" -> "module.app.module.database"
		parts := strings.Split(resourceAddress, ".")
		var moduleParts []string

		for i := 0; i < len(parts); i++ {
			if parts[i] == "module" && i+1 < len(parts) {
				// Add "module.name" to the prefix
				moduleParts = append(moduleParts, parts[i], parts[i+1])
				i++ // Skip the module name since we just processed it
			} else if len(moduleParts) > 0 && !isResourceType(parts[i]) {
				// If we already have module parts and this isn't a resource type, it might be another module
				break
			} else if len(moduleParts) > 0 {
				// We've hit a resource type, so we're done with the module prefix
				break
			}
		}

		if len(moduleParts) > 0 {
			modulePrefix = strings.Join(moduleParts, ".")
		}
	}

	// Helper function to add edge if it doesn't exist
	addEdge := func(toAddress string, edgeType string) {
		edgeKey := fromID + "->" + sanitizeID(toAddress)
		if !edgeMap[edgeKey] {
			edge := GraphEdge{
				From: fromID,
				To:   sanitizeID(toAddress),
			}
			edges = append(edges, edge)
			edgeMap[edgeKey] = true
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added %s dependency edge: %s -> %s\n", edgeType, fromID, sanitizeID(toAddress))
			}
		} else {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Skipped duplicate edge: %s -> %s\n", fromID, sanitizeID(toAddress))
			}
		}
	}

	// Extract explicit dependencies from depends_on
	if resource.DependsOn != nil {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing %d explicit dependencies\n", len(resource.DependsOn))
		}
		for _, dependsOn := range resource.DependsOn {
			// Resolve the dependency address
			resolvedDep := resolveDependencyAddress(dependsOn, modulePrefix, resourceMap, opts)
			if resolvedDep == "" {
				continue
			}

			addEdge(resolvedDep, "explicit")
		}
	}

	// Extract implicit dependencies from expressions
	if resource.Expressions != nil {
		implicitDeps := extractResourceReferences(resource.Expressions, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d implicit dependencies in resource expressions\n", len(implicitDeps))
			for _, dep := range implicitDeps {
				fmt.Fprintf(os.Stderr, "Debug: Implicit dependency: %s\n", dep)
			}
		}
		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := resolveDependencyAddress(ref, modulePrefix, resourceMap, opts)
			if resolvedDep == "" {
				continue
			}

			// Avoid self-references
			if resolvedDep == resourceAddress {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Skipping self-reference: %s\n", resolvedDep)
				}
				continue
			}

			addEdge(resolvedDep, "implicit")
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractResourceDependencies for %s generated %d edges\n", resourceAddress, len(edges))
	}

	return edges
}

// resolveDependencyAddress resolves a dependency reference to its full address
func resolveDependencyAddress(ref string, modulePrefix string, resourceMap map[string]string, opts Options) string {
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: Resolving dependency: %s (module prefix: %s)\n", ref, modulePrefix)
	}

	// Try the reference as-is first
	if _, exists := resourceMap[ref]; exists {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found exact match for dependency: %s\n", ref)
		}
		return ref
	}

	// If we have a module prefix and the reference doesn't start with "module."
	// then it's likely a local reference within the module
	if modulePrefix != "" && !strings.HasPrefix(ref, "module.") {
		// Remove any attribute references (e.g., ".id", ".name")
		refParts := strings.Split(ref, ".")
		if len(refParts) >= 2 {
			// Reconstruct the resource reference without attributes
			resourceRef := strings.Join(refParts[:2], ".")
			fullRef := modulePrefix + "." + resourceRef

			if _, exists := resourceMap[fullRef]; exists {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Resolved local reference %s to %s\n", ref, fullRef)
				}
				return fullRef
			}
		}
	}

	// Handle module output references (e.g., module.network.network_id)
	if strings.HasPrefix(ref, "module.") {
		// This is a module output reference, we need to find the actual resource
		// For now, we'll use a heuristic based on common output patterns
		if strings.Contains(ref, ".network_id") {
			// Look for google_compute_network resources in that module
			moduleNameParts := strings.Split(ref, ".")
			if len(moduleNameParts) >= 2 {
				moduleName := strings.Join(moduleNameParts[:2], ".")
				for resourceAddr := range resourceMap {
					if strings.HasPrefix(resourceAddr, moduleName+".") && strings.Contains(resourceAddr, "google_compute_network") {
						if opts.Verbose {
							fmt.Fprintf(os.Stderr, "Debug: Resolved module output %s to network resource %s\n", ref, resourceAddr)
						}
						return resourceAddr
					}
				}
			}
		} else if strings.Contains(ref, ".subnet_id") {
			// Look for google_compute_subnetwork resources in that module
			moduleNameParts := strings.Split(ref, ".")
			if len(moduleNameParts) >= 2 {
				moduleName := strings.Join(moduleNameParts[:2], ".")
				for resourceAddr := range resourceMap {
					if strings.HasPrefix(resourceAddr, moduleName+".") && strings.Contains(resourceAddr, "google_compute_subnetwork") {
						if opts.Verbose {
							fmt.Fprintf(os.Stderr, "Debug: Resolved module output %s to subnet resource %s\n", ref, resourceAddr)
						}
						return resourceAddr
					}
				}
			}
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: Could not resolve dependency: %s\n", ref)
	}
	return ""
}

// extractResourceReferences recursively searches through expressions to find resource references
func extractResourceReferences(expressions map[string]interface{}, opts Options) []string {
	var references []string

	for key, expr := range expressions {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing expression key: %s\n", key)
		}
		refs := findResourceRefsInExpression(expr, opts)
		references = append(references, refs...)
	}

	return references
}

// findResourceRefsInExpression recursively searches for resource references in an expression
func findResourceRefsInExpression(expr interface{}, opts Options) []string {
	var refs []string

	switch v := expr.(type) {
	case map[string]interface{}:
		// Check for references in expression structure
		if ref, ok := v["references"]; ok {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Found references field in expression\n")
			}
			if refList, ok := ref.([]interface{}); ok {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: References list has %d items\n", len(refList))
				}
				for _, r := range refList {
					if refStr, ok := r.(string); ok {
						if opts.Verbose {
							fmt.Fprintf(os.Stderr, "Debug: Checking reference: %s\n", refStr)
						}
						// Check if this looks like a resource reference
						if isResourceReference(refStr, opts) {
							refs = append(refs, refStr)
							if opts.Verbose {
								fmt.Fprintf(os.Stderr, "Debug: Added resource reference: %s\n", refStr)
							}
						} else {
							if opts.Verbose {
								fmt.Fprintf(os.Stderr, "Debug: Skipped non-resource reference: %s\n", refStr)
							}
						}
					}
				}
			}
		}

		// Recursively search nested expressions
		for key, val := range v {
			if opts.Verbose && key != "references" {
				fmt.Fprintf(os.Stderr, "Debug: Recursively processing nested key: %s\n", key)
			}
			nestedRefs := findResourceRefsInExpression(val, opts)
			refs = append(refs, nestedRefs...)
		}

	case []interface{}:
		// Search through array elements
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing array with %d elements\n", len(v))
		}
		for _, item := range v {
			nestedRefs := findResourceRefsInExpression(item, opts)
			refs = append(refs, nestedRefs...)
		}
	}

	return refs
}

// isResourceReference checks if a string looks like a Terraform resource reference
func isResourceReference(ref string, opts Options) bool {
	// Resource references typically follow the pattern: resource_type.resource_name
	// or module.resource_type.resource_name
	parts := strings.Split(ref, ".")
	if len(parts) < 2 {
		return false
	}

	// Check if it starts with a valid resource type pattern
	// This is a simplified check - in practice, you'd want a more comprehensive list
	validPrefixes := []string{
		"aws_", "google_", "azurerm_", "kubernetes_", "docker_", "null_", "random_",
		"local_", "template_", "archive_", "external_", "http_", "tls_", "time_",
	}

	firstPart := parts[0]
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(firstPart, prefix) {
			return true
		}
	}

	// Also check for module references
	if firstPart == "module" && len(parts) >= 3 {
		return true
	}

	return false
}

// extractOutputDependencies extracts dependency relationships for output configurations
func extractOutputDependencies(outputAddress string, outputConfig OutputConfig, resourceMap map[string]string, opts Options) []GraphEdge {
	var edges []GraphEdge

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractOutputDependencies for output: %s\n", outputAddress)
	}

	// Extract dependencies from output expression
	if outputConfig.Expression != nil {
		implicitDeps := extractResourceReferences(outputConfig.Expression, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d dependencies in output expression\n", len(implicitDeps))
			for _, dep := range implicitDeps {
				fmt.Fprintf(os.Stderr, "Debug: Output dependency: %s\n", dep)
			}
		}

		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := resolveDependencyAddress(ref, "", resourceMap, opts)
			if resolvedDep == "" {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Could not resolve output dependency %s\n", ref)
				}
				continue
			}

			// Create edge from the resource to the output (output depends on resource)
			edge := GraphEdge{
				From: sanitizeID(resolvedDep),
				To:   sanitizeID(outputAddress),
			}
			edges = append(edges, edge)

			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Added output dependency edge: %s -> %s\n", sanitizeID(resolvedDep), sanitizeID(outputAddress))
			}
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractOutputDependencies for %s generated %d edges\n", outputAddress, len(edges))
	}

	return edges
}

// GenerateGraph generates a graph in the specified format
func GenerateGraph(graphData *GraphData, opts Options) (string, error) {
	switch opts.Format {
	case FormatGraphviz:
		return generateGraphviz(graphData, opts)
	case FormatMermaid:
		return generateMermaid(graphData, opts)
	case FormatPlantUML:
		return generatePlantUML(graphData, opts)
	default:
		return "", fmt.Errorf("unsupported format: %s", opts.Format)
	}
}

// generateGraphviz generates a Graphviz DOT format graph
func generateGraphviz(graphData *GraphData, opts Options) (string, error) {
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
			if node.Type == string(NodeTypeResource) {
				color = getActionColor(actionType)
			} else {
				color = getNodeTypeColor(node.Type)
			}

			shape := getNodeTypeShape(node.Type)
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

// generateMermaid generates a Mermaid format graph
func generateMermaid(graphData *GraphData, opts Options) (string, error) {
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

	// Add CSS class definitions for Terraform action colors
	builder.WriteString("classDef create fill:#d4edda,stroke:#c3e6cb,stroke-width:2px,color:#155724\n")
	builder.WriteString("classDef update fill:#fff3cd,stroke:#ffeaa7,stroke-width:2px,color:#856404\n")
	builder.WriteString("classDef delete fill:#f8d7da,stroke:#f5c6cb,stroke-width:2px,color:#721c24\n")
	builder.WriteString("classDef replace fill:#fde2e2,stroke:#fecaca,stroke-width:2px,color:#991b1b\n")
	builder.WriteString("classDef noop fill:#e9ecef,stroke:#dee2e6,stroke-width:2px,color:#495057\n")
	builder.WriteString("classDef default fill:#f8f9fa,stroke:#dee2e6,stroke-width:2px,color:#495057\n")
	builder.WriteString("classDef resource fill:#d4edda,stroke:#c3e6cb,stroke-width:2px,color:#155724\n")
	builder.WriteString("classDef datasource fill:#d1ecf1,stroke:#bee5eb,stroke-width:2px,color:#0c5460\n")
	builder.WriteString("classDef output fill:#cce5ff,stroke:#b3d9ff,stroke-width:2px,color:#004085\n")
	builder.WriteString("classDef variable fill:#fff3cd,stroke:#ffeaa7,stroke-width:2px,color:#856404\n")
	builder.WriteString("classDef local fill:#f8d7da,stroke:#f5c6cb,stroke-width:2px,color:#721c24\n\n")

	builder.WriteString("graph TB\n")

	// Group nodes by module
	moduleGroups := groupNodesByModule(graphData.Nodes)

	for moduleName, nodes := range moduleGroups {
		if moduleName == "" {
			moduleName = "Root Module"
		}

		builder.WriteString(fmt.Sprintf("  subgraph %s[\"%s\"]\n",
			sanitizeID(moduleName), moduleName))

		for _, node := range nodes {
			actionType := getActionType(node.Actions)
			label := fmt.Sprintf("%s<br/>[%s]", node.Address, actionType)

			// Get color based on action type for resources, or node type for others
			var color string
			if node.Type == string(NodeTypeResource) {
				color = getMermaidActionColor(actionType)
			} else {
				color = getMermaidNodeTypeColor(node.Type)
			}

			// Use different syntax for each node type in Mermaid with styling
			switch node.Type {
			case string(NodeTypeResource):
				builder.WriteString(fmt.Sprintf("    %s[\"%s\"]:::%s\n", node.ID, label, color)) // Rectangle with color
			case string(NodeTypeData):
				builder.WriteString(fmt.Sprintf("    %s{\"%s\"}:::%s\n", node.ID, label, color)) // Rhombus/Diamond with color
			case string(NodeTypeOutput):
				builder.WriteString(fmt.Sprintf("    %s((%s)):::%s\n", node.ID, label, color)) // Circle with color
			case string(NodeTypeVariable):
				builder.WriteString(fmt.Sprintf("    %s[/%s/]:::%s\n", node.ID, label, color)) // Parallelogram with color
			case string(NodeTypeLocal):
				builder.WriteString(fmt.Sprintf("    %s{{%s}}:::%s\n", node.ID, label, color)) // Hexagon with color
			default:
				builder.WriteString(fmt.Sprintf("    %s[\"%s\"]:::%s\n", node.ID, label, color)) // Default rectangle with color
			}
		}

		builder.WriteString("  end\n\n")
	}

	// Add edges
	for _, edge := range graphData.Edges {
		builder.WriteString(fmt.Sprintf("  %s --> %s\n", edge.From, edge.To))
	}

	return builder.String(), nil
}

// generatePlantUML generates a PlantUML format graph
func generatePlantUML(graphData *GraphData, opts Options) (string, error) {
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

			// Use different notation for each node type in PlantUML
			switch node.Type {
			case string(NodeTypeResource):
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Component
			case string(NodeTypeData):
				builder.WriteString(fmt.Sprintf("  <%s> as %s\n", label, node.ID)) // Database
			case string(NodeTypeOutput):
				builder.WriteString(fmt.Sprintf("  (%s) as %s\n", label, node.ID)) // Use case/Circle
			case string(NodeTypeVariable):
				builder.WriteString(fmt.Sprintf("  \"%s\" as %s\n", label, node.ID)) // Note
			case string(NodeTypeLocal):
				builder.WriteString(fmt.Sprintf("  {%s} as %s\n", label, node.ID)) // Frame
			default:
				builder.WriteString(fmt.Sprintf("  [%s] as %s\n", label, node.ID)) // Default component
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

func getActionType(actions []string) ActionType {
	if len(actions) == 0 {
		return ActionNoOp
	}

	// Sort actions for consistent comparison
	sortedActions := make([]string, len(actions))
	copy(sortedActions, actions)
	sort.Strings(sortedActions)

	actionStr := strings.Join(sortedActions, ",")

	switch actionStr {
	case "create":
		return ActionCreate
	case "update":
		return ActionUpdate
	case "delete":
		return ActionDelete
	case "create,delete", "delete,create":
		return ActionReplace
	case "no-op":
		return ActionNoOp
	default:
		return ActionNoOp
	}
}

func getActionColor(actionType ActionType) string {
	switch actionType {
	case ActionCreate:
		return "lightgreen"
	case ActionUpdate:
		return "yellow"
	case ActionDelete:
		return "lightcoral"
	case ActionReplace:
		return "orange"
	case ActionNoOp:
		return "lightgrey"
	default:
		return "white"
	}
}

// getNodeTypeColor returns the color for a specific node type
func getNodeTypeColor(nodeType string) string {
	switch nodeType {
	case string(NodeTypeResource):
		return "lightgreen" // Default resource color
	case string(NodeTypeData):
		return "lightcyan" // Data sources
	case string(NodeTypeOutput):
		return "lightblue" // Outputs
	case string(NodeTypeVariable):
		return "lightyellow" // Variables
	case string(NodeTypeLocal):
		return "lightpink" // Locals
	default:
		return "white"
	}
}

// getNodeTypeShape returns the shape for a specific node type in Graphviz
func getNodeTypeShape(nodeType string) string {
	switch nodeType {
	case string(NodeTypeResource):
		return "box" // Managed resources
	case string(NodeTypeData):
		return "diamond" // Data sources
	case string(NodeTypeOutput):
		return "ellipse" // Outputs
	case string(NodeTypeVariable):
		return "parallelogram" // Variables
	case string(NodeTypeLocal):
		return "hexagon" // Locals
	default:
		return "box"
	}
}

func hasSensitiveValues(sensitive interface{}) bool {
	switch v := sensitive.(type) {
	case bool:
		return v
	case map[string]interface{}:
		return len(v) > 0
	default:
		return false
	}
}

func groupNodesByModule(nodes []GraphNode) map[string][]GraphNode {
	groups := make(map[string][]GraphNode)

	for _, node := range nodes {
		module := node.Module
		if module == "" {
			module = "root"
		}
		groups[module] = append(groups[module], node)
	}

	return groups
}

// isResourceType checks if a string looks like a Terraform resource type
func isResourceType(s string) bool {
	// Resource types typically contain underscores (e.g., google_compute_network)
	return strings.Contains(s, "_")
}

// getMermaidActionColor returns the Mermaid color class for a specific action type
func getMermaidActionColor(actionType ActionType) string {
	switch actionType {
	case ActionCreate:
		return "create"
	case ActionUpdate:
		return "update"
	case ActionDelete:
		return "delete"
	case ActionReplace:
		return "replace"
	case ActionNoOp:
		return "noop"
	default:
		return "default"
	}
}

// getMermaidNodeTypeColor returns the Mermaid color class for a specific node type
func getMermaidNodeTypeColor(nodeType string) string {
	switch nodeType {
	case string(NodeTypeResource):
		return "resource" // Default resource color
	case string(NodeTypeData):
		return "datasource" // Data sources
	case string(NodeTypeOutput):
		return "output" // Outputs
	case string(NodeTypeVariable):
		return "variable" // Variables
	case string(NodeTypeLocal):
		return "local" // Locals
	default:
		return "default"
	}
}
