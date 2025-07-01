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

// MarkdownFormatter formats plan summaries as markdown
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// Format formats a plan summary as markdown
func (f *MarkdownFormatter) Format(summary *core.PlanSummary, opts core.SummaryOptions) (string, error) {
	var builder strings.Builder

	// Header
	f.writeHeader(&builder, summary.PlanInfo)

	// Statistics
	f.writeStatistics(&builder, summary.Statistics)

	// Resource Changes
	f.writeResourceChanges(&builder, summary.Changes, opts)

	// Output Changes
	if len(summary.Outputs) > 0 {
		f.writeOutputChanges(&builder, summary.Outputs)
	}

	return builder.String(), nil
}

// writeHeader writes the plan header information
func (f *MarkdownFormatter) writeHeader(builder *strings.Builder, info core.PlanInfo) {
	builder.WriteString("# Terraform Plan Summary\n\n")

	status := "✅ Applicable"
	if !info.Applicable {
		status = "❌ Not Applicable"
	}
	if info.Errored {
		status = "💥 Errored"
	}

	fmt.Fprintf(builder, "**Plan Status:** %s  \n", status)
	fmt.Fprintf(builder, "**Format Version:** %s  \n", info.FormatVersion)
	fmt.Fprintf(builder, "**Complete:** %t  \n\n", info.Complete)
}

// writeStatistics writes the statistics section
func (f *MarkdownFormatter) writeStatistics(builder *strings.Builder, stats core.Statistics) {
	builder.WriteString("## 📊 Statistics\n\n")
	fmt.Fprintf(builder, "**Total Changes:** %d\n\n", stats.TotalChanges)

	// Action breakdown
	if len(stats.ActionBreakdown) > 0 {
		builder.WriteString("### By Action\n\n")
		for action, count := range stats.ActionBreakdown {
			icon := f.getActionIcon(action)
			fmt.Fprintf(builder, "- %s **%s:** %d\n", icon, action, count)
		}
		builder.WriteString("\n")
	}

	// Provider breakdown
	if len(stats.ProviderBreakdown) > 0 {
		builder.WriteString("### By Provider\n\n")
		for provider, count := range stats.ProviderBreakdown {
			fmt.Fprintf(builder, "- 🏢 **%s:** %d\n", provider, count)
		}
		builder.WriteString("\n")
	}

	// Module breakdown
	if len(stats.ModuleBreakdown) > 0 {
		builder.WriteString("### By Module\n\n")
		for module, count := range stats.ModuleBreakdown {
			moduleName := module
			if module == "root" {
				moduleName = "Root Module"
			}
			fmt.Fprintf(builder, "- 📦 **%s:** %d\n", moduleName, count)
		}
		builder.WriteString("\n")
	}
}

// writeResourceChanges writes the resource changes section
func (f *MarkdownFormatter) writeResourceChanges(builder *strings.Builder, changes core.Changes, opts core.SummaryOptions) {
	builder.WriteString("## 🔄 Resource Changes\n\n")

	// Create
	if len(changes.Create) > 0 {
		f.writeActionGroup(builder, "➕ Create", changes.Create, opts)
	}

	// Update
	if len(changes.Update) > 0 {
		f.writeActionGroup(builder, "🔄 Update", changes.Update, opts)
	}

	// Replace
	if len(changes.Replace) > 0 {
		f.writeActionGroup(builder, "🔄 Replace", changes.Replace, opts)
	}

	// Delete
	if len(changes.Delete) > 0 {
		f.writeActionGroup(builder, "❌ Delete", changes.Delete, opts)
	}

	// No-op
	if len(changes.NoOp) > 0 {
		f.writeActionGroup(builder, "➖ No-op", changes.NoOp, opts)
	}
}

// writeActionGroup writes a group of resources with the same action
func (f *MarkdownFormatter) writeActionGroup(builder *strings.Builder, title string, resources []core.ResourceSummary, opts core.SummaryOptions) {
	fmt.Fprintf(builder, "### %s (%d)\n\n", title, len(resources))

	for _, resource := range resources {
		f.writeResource(builder, resource, opts)
	}
	builder.WriteString("\n")
}

// writeResource writes a single resource
func (f *MarkdownFormatter) writeResource(builder *strings.Builder, resource core.ResourceSummary, opts core.SummaryOptions) {
	// Resource address
	address := resource.Address
	fmt.Fprintf(builder, "- **%s**\n", address)

	// Sensitive indicator
	if resource.Sensitive {
		builder.WriteString("  - 🔒 Contains sensitive values\n")
	}

	// Key changes if details are requested
	if opts.ShowDetails && len(resource.KeyChanges) > 0 {
		builder.WriteString("  - **Changes:**\n")
		for key, change := range resource.KeyChanges {
			if changeMap, ok := change.(map[string]interface{}); ok {
				from := changeMap["from"]
				to := changeMap["to"]
				fmt.Fprintf(builder, "    - `%s`: `%v` → `%v`\n", key, from, to)
			}
		}
	}
}

// writeOutputChanges writes the output changes section
func (f *MarkdownFormatter) writeOutputChanges(builder *strings.Builder, outputs []core.OutputSummary) {
	builder.WriteString("## 📤 Output Changes\n\n")

	for _, output := range outputs {
		fmt.Fprintf(builder, "- **%s**\n", output.Name)
		if output.Sensitive {
			builder.WriteString("  - 🔒 Sensitive value\n")
		} else if output.Value != nil {
			fmt.Fprintf(builder, "  - **Value:** `%v`\n", output.Value)
		}
	}
	builder.WriteString("\n")
}

// getActionIcon returns an icon for the given action
func (f *MarkdownFormatter) getActionIcon(action string) string {
	switch action {
	case "create":
		return "➕"
	case "update":
		return "🔄"
	case "delete":
		return "❌"
	case "replace":
		return "🔄"
	case "no-op":
		return "➖"
	default:
		return "❓"
	}
}
