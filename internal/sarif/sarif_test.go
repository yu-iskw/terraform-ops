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

package sarif

import (
	"encoding/json"
	"testing"

	"github.com/yu/terraform-ops/internal/report"
	"github.com/yu/terraform-ops/internal/sourceindex"
)

func TestRenderEmitsOnlyLocatedFindings(t *testing.T) {
	analysis := report.AnalysisReport{Tool: report.ToolMetadata{Name: "terraform-ops", Version: "v1.2.3"}}
	located := []sourceindex.LocatedFinding{{
		Finding: report.Finding{
			RuleID:     "TFOPS-LIFECYCLE-DELETE",
			Title:      "Managed resource deletion",
			Category:   report.CategoryLifecycle,
			Severity:   report.SeverityMedium,
			Confidence: report.ConfidenceExact,
			Resource:   &report.ResourceRef{Address: "terraform_data.example", Type: "terraform_data"},
			Message:    "A managed resource is planned for deletion.",
		},
		Location: sourceindex.Location{Path: "main.tf", StartLine: 4, StartColumn: 1, EndLine: 4, EndColumn: 44},
	}}

	data, err := Render(analysis, located)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	runs := document["runs"].([]any)
	run := runs[0].(map[string]any)
	results := run["results"].([]any)
	if len(results) != 1 {
		t.Fatalf("results = %d, want 1", len(results))
	}
	result := results[0].(map[string]any)
	if result["level"] != "warning" {
		t.Fatalf("level = %v, want warning", result["level"])
	}
	locations := result["locations"].([]any)
	physical := locations[0].(map[string]any)["physicalLocation"].(map[string]any)
	artifact := physical["artifactLocation"].(map[string]any)
	if artifact["uri"] != "main.tf" {
		t.Fatalf("artifact uri = %v", artifact["uri"])
	}
	region := physical["region"].(map[string]any)
	if region["startLine"] != float64(4) {
		t.Fatalf("startLine = %v", region["startLine"])
	}
}

func TestRenderAllowsEmptyResultsWhenNothingResolves(t *testing.T) {
	analysis := report.AnalysisReport{Tool: report.ToolMetadata{Name: "terraform-ops", Version: "v1.2.3"}}
	data, err := Render(analysis, nil)
	if err != nil {
		t.Fatal(err)
	}
	var document struct {
		Version string `json:"version"`
		Runs    []struct {
			Results []any `json:"results"`
		} `json:"runs"`
	}
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	if document.Version != "2.1.0" || len(document.Runs) != 1 || len(document.Runs[0].Results) != 0 {
		t.Fatalf("unexpected SARIF document: %#v", document)
	}
}
