// Copyright 2026 yu-iskw
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
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/yu/terraform-ops/internal/analysis"
	"github.com/yu/terraform-ops/internal/ir"
	"github.com/yu/terraform-ops/internal/report"
	"github.com/yu/terraform-ops/internal/sarif"
	terraformsource "github.com/yu/terraform-ops/internal/source/terraform"
	"github.com/yu/terraform-ops/internal/sourceindex"
	"github.com/yu/terraform-ops/internal/version"
)

type AnalyzeCommand struct {
	registry *analysis.Registry
	stdin    io.Reader
	stdout   io.Writer
}

type analyzeOptions struct {
	format        string
	engine        string
	redaction     string
	failOn        string
	output        string
	workspaceRoot string
	maxPlanSize   int64
}

type analysisOutputFormat string

const (
	analysisFormatText     analysisOutputFormat = "text"
	analysisFormatJSON     analysisOutputFormat = "json"
	analysisFormatMarkdown analysisOutputFormat = "markdown"
	analysisFormatSARIF    analysisOutputFormat = "sarif"
)

func NewAnalyzeCommand(registry *analysis.Registry, stdin io.Reader, stdout io.Writer) *AnalyzeCommand {
	return &AnalyzeCommand{registry: registry, stdin: stdin, stdout: stdout}
}

func DefaultAnalyzeCommand() *AnalyzeCommand {
	return NewAnalyzeCommand(analysis.DefaultRegistry(), os.Stdin, os.Stdout)
}

func (c *AnalyzeCommand) Command() *cobra.Command {
	opts := analyzeOptions{}
	cmd := &cobra.Command{
		Use:   "analyze <PLAN_JSON>",
		Short: "Analyze Terraform/OpenTofu changes, causes, uncertainty, and blast radius",
		Long: `Analyze a Terraform/OpenTofu JSON plan without executing Terraform/OpenTofu.

Raw plan values are sanitized before they enter the normalized analysis model. Use "-"
to read a plan JSON document from stdin. SARIF output is source-aware and requires
--workspace-root so terraform-ops can map resource addresses to exact local .tf blocks.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.run(cmd.Context(), args[0], opts)
		},
	}
	cmd.Flags().StringVarP(&opts.format, "format", "f", string(analysisFormatText), "Output format (text, json, markdown, sarif)")
	cmd.Flags().StringVar(&opts.engine, "engine", "auto", "Source engine (auto, terraform, opentofu)")
	cmd.Flags().StringVar(&opts.redaction, "redaction", string(ir.RedactionStandard), "Redaction mode (standard, strict)")
	cmd.Flags().StringVar(&opts.failOn, "fail-on", "none", "Fail when a finding meets the severity threshold (none, info, low, medium, high, critical)")
	cmd.Flags().StringVarP(&opts.output, "output", "o", "", "Write the rendered report to a file instead of stdout")
	cmd.Flags().StringVar(&opts.workspaceRoot, "workspace-root", "", "Terraform workspace root used for exact source mapping (required for SARIF)")
	cmd.Flags().Int64Var(&opts.maxPlanSize, "max-plan-bytes", terraformsource.DefaultMaxPlanBytes, "Maximum accepted plan JSON size in bytes")
	return cmd
}

func (c *AnalyzeCommand) run(ctx context.Context, planPath string, opts analyzeOptions) error {
	engine, err := parseEngine(opts.engine)
	if err != nil {
		return err
	}
	redaction, err := parseRedaction(opts.redaction)
	if err != nil {
		return err
	}
	format, err := parseAnalysisFormat(opts.format)
	if err != nil {
		return err
	}
	threshold, err := parseFailOn(opts.failOn)
	if err != nil {
		return err
	}
	if format == analysisFormatSARIF && opts.workspaceRoot == "" {
		return fmt.Errorf("SARIF output requires --workspace-root for exact source locations")
	}

	var plan *terraformsource.Plan
	if planPath == "-" {
		plan, err = terraformsource.ParseReader(c.stdin, opts.maxPlanSize)
	} else {
		plan, err = terraformsource.ParseFile(planPath, opts.maxPlanSize)
	}
	if err != nil {
		return err
	}

	changeSet, err := terraformsource.Normalize(plan, engine, redaction)
	if err != nil {
		return err
	}
	findings, err := c.registry.Analyze(ctx, changeSet)
	if err != nil {
		return err
	}
	analysisReport := report.Build(changeSet, findings, version.Version)

	var rendered []byte
	if format == analysisFormatSARIF {
		index, indexErr := sourceindex.Build(opts.workspaceRoot)
		if indexErr != nil {
			return fmt.Errorf("build source index: %w", indexErr)
		}
		located := sourceindex.LocateFindings(findings, index)
		rendered, err = sarif.Render(analysisReport, located)
	} else {
		rendered, err = report.Render(analysisReport, report.Format(format))
	}
	if err != nil {
		return err
	}

	if opts.output != "" {
		if err := os.WriteFile(opts.output, rendered, 0o600); err != nil {
			return fmt.Errorf("write analysis report: %w", err)
		}
	} else {
		if _, err := c.stdout.Write(rendered); err != nil {
			return fmt.Errorf("write analysis report: %w", err)
		}
	}

	if threshold != "" && report.MeetsThreshold(report.HighestSeverity(findings), threshold) {
		return &FindingThresholdError{Threshold: threshold, Highest: report.HighestSeverity(findings)}
	}
	return nil
}

type FindingThresholdError struct {
	Threshold report.Severity
	Highest   report.Severity
}

func (e *FindingThresholdError) Error() string {
	return fmt.Sprintf("analysis finding threshold %s exceeded (highest severity: %s)", e.Threshold, e.Highest)
}

func parseEngine(value string) (ir.Engine, error) {
	switch strings.ToLower(value) {
	case "auto":
		return ir.EngineUnknown, nil
	case "terraform":
		return ir.EngineTerraform, nil
	case "opentofu", "tofu":
		return ir.EngineOpenTofu, nil
	default:
		return "", fmt.Errorf("unsupported engine %q: use auto, terraform, or opentofu", value)
	}
}

func parseRedaction(value string) (ir.RedactionMode, error) {
	switch ir.RedactionMode(strings.ToLower(value)) {
	case ir.RedactionStandard:
		return ir.RedactionStandard, nil
	case ir.RedactionStrict:
		return ir.RedactionStrict, nil
	default:
		return "", fmt.Errorf("unsupported redaction mode %q: use standard or strict", value)
	}
}

func parseAnalysisFormat(value string) (analysisOutputFormat, error) {
	switch analysisOutputFormat(strings.ToLower(value)) {
	case analysisFormatText:
		return analysisFormatText, nil
	case analysisFormatJSON:
		return analysisFormatJSON, nil
	case analysisFormatMarkdown:
		return analysisFormatMarkdown, nil
	case analysisFormatSARIF:
		return analysisFormatSARIF, nil
	default:
		return "", fmt.Errorf("unsupported analysis format %q: use text, json, markdown, or sarif", value)
	}
}

func parseFailOn(value string) (report.Severity, error) {
	switch report.Severity(strings.ToLower(value)) {
	case "none":
		return "", nil
	case report.SeverityInfo:
		return report.SeverityInfo, nil
	case report.SeverityLow:
		return report.SeverityLow, nil
	case report.SeverityMedium:
		return report.SeverityMedium, nil
	case report.SeverityHigh:
		return report.SeverityHigh, nil
	case report.SeverityCritical:
		return report.SeverityCritical, nil
	default:
		return "", fmt.Errorf("unsupported fail-on threshold %q", value)
	}
}
