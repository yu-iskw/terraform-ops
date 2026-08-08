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
	"sort"

	"github.com/yu/terraform-ops/internal/ir"
)

const SchemaVersion = "1.0"

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityLow      Severity = "low"
	SeverityMedium   Severity = "medium"
	SeverityHigh     Severity = "high"
	SeverityCritical Severity = "critical"
)

type Confidence string

const (
	ConfidenceExact     Confidence = "exact"
	ConfidenceStrong    Confidence = "strong"
	ConfidenceHeuristic Confidence = "heuristic"
	ConfidenceUnknown   Confidence = "unknown"
)

type Category string

const (
	CategoryPlan        Category = "plan"
	CategoryLifecycle   Category = "lifecycle"
	CategoryValidation  Category = "validation"
	CategoryDrift       Category = "drift"
	CategorySensitivity Category = "sensitivity"
	CategoryUncertainty Category = "uncertainty"
)

type ToolMetadata struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ResourceRef struct {
	Address string `json:"address"`
	Type    string `json:"type,omitempty"`
}

type Evidence struct {
	Kind        string `json:"kind"`
	Description string `json:"description,omitempty"`
	Path        string `json:"path,omitempty"`
	Source      string `json:"source,omitempty"`
}

type Finding struct {
	RuleID      string       `json:"rule_id"`
	Title       string       `json:"title"`
	Category    Category     `json:"category"`
	Severity    Severity     `json:"severity"`
	Confidence  Confidence   `json:"confidence"`
	Resource    *ResourceRef `json:"resource,omitempty"`
	Evidence    []Evidence   `json:"evidence,omitempty"`
	Message     string       `json:"message"`
	Remediation string       `json:"remediation,omitempty"`
}

type FindingCounts struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
}

type Summary struct {
	Create   int           `json:"create"`
	Update   int           `json:"update"`
	Delete   int           `json:"delete"`
	Replace  int           `json:"replace"`
	Read     int           `json:"read"`
	NoOp     int           `json:"no_op"`
	Unknown  int           `json:"unknown"`
	Findings FindingCounts `json:"findings"`
}

type BlastRadius struct {
	DirectDependents     int `json:"direct_dependents"`
	TransitiveDependents int `json:"transitive_dependents"`
}

type ChangeReport struct {
	Address         string      `json:"address"`
	PreviousAddress string      `json:"previous_address,omitempty"`
	Type            string      `json:"type"`
	Mode            string      `json:"mode"`
	Action          string      `json:"action"`
	RawActions      []string    `json:"raw_actions"`
	ActionReason    string      `json:"action_reason,omitempty"`
	ReplacePaths    []string    `json:"replace_paths,omitempty"`
	UnknownPaths    []string    `json:"unknown_paths,omitempty"`
	SensitivePaths  []string    `json:"sensitive_paths,omitempty"`
	BlastRadius     BlastRadius `json:"blast_radius"`
}

type DriftReport struct {
	Address  string   `json:"address"`
	Action   string   `json:"action"`
	Relevant []string `json:"relevant_attributes,omitempty"`
}

type CheckReport struct {
	Address string `json:"address"`
	Kind    string `json:"kind,omitempty"`
	Status  string `json:"status"`
}

type GraphSummary struct {
	Nodes int `json:"nodes"`
	Edges int `json:"edges"`
}

type AnalysisReport struct {
	SchemaVersion string              `json:"schema_version"`
	Tool          ToolMetadata        `json:"tool"`
	Source        ir.SourceMetadata   `json:"source"`
	Plan          ir.PlanMetadata     `json:"plan"`
	Summary       Summary             `json:"summary"`
	Findings      []Finding           `json:"findings,omitempty"`
	Changes       []ChangeReport      `json:"changes,omitempty"`
	Drift         []DriftReport       `json:"drift,omitempty"`
	Checks        []CheckReport       `json:"checks,omitempty"`
	Graph         GraphSummary        `json:"graph"`
	Redaction     ir.RedactionSummary `json:"redaction"`
}

