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
	"sort"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// getActionType determines the action type from a list of actions
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

// groupNodesByModule groups nodes by their module
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

// getNodeShape determines the shape for a node based on its type
func getNodeShape(nodeType string, resourceType string) string {
	switch nodeType {
	case string(core.NodeTypeResource):
		return "house" // House shape for infrastructure resources
	case string(core.NodeTypeData):
		return "diamond" // Diamond for data sources
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

// isResourceType checks if a string represents a Terraform resource type
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
