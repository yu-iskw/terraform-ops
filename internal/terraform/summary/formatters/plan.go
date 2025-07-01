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
	"sort"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// PlanFormatter formats plan summaries like terraform plan output
type PlanFormatter struct {
	useColor bool
}

// NewPlanFormatter creates a new plan formatter
func NewPlanFormatter(useColor bool) *PlanFormatter {
	return &PlanFormatter{
		useColor: useColor,
	}
}

// Format formats a plan summary like terraform plan output
func (f *PlanFormatter) Format(summary *core.PlanSummary, opts core.SummaryOptions) (string, error) {
	var builder strings.Builder

	// Header
	f.writeHeader(&builder, summary.PlanInfo, summary.Statistics)

	// Collect all resources in the order they would appear in terraform plan
	resources := f.collectAllResources(summary.Changes)

	// Sort resources by address for consistent output
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Address < resources[j].Address
	})

	// Write resource changes in terraform plan style
	f.writeResourceChanges(&builder, resources, opts)

	// Write plan summary
	f.writePlanSummary(&builder, summary.Statistics)

	// Write output changes if any
	if len(summary.Outputs) > 0 {
		f.writeOutputChanges(&builder, summary.Outputs)
	}

	// Write footer
	f.writeFooter(&builder)

	return builder.String(), nil
}

// writeHeader writes the terraform plan header
func (f *PlanFormatter) writeHeader(builder *strings.Builder, info core.PlanInfo, stats core.Statistics) {
	builder.WriteString("Terraform used the selected providers to generate the following execution plan. Resource\n")
	builder.WriteString("actions are indicated with the following symbols:\n")
	builder.WriteString("  + create\n")
	builder.WriteString("  ~ update in-place\n")
	builder.WriteString("  - destroy\n")
	builder.WriteString("-/+ destroy and then create replacement\n\n")
}

// collectAllResources collects all resources from changes into a single slice
func (f *PlanFormatter) collectAllResources(changes core.Changes) []core.ResourceSummary {
	var resources []core.ResourceSummary

	resources = append(resources, changes.Create...)
	resources = append(resources, changes.Update...)
	resources = append(resources, changes.Replace...)
	resources = append(resources, changes.Delete...)
	resources = append(resources, changes.NoOp...)

	return resources
}

// writeResourceChanges writes resource changes in terraform plan style
func (f *PlanFormatter) writeResourceChanges(builder *strings.Builder, resources []core.ResourceSummary, opts core.SummaryOptions) {
	if len(resources) == 0 {
		builder.WriteString("No changes. Your infrastructure matches the configuration.\n\n")
		return
	}

	builder.WriteString("Terraform will perform the following actions:\n\n")

	for _, resource := range resources {
		f.writeResourceChange(builder, resource, opts)
		builder.WriteString("\n")
	}
}

// writeResourceChange writes a single resource change in terraform plan style
func (f *PlanFormatter) writeResourceChange(builder *strings.Builder, resource core.ResourceSummary, opts core.SummaryOptions) {
	// Determine the action symbol and description
	actionSymbol, actionColor := f.getActionSymbolAndColor(resource.Actions)
	actionDescription := f.getActionDescription(resource.Actions)

	// Write the resource header
	address := resource.Address
	if resource.ModuleAddress != "" {
		address = resource.Address // Keep original address as it already includes module info
	}

	// Write the comment line with action description
	fmt.Fprintf(builder, "  # %s %s\n", address, actionDescription)

	// Write the resource block header
	if f.useColor && actionColor != "" {
		fmt.Fprintf(builder, "%s resource \"%s\" \"%s\" {\n", f.colorize(actionSymbol, actionColor), resource.Type, resource.Name)
	} else {
		fmt.Fprintf(builder, "%s resource \"%s\" \"%s\" {\n", actionSymbol, resource.Type, resource.Name)
	}

	// Write resource details if showing details
	if opts.ShowDetails && len(resource.KeyChanges) > 0 {
		f.writeResourceDetails(builder, resource, opts)
	} else {
		// Write minimal info for non-detailed view
		if resource.Sensitive {
			builder.WriteString("      # (sensitive value)\n")
		}
	}

	builder.WriteString("    }\n")
}

// getActionDescription returns a human-readable description of the action
func (f *PlanFormatter) getActionDescription(actions []string) string {
	if len(actions) == 0 {
		return "will be created"
	}

	// Handle multiple actions (like replace)
	if len(actions) == 2 && f.containsAction(actions, "delete") && f.containsAction(actions, "create") {
		return "must be replaced"
	}

	switch actions[0] {
	case "create":
		return "will be created"
	case "update":
		return "will be updated in-place"
	case "delete":
		return "will be destroyed"
	case "no-op":
		return "will be unchanged"
	default:
		return "will be modified"
	}
}

// writeResourceDetails writes detailed resource changes
func (f *PlanFormatter) writeResourceDetails(builder *strings.Builder, resource core.ResourceSummary, opts core.SummaryOptions) {
	// Sort keys for consistent output
	var keys []string
	for key := range resource.KeyChanges {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		change := resource.KeyChanges[key]
		if changeMap, ok := change.(map[string]interface{}); ok {
			from := changeMap["from"]
			to := changeMap["to"]

			if from == nil && to != nil {
				// New value
				f.printValueWithSymbol(builder, key, to, "+", 6)
			} else if from != nil && to == nil {
				// Deleted value
				f.printValueWithSymbol(builder, key, from, "-", 6)
			} else if fmt.Sprintf("%v", from) != fmt.Sprintf("%v", to) {
				// Changed value
				f.printValueWithSymbol(builder, key, from, "~", 6)
				builder.WriteString(strings.Repeat(" ", 8) + "-> ")
				f.printValue(builder, to, 0, false)
				builder.WriteString("\n")
			}
		}
	}
}

