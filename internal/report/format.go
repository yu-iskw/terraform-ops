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

package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

type Format string

const (
	FormatText     Format = "text"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

func Render(report AnalysisReport, format Format) ([]byte, error) {
	switch format {
	case FormatJSON:
		return renderJSON(report)
	case FormatMarkdown:
		return []byte(renderMarkdown(report)), nil
	case FormatText, "":
		return []byte(renderText(report)), nil
	default:
		return nil, fmt.Errorf("unsupported analysis format %q", format)
	}
}

func renderJSON(report AnalysisReport) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func renderText(report AnalysisReport) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Terraform/OpenTofu Change Analysis")
	fmt.Fprintln(&b, "=================================")
	fmt.Fprintf(&b, "Engine: %s", report.Source.Engine)
	if report.Source.EngineVersion != "" {
		fmt.Fprintf(&b, " %s", report.Source.EngineVersion)
	}
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "Plan format: %s\n", report.Source.PlanFormatVersion)
	fmt.Fprintf(&b, "Applyable: %t  Complete: %t  Errored: %t\n\n", report.Plan.Applyable, report.Plan.Complete, report.Plan.Errored)
	fmt.Fprintf(&b, "Changes: +%d ~%d -%d replace:%d read:%d no-op:%d\n", report.Summary.Create, report.Summary.Update, report.Summary.Delete, report.Summary.Replace, report.Summary.Read, report.Summary.NoOp)
	fmt.Fprintf(&b, "Findings: critical:%d high:%d medium:%d low:%d info:%d\n", report.Summary.Findings.Critical, report.Summary.Findings.High, report.Summary.Findings.Medium, report.Summary.Findings.Low, report.Summary.Findings.Info)
	fmt.Fprintf(&b, "Graph: %d nodes, %d edges\n", report.Graph.Nodes, report.Graph.Edges)
	fmt.Fprintf(&b, "Redaction: %s (%d sensitive paths, %d variable values removed)\n", report.Redaction.Mode, report.Redaction.TerraformSensitivePaths, report.Redaction.VariableValuesRemoved)

	if len(report.Findings) > 0 {
		fmt.Fprintln(&b, "\nFindings")
		fmt.Fprintln(&b, "--------")
		for _, finding := range report.Findings {
			resource := ""
			if finding.Resource != nil {
				resource = " " + finding.Resource.Address
			}
			fmt.Fprintf(&b, "[%s] %s%s: %s\n", strings.ToUpper(string(finding.Severity)), finding.RuleID, resource, finding.Message)
			for _, evidence := range finding.Evidence {
				if evidence.Path != "" {
					fmt.Fprintf(&b, "  - %s: %s\n", evidence.Kind, evidence.Path)
				} else if evidence.Description != "" {
					fmt.Fprintf(&b, "  - %s: %s\n", evidence.Kind, evidence.Description)
				}
			}
		}
	}

	if len(report.Changes) > 0 {
		fmt.Fprintln(&b, "\nChanges")
		fmt.Fprintln(&b, "-------")
		for _, change := range report.Changes {
			fmt.Fprintf(&b, "%s  %s", change.Action, change.Address)
			if change.ActionReason != "" {
				fmt.Fprintf(&b, " (%s)", change.ActionReason)
			}
			fmt.Fprintln(&b)
			if len(change.ReplacePaths) > 0 {
				fmt.Fprintf(&b, "  replacement paths: %s\n", strings.Join(change.ReplacePaths, ", "))
			}
			if change.BlastRadius.DirectDependents > 0 || change.BlastRadius.TransitiveDependents > 0 {
				fmt.Fprintf(&b, "  blast radius: %d direct / %d transitive dependents\n", change.BlastRadius.DirectDependents, change.BlastRadius.TransitiveDependents)
			}
		}
	}
	return b.String()
}

func renderMarkdown(report AnalysisReport) string {
	var b strings.Builder
	fmt.Fprintln(&b, "## Terraform/OpenTofu change analysis")
	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "**Engine:** `%s`  \n", report.Source.Engine)
	fmt.Fprintf(&b, "**Plan:** applyable `%t`, complete `%t`, errored `%t`  \n", report.Plan.Applyable, report.Plan.Complete, report.Plan.Errored)
	fmt.Fprintf(&b, "**Changes:** +%d / ~%d / -%d / replace %d  \n", report.Summary.Create, report.Summary.Update, report.Summary.Delete, report.Summary.Replace)
	fmt.Fprintf(&b, "**Redaction:** `%s`; %d sensitive paths; %d variable values removed\n", report.Redaction.Mode, report.Redaction.TerraformSensitivePaths, report.Redaction.VariableValuesRemoved)

	if len(report.Findings) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "### Findings")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Severity | Rule | Resource | Finding |")
		fmt.Fprintln(&b, "|---|---|---|---|")
		for _, finding := range report.Findings {
			resource := ""
			if finding.Resource != nil {
				resource = "`" + escapeTable(finding.Resource.Address) + "`"
			}
			fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", strings.ToUpper(string(finding.Severity)), escapeTable(finding.RuleID), resource, escapeTable(finding.Message))
		}
	}

	if len(report.Changes) > 0 {
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "### Changes")
		fmt.Fprintln(&b)
		fmt.Fprintln(&b, "| Action | Resource | Why / replacement paths | Blast radius |")
		fmt.Fprintln(&b, "|---|---|---|---:|")
		for _, change := range report.Changes {
			why := change.ActionReason
			if len(change.ReplacePaths) > 0 {
				if why != "" {
					why += "; "
				}
				why += strings.Join(change.ReplacePaths, ", ")
			}
			fmt.Fprintf(&b, "| `%s` | `%s` | %s | %d direct / %d transitive |\n", change.Action, escapeTable(change.Address), escapeTable(why), change.BlastRadius.DirectDependents, change.BlastRadius.TransitiveDependents)
		}
	}
	return b.String()
}

func escapeTable(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "|", "\\|"), "\n", " ")
}
