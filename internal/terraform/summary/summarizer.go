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

package summary

import (
	"fmt"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// Summarizer implements the core.PlanSummarizer interface
type Summarizer struct{}

// NewSummarizer creates a new plan summarizer
func NewSummarizer() *Summarizer {
	return &Summarizer{}
}

// SummarizePlan creates a summary of the Terraform plan
func (s *Summarizer) SummarizePlan(plan *core.TerraformPlan, opts core.SummaryOptions) (*core.PlanSummary, error) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: plan.FormatVersion,
			Applicable:    plan.Applicable,
			Complete:      plan.Complete,
			Errored:       plan.Errored,
		},
		Statistics: s.calculateStatistics(plan),
		Changes:    s.groupResourceChanges(plan),
		Outputs:    s.summarizeOutputs(plan),
	}

	return summary, nil
}

// calculateStatistics calculates various statistics from the plan
func (s *Summarizer) calculateStatistics(plan *core.TerraformPlan) core.Statistics {
	stats := core.Statistics{
		ActionBreakdown:   make(map[string]int),
		ProviderBreakdown: make(map[string]int),
		ResourceBreakdown: make(map[string]int),
		ModuleBreakdown:   make(map[string]int),
	}

	// Process resource changes
	for _, change := range plan.ResourceChanges {
		stats.TotalChanges++

		// Count by action
		for _, action := range change.Change.Actions {
			stats.ActionBreakdown[action]++
		}

		// Count by provider
		provider := s.extractProvider(change.Address)
		stats.ProviderBreakdown[provider]++

		// Count by resource type
		stats.ResourceBreakdown[change.Type]++

		// Count by module
		module := change.ModuleAddress
		if module == "" {
			module = "root"
		}
		stats.ModuleBreakdown[module]++
	}

	return stats
}

// groupResourceChanges groups resource changes by action type
func (s *Summarizer) groupResourceChanges(plan *core.TerraformPlan) core.Changes {
	changes := core.Changes{}

	for _, change := range plan.ResourceChanges {
		summary := core.ResourceSummary{
			Address:       change.Address,
			ModuleAddress: change.ModuleAddress,
			Type:          change.Type,
			Name:          change.Name,
			Provider:      s.extractProvider(change.Address),
			Actions:       change.Change.Actions,
			Sensitive:     s.hasSensitiveValues(change.Change),
		}

		// Add key changes if details are requested
		summary.KeyChanges = s.extractKeyChanges(change)

		// Group by primary action
		primaryAction := s.getPrimaryAction(change.Change.Actions)
		switch primaryAction {
		case "create":
			changes.Create = append(changes.Create, summary)
		case "update":
			changes.Update = append(changes.Update, summary)
		case "delete":
			changes.Delete = append(changes.Delete, summary)
		case "replace":
			changes.Replace = append(changes.Replace, summary)
		case "no-op":
			changes.NoOp = append(changes.NoOp, summary)
		}
	}

	return changes
}

// summarizeOutputs creates summaries of output changes
func (s *Summarizer) summarizeOutputs(plan *core.TerraformPlan) []core.OutputSummary {
	var outputs []core.OutputSummary

	for name, outputChange := range plan.OutputChanges {
		summary := core.OutputSummary{
			Name:      name,
			Actions:   outputChange.Change.Actions,
			Sensitive: s.hasSensitiveValues(outputChange.Change),
		}

		// Add value if not sensitive
		if !summary.Sensitive {
			summary.Value = outputChange.Change.After
		}

		outputs = append(outputs, summary)
	}

	return outputs
}

// extractProvider extracts the provider name from a resource address
func (s *Summarizer) extractProvider(address string) string {
	// Resource addresses typically follow the pattern: provider_type.name
	// For module resources: module.name.provider_type.name
	// For nested modules: module.name.module.subname.provider_type.name
	parts := strings.Split(address, ".")

	// Find the resource type part (the part that contains underscore)
	for _, part := range parts {
		if strings.Contains(part, "_") {
			// This is the resource type, extract provider from it
			providerParts := strings.Split(part, "_")
			if len(providerParts) >= 2 {
				return providerParts[0]
			}
			return part
		}
	}

	// Fallback: try to extract from the last part if no underscore found
	if len(parts) >= 2 {
		lastPart := parts[len(parts)-2] // The part before the resource name
		if strings.Contains(lastPart, "_") {
			providerParts := strings.Split(lastPart, "_")
			if len(providerParts) >= 2 {
				return providerParts[0]
			}
		}
	}

	return "unknown"
}

// hasSensitiveValues checks if a change has sensitive values
func (s *Summarizer) hasSensitiveValues(change core.Change) bool {
	// Check after_sensitive
	if change.AfterSensitive != nil {
		if sensitiveMap, ok := change.AfterSensitive.(map[string]interface{}); ok {
			for _, sensitive := range sensitiveMap {
				if sensitiveBool, ok := sensitive.(bool); ok && sensitiveBool {
					return true
				}
			}
		}
	}

	// Check before_sensitive
	if change.BeforeSensitive != nil {
		if sensitiveMap, ok := change.BeforeSensitive.(map[string]interface{}); ok {
			for _, sensitive := range sensitiveMap {
				if sensitiveBool, ok := sensitive.(bool); ok && sensitiveBool {
					return true
				}
			}
		}
	}

	return false
}

// extractKeyChanges extracts key changes from a resource change
func (s *Summarizer) extractKeyChanges(change core.ResourceChange) map[string]interface{} {
	keyChanges := make(map[string]interface{})

	// Extract key changes from before/after values
	if change.Change.Before != nil && change.Change.After != nil {
		if beforeMap, ok := change.Change.Before.(map[string]interface{}); ok {
			if afterMap, ok := change.Change.After.(map[string]interface{}); ok {
				// Find changed keys
				for key, afterValue := range afterMap {
					if beforeValue, exists := beforeMap[key]; exists {
						if fmt.Sprintf("%v", beforeValue) != fmt.Sprintf("%v", afterValue) {
							keyChanges[key] = map[string]interface{}{
								"from": beforeValue,
								"to":   afterValue,
							}
						}
					} else {
						// New key
						keyChanges[key] = map[string]interface{}{
							"from": nil,
							"to":   afterValue,
						}
					}
				}

				// Find deleted keys
				for key, beforeValue := range beforeMap {
					if _, exists := afterMap[key]; !exists {
						keyChanges[key] = map[string]interface{}{
							"from": beforeValue,
							"to":   nil,
						}
					}
				}
			}
		}
	} else if change.Change.Before == nil && change.Change.After != nil {
		// Creating new resource
		if afterMap, ok := change.Change.After.(map[string]interface{}); ok {
			for key, value := range afterMap {
				keyChanges[key] = map[string]interface{}{
					"from": nil,
					"to":   value,
				}
			}
		}
	} else if change.Change.Before != nil && change.Change.After == nil {
		// Deleting resource
		if beforeMap, ok := change.Change.Before.(map[string]interface{}); ok {
			for key, value := range beforeMap {
				keyChanges[key] = map[string]interface{}{
					"from": value,
					"to":   nil,
				}
			}
		}
	}

	return keyChanges
}

// getPrimaryAction determines the primary action from a list of actions
func (s *Summarizer) getPrimaryAction(actions []string) string {
	if len(actions) == 0 {
		return "no-op"
	}

	// Handle special cases
	if len(actions) == 2 && contains(actions, "delete") && contains(actions, "create") {
		return "replace"
	}

	// Return the first action as primary
	return actions[0]
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
