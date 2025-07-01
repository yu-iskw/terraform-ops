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

package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/terraform/plan"
	"github.com/yu/terraform-ops/internal/terraform/summary"
	"github.com/yu/terraform-ops/internal/terraform/summary/formatters"
)

// SummarizePlanCommand represents the summarize-plan command with dependency injection
type SummarizePlanCommand struct {
	planParser       core.PlanParser
	planSummarizer   core.PlanSummarizer
	formatterFactory *formatters.Factory
}

// NewSummarizePlanCommand creates a new summarize-plan command with injected dependencies
func NewSummarizePlanCommand(
	planParser core.PlanParser,
	planSummarizer core.PlanSummarizer,
	formatterFactory *formatters.Factory,
) *SummarizePlanCommand {
	return &SummarizePlanCommand{
		planParser:       planParser,
		planSummarizer:   planSummarizer,
		formatterFactory: formatterFactory,
	}
}

// Command returns the cobra command for summarize-plan
func (c *SummarizePlanCommand) Command() *cobra.Command {
	var opts core.SummaryOptions

	cmd := &cobra.Command{
		Use:   "summarize-plan <PLAN_FILE>",
		Short: "Generate a human-readable summary of Terraform plan changes",
		Long: `Generate a human-readable summary of Terraform plan changes for the given workspace.
The summary provides a clear overview of all resource changes, organized by action type (create, update, delete, replace),
with statistics and breakdowns by provider, module, and resource type.

Supported output formats:
- text: Human-readable console output (default)
- json: Machine-readable structured data
- markdown: GitHub-compatible markdown format
- table: Tabular format for easy parsing
- plan: Terraform plan-like output format

Examples:
  terraform-ops summarize-plan plan.json
  terraform-ops summarize-plan --format markdown plan.json
  terraform-ops summarize-plan --format json plan.json
  terraform-ops summarize-plan --format plan plan.json
  terraform-ops summarize-plan --show-details plan.json
  terraform-ops summarize-plan --output summary.md plan.json
  terraform-ops summarize-plan --group-by provider plan.json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.runSummarizePlan(args[0], opts)
		},
	}

	// Add flags
	cmd.Flags().StringVarP((*string)(&opts.Format), "format", "f", string(core.FormatText), "Output format (text, json, markdown, table, plan)")
	cmd.Flags().StringVarP(&opts.Output, "output", "o", "", "Output file path (default: stdout)")
	cmd.Flags().StringVarP((*string)(&opts.GroupBy), "group-by", "g", string(core.GroupByAction), "Grouping strategy (action, module, provider, resource_type)")
	cmd.Flags().BoolVar(&opts.NoSensitive, "no-sensitive", false, "Hide sensitive value indicators")
	cmd.Flags().BoolVarP(&opts.Compact, "compact", "c", false, "Compact output format")
	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Enable verbose output for debugging")
	cmd.Flags().BoolVar(&opts.ShowDetails, "show-details", false, "Show detailed change information")
	cmd.Flags().StringVarP((*string)(&opts.Color), "color", "", string(core.ColorAuto), "Color output mode (auto, always, never)")

	return cmd
}

// runSummarizePlan executes the summarize-plan command
func (c *SummarizePlanCommand) runSummarizePlan(planFile string, opts core.SummaryOptions) error {
	// Validate format
	if !isValidSummaryFormat(opts.Format) {
		return fmt.Errorf("unsupported format: %s. Supported formats: text, json, markdown, table, plan", opts.Format)
	}

	// Validate grouping strategy
	if !isValidSummaryGrouping(opts.GroupBy) {
		return fmt.Errorf("unsupported grouping: %s. Supported groupings: action, module, provider, resource_type", opts.GroupBy)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Parsing plan file: %s\n", planFile)
	}

	// Parse the plan file
	plan, err := c.planParser.ParsePlanFile(planFile)
	if err != nil {
		return fmt.Errorf("failed to parse plan file: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d resource changes\n", len(plan.ResourceChanges))
	}

	// Generate summary
	summary, err := c.planSummarizer.SummarizePlan(plan, opts)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generated summary with %d total changes\n", summary.Statistics.TotalChanges)
	}

	// Determine color usage
	useColor := shouldUseColor(opts.Color)

	// Create formatter
	formatter, err := c.formatterFactory.CreateFormatter(opts.Format, useColor)
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}

	// Generate formatted output
	output, err := formatter.Format(summary, opts)
	if err != nil {
		return fmt.Errorf("failed to format summary: %w", err)
	}

	// Write output
	if opts.Output != "" {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", opts.Output)
		}
		if err := os.WriteFile(opts.Output, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Print(output)
	}

	return nil
}

// isValidSummaryFormat checks if the format is supported
func isValidSummaryFormat(format core.SummaryFormat) bool {
	switch format {
	case core.FormatText, core.FormatJSON, core.FormatMarkdown, core.FormatTable, core.FormatPlan:
		return true
	default:
		return false
	}
}

// isValidSummaryGrouping checks if the grouping strategy is supported
func isValidSummaryGrouping(grouping core.SummaryGrouping) bool {
	switch grouping {
	case core.GroupByAction, core.GroupByModule, core.GroupByProvider, core.GroupByResourceType:
		return true
	default:
		return false
	}
}

// shouldUseColor determines if color should be used based on the color mode
func shouldUseColor(colorMode core.ColorMode) bool {
	switch colorMode {
	case core.ColorAlways:
		return true
	case core.ColorNever:
		return false
	case core.ColorAuto:
		// Check if stdout is a terminal
		fileInfo, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	default:
		return false
	}
}

// DefaultSummarizePlanCommand creates a summarize-plan command with default dependencies
func DefaultSummarizePlanCommand() *SummarizePlanCommand {
	return NewSummarizePlanCommand(
		plan.NewParser(),
		summary.NewSummarizer(),
		formatters.NewFactory(),
	)
}
