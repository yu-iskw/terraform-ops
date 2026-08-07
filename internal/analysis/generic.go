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

package analysis

import (
	"context"
	"fmt"

	"github.com/yu/terraform-ops/internal/ir"
	"github.com/yu/terraform-ops/internal/report"
)

type planAnalyzer struct{}

func (planAnalyzer) ID() string { return "plan" }
func (planAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	if cs.Plan.Errored {
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-PLAN-ERRORED",
			Title:      "Terraform/OpenTofu planning errored",
			Category:   report.CategoryPlan,
			Severity:   report.SeverityHigh,
			Confidence: report.ConfidenceExact,
			Message:    "Planning reported an error; the available change set may be partial and cannot be safely treated as a complete plan.",
		})
	}
	if !cs.Plan.Complete {
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-PLAN-INCOMPLETE",
			Title:      "Plan is incomplete",
			Category:   report.CategoryUncertainty,
			Severity:   report.SeverityMedium,
			Confidence: report.ConfidenceExact,
			Message:    "The plan is not complete; an additional plan/apply round may be required to converge.",
		})
	}
	return findings, nil
}

type checkAnalyzer struct{}

func (checkAnalyzer) ID() string { return "checks" }
func (checkAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	for _, check := range cs.Checks {
		if check.Status != ir.CheckFail && check.Status != ir.CheckError {
			continue
		}
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-CHECK-FAILED",
			Title:      "Terraform/OpenTofu check failed",
			Category:   report.CategoryValidation,
			Severity:   report.SeverityHigh,
			Confidence: report.ConfidenceExact,
			Resource:   &report.ResourceRef{Address: check.Address},
			Evidence: []report.Evidence{{
				Kind:        "check_status",
				Description: string(check.Status),
				Source:      "plan.checks",
			}},
			Message: fmt.Sprintf("Check %s has status %s.", check.Address, check.Status),
		})
	}
	return findings, nil
}

type lifecycleAnalyzer struct{}

func (lifecycleAnalyzer) ID() string { return "lifecycle" }
func (lifecycleAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	for _, resource := range cs.Resources {
		if resource.Mode != ir.ResourceModeManaged {
			continue
		}
		switch {
		case resource.Action.Semantic == ir.ActionDelete:
			findings = append(findings, report.Finding{
				RuleID:     "TFOPS-LIFECYCLE-DELETE",
				Title:      "Managed resource deletion",
				Category:   report.CategoryLifecycle,
				Severity:   report.SeverityMedium,
				Confidence: report.ConfidenceExact,
				Resource:   resourceRef(resource),
				Evidence: []report.Evidence{{
					Kind:        "action",
					Description: "delete",
					Source:      "resource_changes.change.actions",
				}},
				Message: "A managed resource is planned for deletion.",
			})
		case resource.Action.IsReplace():
			evidence := []report.Evidence{{
				Kind:        "replacement_order",
				Description: string(resource.Action.Semantic),
				Source:      "resource_changes.change.actions",
			}}
			for _, path := range resource.ReplacePaths {
				evidence = append(evidence, report.Evidence{
					Kind:   "replace_path",
					Path:   path.String(),
					Source: "resource_changes.change.replace_paths",
				})
			}
			if resource.ActionReason != "" {
				evidence = append(evidence, report.Evidence{
					Kind:        "action_reason",
					Description: resource.ActionReason,
					Source:      "resource_changes.action_reason",
				})
			}
			findings = append(findings, report.Finding{
				RuleID:     "TFOPS-LIFECYCLE-REPLACE",
				Title:      "Managed resource replacement",
				Category:   report.CategoryLifecycle,
				Severity:   report.SeverityMedium,
				Confidence: report.ConfidenceExact,
				Resource:   resourceRef(resource),
				Evidence:   evidence,
				Message:    "A managed resource is planned for replacement.",
			})
		}
	}
	return findings, nil
}

type driftAnalyzer struct{}

func (driftAnalyzer) ID() string { return "drift" }
func (driftAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	for _, drift := range cs.Drift {
		evidence := []report.Evidence{{
			Kind:        "resource_drift",
			Description: string(drift.Resource.Action.Semantic),
			Source:      "resource_drift",
		}}
		for _, relevant := range drift.Relevant {
			evidence = append(evidence, report.Evidence{
				Kind:   "relevant_attribute",
				Path:   relevant.Path.String(),
				Source: "relevant_attributes",
			})
		}
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-DRIFT-DETECTED",
			Title:      "External drift detected",
			Category:   report.CategoryDrift,
			Severity:   report.SeverityMedium,
			Confidence: report.ConfidenceExact,
			Resource:   resourceRef(drift.Resource),
			Evidence:   evidence,
			Message:    "The plan reports external drift for this resource; relevant attributes are correlation evidence, not proof of causation.",
		})
	}
	return findings, nil
}

type sensitivityAnalyzer struct{}

func (sensitivityAnalyzer) ID() string { return "sensitivity" }
func (sensitivityAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	for _, resource := range cs.Resources {
		if len(resource.SensitivePaths) == 0 || resource.Action.Semantic == ir.ActionNoOp {
			continue
		}
		evidence := make([]report.Evidence, 0, len(resource.SensitivePaths))
		for _, path := range resource.SensitivePaths {
			evidence = append(evidence, report.Evidence{
				Kind:   "sensitive_path",
				Path:   path.String(),
				Source: "before_sensitive/after_sensitive",
			})
		}
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-SENSITIVE-MUTATION",
			Title:      "Sensitive data participates in a resource change",
			Category:   report.CategorySensitivity,
			Severity:   report.SeverityInfo,
			Confidence: report.ConfidenceExact,
			Resource:   resourceRef(resource),
			Evidence:   evidence,
			Message:    "One or more Terraform/OpenTofu-sensitive paths participate in this resource change; values have been redacted before analysis output.",
		})
	}
	return findings, nil
}

type unknownAnalyzer struct{}

func (unknownAnalyzer) ID() string { return "unknown" }
func (unknownAnalyzer) Analyze(_ context.Context, cs *ir.ChangeSet) ([]report.Finding, error) {
	var findings []report.Finding
	for _, resource := range cs.Resources {
		if len(resource.UnknownPaths) == 0 || resource.Action.Semantic == ir.ActionNoOp {
			continue
		}
		evidence := make([]report.Evidence, 0, len(resource.UnknownPaths))
		for _, path := range resource.UnknownPaths {
			evidence = append(evidence, report.Evidence{
				Kind:   "unknown_after_path",
				Path:   path.String(),
				Source: "after_unknown",
			})
		}
		findings = append(findings, report.Finding{
			RuleID:     "TFOPS-UNKNOWN-AFTER",
			Title:      "Post-apply values remain unknown",
			Category:   report.CategoryUncertainty,
			Severity:   report.SeverityInfo,
			Confidence: report.ConfidenceExact,
			Resource:   resourceRef(resource),
			Evidence:   evidence,
			Message:    "One or more changed attributes are unknown until apply.",
		})
	}
	return findings, nil
}

func resourceRef(resource ir.ResourceChange) *report.ResourceRef {
	return &report.ResourceRef{Address: string(resource.Address), Type: resource.Type}
}
