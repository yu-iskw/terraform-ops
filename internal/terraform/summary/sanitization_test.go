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
	"github.com/yu/terraform-ops/internal/ir"
	terraformsource "github.com/yu/terraform-ops/internal/source/terraform"
)

func TestSummarizePlanCannotRecoverNestedSensitiveValues(t *testing.T) {
	const canary = "TFOPS_CHANGESET_SUMMARY_CANARY_4d831a"
	planJSON := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "resource_changes":[{
    "address":"test_resource.example",
    "mode":"managed",
    "type":"test_resource",
    "name":"example",
    "change":{
      "actions":["update"],
      "before":{"settings":{"password":"` + canary + `","name":"before"}},
      "after":{"settings":{"password":"` + canary + `","name":"after"}},
      "before_sensitive":{"settings":{"password":true}},
      "after_sensitive":{"settings":{"password":true}}
    }
  }],
  "output_changes":{"secret":{
    "change":{"actions":["update"],"after":"` + canary + `","after_sensitive":true}
  }},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	plan, err := terraformsource.ParseReader(strings.NewReader(planJSON), terraformsource.DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := terraformsource.Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}

	got, err := NewSummarizer().SummarizePlan(changeSet, core.SummaryOptions{ShowDetails: true})
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), canary) {
		t.Fatal("summary leaked a value removed at the ChangeSet sanitization boundary")
	}
	if len(got.Changes.Update) != 1 || !got.Changes.Update[0].Sensitive {
		t.Fatal("nested sensitivity evidence was not preserved")
	}
	if !strings.Contains(fmt.Sprint(got.Changes.Update[0].KeyChanges), "<redacted>") {
		t.Fatalf("expected redacted marker in safe key changes, got %#v", got.Changes.Update[0].KeyChanges)
	}
	if len(got.Outputs) != 1 || !got.Outputs[0].Sensitive || got.Outputs[0].Value != nil {
		t.Fatalf("sensitive output was not suppressed: %#v", got.Outputs)
	}
}
