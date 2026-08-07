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

package terraform

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/yu/terraform-ops/internal/ir"
)

const canary = "TFOPS_CANARY_SECRET_92f84d"

func TestNormalizeRedactsSensitiveValuesAndPreservesReplacementEvidence(t *testing.T) {
	planJSON := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "errored":false,
  "variables":{"password":{"value":"` + canary + `"}},
  "resource_changes":[{
    "address":"test_resource.db",
    "mode":"managed",
    "type":"test_resource",
    "name":"db",
    "action_reason":"replace_because_cannot_update",
    "change":{
      "actions":["delete","create"],
      "before":{"name":"db","password":"` + canary + `","disk":"old"},
      "after":{"name":"db","password":"` + canary + `","disk":"new"},
      "before_sensitive":{"password":true},
      "after_sensitive":{"password":true},
      "after_unknown":{"endpoint":true},
      "replace_paths":[["disk"]]
    }
  }],
  "resource_drift":[],
  "output_changes":{},
  "checks":[],
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	plan, err := ParseReader(strings.NewReader(planJSON), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}
	if !changeSet.Plan.Applyable {
		t.Fatal("applyable was not decoded")
	}
	if len(changeSet.Resources) != 1 {
		t.Fatalf("got %d resources", len(changeSet.Resources))
	}
	resource := changeSet.Resources[0]
	if !resource.Action.IsReplace() {
		t.Fatalf("unexpected action: %s", resource.Action.Semantic)
	}
	if got := resource.ReplacePaths[0].String(); got != "disk" {
		t.Fatalf("replace path = %q", got)
	}
	if got := resource.UnknownPaths[0].String(); got != "endpoint" {
		t.Fatalf("unknown path = %q", got)
	}
	if got := resource.SensitivePaths[0].String(); got != "password" {
		t.Fatalf("sensitive path = %q", got)
	}
	if changeSet.Redaction.VariableValuesRemoved != 1 {
		t.Fatalf("variables removed = %d", changeSet.Redaction.VariableValuesRemoved)
	}
	data, err := json.Marshal(changeSet)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), canary) {
		t.Fatal("sanitized change set leaked canary")
	}
}

func TestNormalizeBuildsDependencyBlastRadius(t *testing.T) {
	planJSON := `{
  "format_version":"1.0",
  "applyable":true,
  "complete":true,
  "errored":false,
  "resource_changes":[
    {"address":"test_resource.a","mode":"managed","type":"test_resource","name":"a","change":{"actions":["update"]}},
    {"address":"test_resource.b","mode":"managed","type":"test_resource","name":"b","change":{"actions":["update"]}},
    {"address":"test_resource.c","mode":"managed","type":"test_resource","name":"c","change":{"actions":["update"]}}
  ],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[
    {"address":"test_resource.a","mode":"managed","type":"test_resource","name":"a","expressions":{}},
    {"address":"test_resource.b","mode":"managed","type":"test_resource","name":"b","expressions":{"x":{"references":["test_resource.a.id"]}}},
    {"address":"test_resource.c","mode":"managed","type":"test_resource","name":"c","expressions":{"x":{"references":["test_resource.b"]}}}
  ],"module_calls":{},"outputs":{}}}
}`
	plan, err := ParseReader(strings.NewReader(planJSON), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}
	transitive := changeSet.Graph.TransitiveDependents(ir.NodeID("test_resource.a"))
	if len(transitive) != 2 || transitive[0] != "test_resource.b" || transitive[1] != "test_resource.c" {
		t.Fatalf("unexpected transitive dependents: %#v", transitive)
	}
}

func TestParseRejectsUnsupportedMajorVersion(t *testing.T) {
	_, err := ParseReader(strings.NewReader(`{"format_version":"2.0"}`), DefaultMaxPlanBytes)
	if err == nil || !strings.Contains(err.Error(), "only major version 1") {
		t.Fatalf("expected unsupported major version error, got %v", err)
	}
}
