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

package formatters

import (
	"fmt"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// TableFormatter formats plan summaries as tables
type TableFormatter struct{}

// NewTableFormatter creates a new table formatter
func NewTableFormatter() *TableFormatter {
	return &TableFormatter{}
}

// Format formats a plan summary as tables
func (f *TableFormatter) Format(summary *core.PlanSummary, opts core.SummaryOptions) (string, error) {
	var builder strings.Builder

	// Statistics table
	f.writeStatisticsTable(&builder, summary.Statistics)

	// Resource changes table
	f.writeResourceChangesTable(&builder, summary.Changes, opts)

	// Output changes table
	if len(summary.Outputs) > 0 {
		f.writeOutputChangesTable(&builder, summary.Outputs)
	}

	return builder.String(), nil
}

// writeStatisticsTable writes the statistics as a table
func (f *TableFormatter) writeStatisticsTable(builder *strings.Builder, stats core.Statistics) {
	builder.WriteString("## Statistics\n\n")

	// Action breakdown table
	if len(stats.ActionBreakdown) > 0 {
		builder.WriteString("### Action Breakdown\n\n")
		builder.WriteString("| Action | Count |\n")
		builder.WriteString("|--------|-------|\n")
		for action, count := range stats.ActionBreakdown {
			fmt.Fprintf(builder, "| %s | %d |\n", action, count)
		}
		builder.WriteString("\n")
	}

	// Provider breakdown table
	if len(stats.ProviderBreakdown) > 0 {
		builder.WriteString("### Provider Breakdown\n\n")
		builder.WriteString("| Provider | Count |\n")
		builder.WriteString("|----------|-------|\n")
		for provider, count := range stats.ProviderBreakdown {
			fmt.Fprintf(builder, "| %s | %d |\n", provider, count)
		}
		builder.WriteString("\n")
	}

	// Module breakdown table
	if len(stats.ModuleBreakdown) > 0 {
		builder.WriteString("### Module Breakdown\n\n")
		builder.WriteString("| Module | Count |\n")
		builder.WriteString("|--------|-------|\n")
		for module, count := range stats.ModuleBreakdown {
			moduleName := module
			if module == "root" {
				moduleName = "Root Module"
			}
			fmt.Fprintf(builder, "| %s | %d |\n", moduleName, count)
		}
		builder.WriteString("\n")
	}
}

// writeResourceChangesTable writes the resource changes as a table
func (f *TableFormatter) writeResourceChangesTable(builder *strings.Builder, changes core.Changes, opts core.SummaryOptions) {
	builder.WriteString("## Resource Changes\n\n")

	// Create table
	if len(changes.Create) > 0 {
		f.writeActionTable(builder, "Create", changes.Create, opts)
	}

	// Update table
	if len(changes.Update) > 0 {
		f.writeActionTable(builder, "Update", changes.Update, opts)
	}

	// Replace table
	if len(changes.Replace) > 0 {
		f.writeActionTable(builder, "Replace", changes.Replace, opts)
	}

	// Delete table
	if len(changes.Delete) > 0 {
		f.writeActionTable(builder, "Delete", changes.Delete, opts)
	}

	// No-op table
	if len(changes.NoOp) > 0 {
		f.writeActionTable(builder, "No-op", changes.NoOp, opts)
	}
}

// writeActionTable writes a table for resources with the same action
func (f *TableFormatter) writeActionTable(builder *strings.Builder, action string, resources []core.ResourceSummary, opts core.SummaryOptions) {
	fmt.Fprintf(builder, "### %s (%d)\n\n", action, len(resources))
	builder.WriteString("| Address | Type | Provider | Module | Sensitive |\n")
	builder.WriteString("|---------|------|----------|--------|-----------|\n")

	for _, resource := range resources {
		address := resource.Address
		module := resource.ModuleAddress
		if module == "" {
			module = "root"
		}
		sensitive := "No"
		if resource.Sensitive {
			sensitive = "Yes"
		}
		fmt.Fprintf(builder, "| %s | %s | %s | %s | %s |\n",
			address, resource.Type, resource.Provider, module, sensitive)
	}
	builder.WriteString("\n")
}

// writeOutputChangesTable writes the output changes as a table
func (f *TableFormatter) writeOutputChangesTable(builder *strings.Builder, outputs []core.OutputSummary) {
	builder.WriteString("## Output Changes\n\n")
	builder.WriteString("| Name | Actions | Sensitive | Value |\n")
	builder.WriteString("|------|---------|-----------|-------|\n")

	for _, output := range outputs {
		actions := strings.Join(output.Actions, ", ")
		sensitive := "No"
		if output.Sensitive {
			sensitive = "Yes"
		}

		value := "N/A"
		if !output.Sensitive && output.Value != nil {
			value = fmt.Sprintf("%v", output.Value)
		}

		fmt.Fprintf(builder, "| %s | %s | %s | %s |\n",
			output.Name, actions, sensitive, value)
	}
	builder.WriteString("\n")
}
