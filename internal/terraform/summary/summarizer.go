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
	"reflect"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/ir"
)

// Summarizer projects the normalized ChangeSet into the compatibility summary
// model consumed by the existing formatters. It never reads raw plan JSON.
type Summarizer struct{}

func NewSummarizer() *Summarizer {
	return &Summarizer{}
}

func (s *Summarizer) SummarizePlan(changeSet *ir.ChangeSet, _ core.SummaryOptions) (*core.PlanSummary, error) {
	if changeSet == nil {
		return nil, &core.ValidationError{Field: "change_set", Message: "must not be nil"}
	}

	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: changeSet.Source.PlanFormatVersion,
			Applicable:    changeSet.Plan.Applyable,
			Complete:      changeSet.Plan.Complete,
			Errored:       changeSet.Plan.Errored,
		},
		Statistics: s.calculateStatistics(changeSet),
		Changes:    s.groupResourceChanges(changeSet),
		Outputs:    s.summarizeOutputs(changeSet),
	}
	return summary, nil
}

func (s *Summarizer) calculateStatistics(changeSet *ir.ChangeSet) core.Statistics {
	stats := core.Statistics{
		ActionBreakdown:   make(map[string]int),
		ProviderBreakdown: make(map[string]int),
		ResourceBreakdown: make(map[string]int),
		ModuleBreakdown:   make(map[string]int),
	}

	for _, change := range changeSet.Resources {
		stats.TotalChanges++
		for _, action := range change.Action.Raw {
			stats.ActionBreakdown[action]++
		}
		stats.ProviderBreakdown[extractProviderFromType(change.Type)]++
		stats.ResourceBreakdown[change.Type]++
		module := "root"
		if change.ModuleAddress != nil && *change.ModuleAddress != "" {
			module = string(*change.ModuleAddress)
		}
		stats.ModuleBreakdown[module]++
	}
	return stats
}

func (s *Summarizer) groupResourceChanges(changeSet *ir.ChangeSet) core.Changes {
	changes := core.Changes{}
	for _, change := range changeSet.Resources {
		moduleAddress := ""
		if change.ModuleAddress != nil {
			moduleAddress = string(*change.ModuleAddress)
		}
		item := core.ResourceSummary{
			Address:       string(change.Address),
			ModuleAddress: moduleAddress,
			Type:          change.Type,
			Name:          change.Name,
			Provider:      extractProviderFromType(change.Type),
			Actions:       append([]string(nil), change.Action.Raw...),
			Sensitive:     len(change.SensitivePaths) > 0,
			KeyChanges:    extractKeyChanges(change),
		}

		switch change.Action.Semantic {
		case ir.ActionCreate:
			changes.Create = append(changes.Create, item)
		case ir.ActionUpdate, ir.ActionRead:
			changes.Update = append(changes.Update, item)
		case ir.ActionDelete:
			changes.Delete = append(changes.Delete, item)
		case ir.ActionReplaceDestroyCreate, ir.ActionReplaceCreateDestroy:
			changes.Replace = append(changes.Replace, item)
		case ir.ActionNoOp:
			changes.NoOp = append(changes.NoOp, item)
		default:
			// Preserve visibility for an unrecognized source action rather than
			// silently dropping it from legacy summary renderers.
			changes.Update = append(changes.Update, item)
		}
	}
	return changes
}

func (s *Summarizer) summarizeOutputs(changeSet *ir.ChangeSet) []core.OutputSummary {
	outputs := make([]core.OutputSummary, 0, len(changeSet.Outputs))
	for _, output := range changeSet.Outputs {
		item := core.OutputSummary{
			Name:      output.Name,
			Actions:   append([]string(nil), output.Action.Raw...),
			Sensitive: len(output.SensitivePaths) > 0,
		}
		if !item.Sensitive && !output.After.Redacted {
			item.Value = output.After.Value
		}
		outputs = append(outputs, item)
	}
	return outputs
}

// extractKeyChanges compares values that have already crossed the source
// adapter's sanitization boundary. It deliberately has no access to Terraform
// sensitivity masks or raw values.
func extractKeyChanges(change ir.ResourceChange) map[string]interface{} {
	if change.Before.Redacted || change.After.Redacted {
		return nil
	}
	before := change.Before.Value
	after := change.After.Value
	keyChanges := make(map[string]interface{})

	beforeMap, beforeOK := before.(map[string]interface{})
	afterMap, afterOK := after.(map[string]interface{})

	switch {
	case beforeOK && afterOK:
		for key, afterValue := range afterMap {
			beforeValue, exists := beforeMap[key]
			if !exists || !reflect.DeepEqual(beforeValue, afterValue) {
				keyChanges[key] = map[string]interface{}{"from": beforeValue, "to": afterValue}
			}
		}
		for key, beforeValue := range beforeMap {
			if _, exists := afterMap[key]; !exists {
				keyChanges[key] = map[string]interface{}{"from": beforeValue, "to": nil}
			}
		}
	case before == nil && afterOK:
		for key, value := range afterMap {
			keyChanges[key] = map[string]interface{}{"from": nil, "to": value}
		}
	case beforeOK && after == nil:
		for key, value := range beforeMap {
			keyChanges[key] = map[string]interface{}{"from": value, "to": nil}
		}
	}

	if len(keyChanges) == 0 {
		return nil
	}
	return keyChanges
}

func extractProviderFromType(resourceType string) string {
	provider, _, ok := strings.Cut(resourceType, "_")
	if !ok || provider == "" {
		return "unknown"
	}
	return provider
}