func Build(changeSet *ir.ChangeSet, findings []Finding, toolVersion string) AnalysisReport {
	report := AnalysisReport{
		SchemaVersion: SchemaVersion,
		Tool:          ToolMetadata{Name: "terraform-ops", Version: toolVersion},
		Source:        changeSet.Source,
		Plan:          changeSet.Plan,
		Findings:      append([]Finding(nil), findings...),
		Graph: GraphSummary{
			Nodes: len(changeSet.Graph.Nodes),
			Edges: len(changeSet.Graph.Edges),
		},
		Redaction: changeSet.Redaction,
	}

	for _, resource := range changeSet.Resources {
		switch resource.Action.Semantic {
		case ir.ActionCreate:
			report.Summary.Create++
		case ir.ActionUpdate:
			report.Summary.Update++
		case ir.ActionDelete:
			report.Summary.Delete++
		case ir.ActionReplaceDestroyCreate, ir.ActionReplaceCreateDestroy:
			report.Summary.Replace++
		case ir.ActionRead:
			report.Summary.Read++
		case ir.ActionNoOp:
			report.Summary.NoOp++
		default:
			report.Summary.Unknown++
		}

		change := ChangeReport{
			Address:        string(resource.Address),
			Type:           resource.Type,
			Mode:           string(resource.Mode),
			Action:         string(resource.Action.Semantic),
			RawActions:     append([]string(nil), resource.Action.Raw...),
			ActionReason:   resource.ActionReason,
			ReplacePaths:   pathStrings(resource.ReplacePaths),
			UnknownPaths:   pathStrings(resource.UnknownPaths),
			SensitivePaths: pathStrings(resource.SensitivePaths),
			BlastRadius: BlastRadius{
				DirectDependents:     len(changeSet.Graph.DirectDependents(ir.NodeID(resource.Address))),
				TransitiveDependents: len(changeSet.Graph.TransitiveDependents(ir.NodeID(resource.Address))),
			},
		}
		if resource.PreviousAddress != nil {
			change.PreviousAddress = string(*resource.PreviousAddress)
		}
		report.Changes = append(report.Changes, change)
	}

	for _, drift := range changeSet.Drift {
		item := DriftReport{
			Address: string(drift.Resource.Address),
			Action:  string(drift.Resource.Action.Semantic),
		}
		for _, relevant := range drift.Relevant {
			item.Relevant = append(item.Relevant, relevant.Path.String())
		}
		sort.Strings(item.Relevant)
		report.Drift = append(report.Drift, item)
	}

	for _, check := range changeSet.Checks {
		report.Checks = append(report.Checks, CheckReport{
			Address: check.Address,
			Kind:    check.Kind,
			Status:  string(check.Status),
		})
	}

	for _, finding := range report.Findings {
		switch finding.Severity {
		case SeverityCritical:
			report.Summary.Findings.Critical++
		case SeverityHigh:
			report.Summary.Findings.High++
		case SeverityMedium:
			report.Summary.Findings.Medium++
		case SeverityLow:
			report.Summary.Findings.Low++
		case SeverityInfo:
			report.Summary.Findings.Info++
		}
	}

	Sort(&report)
	return report
}

func Sort(report *AnalysisReport) {
	sort.Slice(report.Findings, func(i, j int) bool {
		a, b := report.Findings[i], report.Findings[j]
		if severityRank(a.Severity) != severityRank(b.Severity) {
			return severityRank(a.Severity) > severityRank(b.Severity)
		}
		if a.RuleID != b.RuleID {
			return a.RuleID < b.RuleID
		}
		return resourceAddress(a.Resource) < resourceAddress(b.Resource)
	})
	sort.Slice(report.Changes, func(i, j int) bool { return report.Changes[i].Address < report.Changes[j].Address })
	sort.Slice(report.Drift, func(i, j int) bool { return report.Drift[i].Address < report.Drift[j].Address })
	sort.Slice(report.Checks, func(i, j int) bool { return report.Checks[i].Address < report.Checks[j].Address })
}

func HighestSeverity(findings []Finding) Severity {
	highest := SeverityInfo
	if len(findings) == 0 {
		return ""
	}
	for _, finding := range findings {
		if severityRank(finding.Severity) > severityRank(highest) {
			highest = finding.Severity
		}
	}
	return highest
}

func MeetsThreshold(severity, threshold Severity) bool {
	if threshold == "" {
		return false
	}
	return severityRank(severity) >= severityRank(threshold)
}

func severityRank(severity Severity) int {
	switch severity {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

func resourceAddress(resource *ResourceRef) string {
	if resource == nil {
		return ""
	}
	return resource.Address
}

func pathStrings(paths []ir.AttributePath) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, path.String())
	}
	sort.Strings(out)
	return out
}
