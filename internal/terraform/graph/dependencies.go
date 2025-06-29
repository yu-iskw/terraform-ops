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

// analyzeDependencies extracts dependency relationships from the plan
func (b *Builder) analyzeDependencies(plan *core.TerraformPlan, opts core.GraphOptions) ([]core.GraphEdge, error) {
	var edges []core.GraphEdge

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
			resourceEdges := b.extractResourceDependencies(resource.Address, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from module calls
		for moduleName, moduleCall := range plan.Configuration.RootModule.ModuleCalls {
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing module: %s\n", moduleName)
			}
			moduleEdges := b.extractModuleDependencies(moduleName, moduleCall, resourceMap, opts)
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
				outputEdges := b.extractOutputDependencies(outputAddress, outputConfig, resourceMap, opts)
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
func (b *Builder) extractModuleDependencies(moduleName string, moduleCall core.ModuleCall, resourceMap map[string]string, opts core.GraphOptions) []core.GraphEdge {
	var edges []core.GraphEdge

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractModuleDependencies for module: %s\n", moduleName)
		fmt.Fprintf(os.Stderr, "Debug: Module has expressions: %v\n", moduleCall.Expressions != nil)
		fmt.Fprintf(os.Stderr, "Debug: Module has nested module: %v\n", moduleCall.Module != nil)
	}

	// Extract dependencies from module expressions
	if moduleCall.Expressions != nil {
		modulePrefix := "module." + moduleName
		implicitDeps := b.extractResourceReferences(moduleCall.Expressions, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d implicit dependencies in module expressions\n", len(implicitDeps))
		}
		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := b.resolveDependencyAddress(ref, "", resourceMap, opts)
			if resolvedDep == "" {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Could not resolve module dependency %s\n", ref)
				}
				continue
			}

			// Create edges from all resources in this module to the dependency
			for resourceAddr, resourceID := range resourceMap {
				if strings.HasPrefix(resourceAddr, modulePrefix) {
					edge := core.GraphEdge{
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
			resourceEdges := b.extractResourceDependencies(fullAddress, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from nested module calls
		for nestedModuleName, nestedModuleCall := range moduleCall.Module.ModuleCalls {
			nestedModulePrefix := modulePrefix + ".module." + nestedModuleName
			if opts.Verbose {
				fmt.Fprintf(os.Stderr, "Debug: Processing nested module call: %s\n", nestedModulePrefix)
			}
			nestedEdges := b.extractNestedModuleDependencies(nestedModulePrefix, nestedModuleCall, resourceMap, opts)
			edges = append(edges, nestedEdges...)
		}
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractModuleDependencies for %s generated %d edges\n", moduleName, len(edges))
	}

	return edges
}

// extractNestedModuleDependencies extracts dependencies from nested modules
func (b *Builder) extractNestedModuleDependencies(modulePrefix string, moduleCall core.ModuleCall, resourceMap map[string]string, opts core.GraphOptions) []core.GraphEdge {
	var edges []core.GraphEdge

	// Extract dependencies from module expressions
	if moduleCall.Expressions != nil {
		implicitDeps := b.extractResourceReferences(moduleCall.Expressions, opts)
		for _, ref := range implicitDeps {
			// Check if the dependency resource exists in our resource changes
			if _, exists := resourceMap[ref]; !exists {
				continue
			}

			// Create edges from all resources in this module to the dependency
			for resourceAddr, resourceID := range resourceMap {
				if strings.HasPrefix(resourceAddr, modulePrefix) {
					edge := core.GraphEdge{
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
			resourceEdges := b.extractResourceDependencies(fullAddress, resource, resourceMap, opts)
			edges = append(edges, resourceEdges...)
		}

		// Extract dependencies from nested module calls
		for nestedModuleName, nestedModuleCall := range moduleCall.Module.ModuleCalls {
			nestedModulePrefix := modulePrefix + ".module." + nestedModuleName
			nestedEdges := b.extractNestedModuleDependencies(nestedModulePrefix, nestedModuleCall, resourceMap, opts)
			edges = append(edges, nestedEdges...)
		}
	}

	return edges
}

// extractResourceDependencies extracts dependencies for a specific resource
func (b *Builder) extractResourceDependencies(resourceAddress string, resource core.ConfigurationResource, resourceMap map[string]string, opts core.GraphOptions) []core.GraphEdge {
	var edges []core.GraphEdge
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
			edge := core.GraphEdge{
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
			resolvedDep := b.resolveDependencyAddress(dependsOn, modulePrefix, resourceMap, opts)
			if resolvedDep == "" {
				continue
			}

			addEdge(resolvedDep, "explicit")
		}
	}

	// Extract implicit dependencies from expressions
	if resource.Expressions != nil {
		implicitDeps := b.extractResourceReferences(resource.Expressions, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d implicit dependencies in resource expressions\n", len(implicitDeps))
			for _, dep := range implicitDeps {
				fmt.Fprintf(os.Stderr, "Debug: Implicit dependency: %s\n", dep)
			}
		}
		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := b.resolveDependencyAddress(ref, modulePrefix, resourceMap, opts)
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
func (b *Builder) resolveDependencyAddress(ref string, modulePrefix string, resourceMap map[string]string, opts core.GraphOptions) string {
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
func (b *Builder) extractResourceReferences(expressions map[string]interface{}, opts core.GraphOptions) []string {
	var references []string

	for key, expr := range expressions {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing expression key: %s\n", key)
		}
		refs := b.findResourceRefsInExpression(expr, opts)
		references = append(references, refs...)
	}

	return references
}

// findResourceRefsInExpression recursively searches for resource references in an expression
func (b *Builder) findResourceRefsInExpression(expr interface{}, opts core.GraphOptions) []string {
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
			nestedRefs := b.findResourceRefsInExpression(val, opts)
			refs = append(refs, nestedRefs...)
		}

	case []interface{}:
		// Search through array elements
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Processing array with %d elements\n", len(v))
		}
		for _, item := range v {
			nestedRefs := b.findResourceRefsInExpression(item, opts)
			refs = append(refs, nestedRefs...)
		}
	}

	return refs
}

// isResourceReference checks if a string looks like a Terraform resource reference
func isResourceReference(ref string, opts core.GraphOptions) bool {
	// Resource references typically follow the pattern: resource_type.resource_name
	// or module.resource_type.resource_name
	parts := strings.Split(ref, ".")
	if len(parts) < 2 {
		return false
	}

	// Check if it starts with a valid resource type pattern
	// Since we're parsing from a valid Terraform plan, any resource type
	// that follows the provider_resource_type pattern should be considered valid
	firstPart := parts[0]
	resourceTypeParts := strings.Split(firstPart, "_")
	if len(resourceTypeParts) >= 2 {
		return true
	}

	// Also check for module references
	if firstPart == "module" && len(parts) >= 3 {
		return true
	}

	return false
}

// extractOutputDependencies extracts dependency relationships for output configurations
func (b *Builder) extractOutputDependencies(outputAddress string, outputConfig core.OutputConfig, resourceMap map[string]string, opts core.GraphOptions) []core.GraphEdge {
	var edges []core.GraphEdge

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Debug: extractOutputDependencies for output: %s\n", outputAddress)
	}

	// Extract dependencies from output expression
	if outputConfig.Expression != nil {
		implicitDeps := b.extractResourceReferences(outputConfig.Expression, opts)
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Debug: Found %d dependencies in output expression\n", len(implicitDeps))
			for _, dep := range implicitDeps {
				fmt.Fprintf(os.Stderr, "Debug: Output dependency: %s\n", dep)
			}
		}

		for _, ref := range implicitDeps {
			// Resolve the dependency address
			resolvedDep := b.resolveDependencyAddress(ref, "", resourceMap, opts)
			if resolvedDep == "" {
				if opts.Verbose {
					fmt.Fprintf(os.Stderr, "Debug: Could not resolve output dependency %s\n", ref)
				}
				continue
			}

			// Create edge from the resource to the output (output depends on resource)
			edge := core.GraphEdge{
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