// printValueWithSymbol prints a key-value pair with a change symbol and indentation
func (f *PlanFormatter) printValueWithSymbol(builder *strings.Builder, key string, value interface{}, symbol string, indent int) {
	builder.WriteString(strings.Repeat(" ", indent))
	builder.WriteString(symbol + " " + key + " = ")
	f.printValue(builder, value, indent, false)
	builder.WriteString("\n")
}

// printValue recursively prints a value (map, slice, or primitive) with indentation
func (f *PlanFormatter) printValue(builder *strings.Builder, value interface{}, indent int, inList bool) {
	switch v := value.(type) {
	case map[string]interface{}:
		builder.WriteString("{")
		if len(v) > 0 {
			builder.WriteString("\n")
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				builder.WriteString(strings.Repeat(" ", indent+4))
				builder.WriteString(k + " = ")
				f.printValue(builder, v[k], indent+4, false)
				builder.WriteString("\n")
			}
			builder.WriteString(strings.Repeat(" ", indent))
		}
		builder.WriteString("}")
	case []interface{}:
		builder.WriteString("[")
		if len(v) > 0 {
			builder.WriteString("\n")
			for _, elem := range v {
				builder.WriteString(strings.Repeat(" ", indent+4))
				f.printValue(builder, elem, indent+4, true)
				builder.WriteString(",\n")
			}
			builder.WriteString(strings.Repeat(" ", indent))
		}
		builder.WriteString("]")
	case string:
		fmt.Fprintf(builder, "%q", v)
	case bool:
		if v {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}
	case nil:
		builder.WriteString("null")
	default:
		fmt.Fprintf(builder, "%v", v)
	}
}

// getActionSymbolAndColor returns the symbol and color for an action
func (f *PlanFormatter) getActionSymbolAndColor(actions []string) (string, string) {
	if len(actions) == 0 {
		return "  ", ""
	}

	// Handle multiple actions (like replace)
	if len(actions) == 2 && f.containsAction(actions, "delete") && f.containsAction(actions, "create") {
		return "-/+", "yellow"
	}

	switch actions[0] {
	case "create":
		return "  +", "green"
	case "update":
		return "  ~", "yellow"
	case "delete":
		return "  -", "red"
	case "no-op":
		return "   ", ""
	default:
		return "  ?", ""
	}
}

// containsAction checks if actions slice contains a specific action
func (f *PlanFormatter) containsAction(actions []string, action string) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}

// colorize applies color codes if color is enabled
func (f *PlanFormatter) colorize(text, color string) string {
	if !f.useColor {
		return text
	}

	switch color {
	case "green":
		return fmt.Sprintf("\033[32m%s\033[0m", text)
	case "yellow":
		return fmt.Sprintf("\033[33m%s\033[0m", text)
	case "red":
		return fmt.Sprintf("\033[31m%s\033[0m", text)
	default:
		return text
	}
}

// formatValue formats a value for display
func (f *PlanFormatter) formatValue(value interface{}) string {
	if value == nil {
		return "(null)"
	}

	switch v := value.(type) {
	case string:
		return fmt.Sprintf(`"%s"`, v)
	case bool:
		if v {
			return "true"
		}
		return "false"
	case map[string]interface{}:
		return "{...}"
	case []interface{}:
		return "[...]"
	default:
		return fmt.Sprintf("%v", v)
	}
}

// writePlanSummary writes the plan summary like terraform plan
func (f *PlanFormatter) writePlanSummary(builder *strings.Builder, stats core.Statistics) {
	creates := stats.ActionBreakdown["create"]
	updates := stats.ActionBreakdown["update"]
	deletes := stats.ActionBreakdown["delete"]
	replaces := stats.ActionBreakdown["replace"]

	total := creates + updates + deletes + replaces
	if total == 0 {
		builder.WriteString("No changes. No objects need to be destroyed.\n")
		return
	}

	fmt.Fprintf(builder, "Plan: %d to add, %d to change, %d to destroy.\n\n",
		creates+replaces, updates, deletes+replaces)
}

// writeOutputChanges writes output changes section
func (f *PlanFormatter) writeOutputChanges(builder *strings.Builder, outputs []core.OutputSummary) {
	builder.WriteString("Changes to Outputs:\n")

	for _, output := range outputs {
		if len(output.Actions) > 0 {
			action := output.Actions[0]
			switch action {
			case "create":
				fmt.Fprintf(builder, "  + %s = %s\n", output.Name, f.formatOutputValue(output))
			case "update":
				fmt.Fprintf(builder, "  ~ %s = %s -> %s\n", output.Name, f.formatOutputValue(output), f.formatOutputValue(output))
			case "delete":
				fmt.Fprintf(builder, "  - %s = %s -> null\n", output.Name, f.formatOutputValue(output))
			}
		} else {
			fmt.Fprintf(builder, "  ~ %s = %s -> %s\n", output.Name, f.formatOutputValue(output), f.formatOutputValue(output))
		}
	}
	builder.WriteString("\n")
}

// formatOutputValue formats an output value for display
func (f *PlanFormatter) formatOutputValue(output core.OutputSummary) string {
	if output.Sensitive {
		return "(sensitive value)"
	}
	if output.Value == nil {
		return "null"
	}
	return fmt.Sprintf("%v", output.Value)
}

// writeFooter writes the terraform plan footer
func (f *PlanFormatter) writeFooter(builder *strings.Builder) {
	builder.WriteString("──────────────────────────────────────────────────────────────────────────────────────────────\n\n")
	builder.WriteString("Note: You didn't use the -out option to save this plan, so Terraform can't guarantee to take\n")
	builder.WriteString("exactly these actions if you run \"terraform apply\" now.\n")
}
