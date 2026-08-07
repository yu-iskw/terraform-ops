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

package summary

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/yu/terraform-ops/internal/core"
)

func TestSummarizePlanRedactsNestedSensitiveKeyChanges(t *testing.T) {
	const canary = "TFOPS_LEGACY_SUMMARY_CANARY_4d831a"
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{{
			Address: "test_resource.example",
			Mode:    "managed",
			Type:    "test_resource",
			Name:    "example",
			Change: core.Change{
				Actions: []string{"update"},
				Before: map[string]interface{}{
					"settings": map[string]interface{}{"password": canary, "name": "before"},
				},
				After: map[string]interface{}{
					"settings": map[string]interface{}{"password": canary, "name": "after"},
				},
				BeforeSensitive: map[string]interface{}{
					"settings": map[string]interface{}{"password": true},
				},
				AfterSensitive: map[string]interface{}{
					"settings": map[string]interface{}{"password": true},
				},
			},
		}},
		OutputChanges: map[string]core.OutputChange{},
	}

	summary, err := NewSummarizer().SummarizePlan(plan, core.SummaryOptions{ShowDetails: true})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), canary) {
		t.Fatal("legacy summary retained nested sensitive canary")
	}
	if len(summary.Changes.Update) != 1 || !summary.Changes.Update[0].Sensitive {
		t.Fatal("nested sensitivity was not detected")
	}
	keyChanges := fmt.Sprint(summary.Changes.Update[0].KeyChanges)
	if !strings.Contains(keyChanges, "<redacted>") {
		t.Fatalf("expected redacted marker in detailed key changes, got %s", keyChanges)
	}
}

func TestSummarizePlanTreatsBooleanSensitiveOutputMaskAsSensitive(t *testing.T) {
	const canary = "TFOPS_OUTPUT_CANARY_f02a77"
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges: map[string]core.OutputChange{
			"secret": {
				Change: core.Change{
					Actions:        []string{"update"},
					After:          canary,
					AfterSensitive: true,
				},
			},
		},
	}

	summary, err := NewSummarizer().SummarizePlan(plan, core.SummaryOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(summary.Outputs) != 1 || !summary.Outputs[0].Sensitive || summary.Outputs[0].Value != nil {
		t.Fatalf("sensitive output was not suppressed: %#v", summary.Outputs)
	}
	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), canary) {
		t.Fatal("sensitive output canary leaked")
	}
}
