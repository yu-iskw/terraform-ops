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
	"github.com/yu/terraform-ops/internal/ir"
	terraformsource "github.com/yu/terraform-ops/internal/source/terraform"
	"github.com/yu/terraform-ops/internal/terraform/summary"
	"github.com/yu/terraform-ops/internal/terraform/summary/formatters"
)

// SummarizePlanCommand projects a normalized ChangeSet into legacy-compatible
// summary renderers. Raw plan DTOs never cross into the command's domain logic.
type SummarizePlanCommand struct {
	planSummarizer   core.PlanSummarizer
	formatterFactory *formatters.Factory
}

func NewSummarizePlanCommand(
	planSummarizer core.PlanSummarizer,
	formatterFactory *formatters.Factory,
) *SummarizePlanCommand {
	return &SummarizePlanCommand{
		planSummarizer:   planSummarizer,
		formatterFactory: formatterFactory,
	}
}

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

func (c *SummarizePlanCommand) runSummarizePlan(planFile string, opts core.SummaryOptions) error {
	if !isValidSummaryFormat(opts.Format) {
		return fmt.Errorf("unsupported format: %s. Supported formats: text, json, markdown, table, plan", opts.Format)
	}
	if !isValidSummaryGrouping(opts.GroupBy) {
		return fmt.Errorf("unsupported grouping: %s. Supported groupings: action, module, provider, resource_type", opts.GroupBy)
	}

	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Loading normalized plan: %s\n", planFile)
	}
	changeSet, err := terraformsource.LoadFile(
		planFile,
		terraformsource.DefaultMaxPlanBytes,
		ir.EngineUnknown,
		ir.RedactionStandard,
	)
	if err != nil {
		// Preserve the established CLI error prefix while the implementation now
		// parses and normalizes through the shared source adapter.
		return fmt.Errorf("failed to parse plan file: %w", err)
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Found %d resource changes\n", len(changeSet.Resources))
	}

	planSummary, err := c.planSummarizer.SummarizePlan(changeSet, opts)
	if err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}
	if opts.Verbose {
		fmt.Fprintf(os.Stderr, "Generated summary with %d total changes\n", planSummary.Statistics.TotalChanges)
	}

	formatter, err := c.formatterFactory.CreateFormatter(opts.Format, shouldUseColor(opts.Color))
	if err != nil {
		return fmt.Errorf("failed to create formatter: %w", err)
	}
	output, err := formatter.Format(planSummary, opts)
	if err != nil {
		return fmt.Errorf("failed to format summary: %w", err)
	}

	if opts.Output != "" {
		if opts.Verbose {
			fmt.Fprintf(os.Stderr, "Writing output to: %s\n", opts.Output)
		}
		if err := os.WriteFile(opts.Output, []byte(output), 0o644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
	} else {
		fmt.Print(output)
	}
	return nil
}

func isValidSummaryFormat(format core.SummaryFormat) bool {
	switch format {
	case core.FormatText, core.FormatJSON, core.FormatMarkdown, core.FormatTable, core.FormatPlan:
		return true
	default:
		return false
	}
}

func isValidSummaryGrouping(grouping core.SummaryGrouping) bool {
	switch grouping {
	case core.GroupByAction, core.GroupByModule, core.GroupByProvider, core.GroupByResourceType:
		return true
	default:
		return false
	}
}

func shouldUseColor(colorMode core.ColorMode) bool {
	switch colorMode {
	case core.ColorAlways:
		return true
	case core.ColorNever:
		return false
	case core.ColorAuto:
		fileInfo, err := os.Stdout.Stat()
		if err != nil {
			return false
		}
		return (fileInfo.Mode() & os.ModeCharDevice) != 0
	default:
		return false
	}
}

func DefaultSummarizePlanCommand() *SummarizePlanCommand {
	return NewSummarizePlanCommand(
		summary.NewSummarizer(),
		formatters.NewFactory(),
	)
}
