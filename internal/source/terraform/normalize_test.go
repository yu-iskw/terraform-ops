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

func TestNormalizeTraversesModuleInputsAndOutputs(t *testing.T) {
	planJSON := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "errored":false,
  "resource_changes":[
    {"address":"test_resource.source","mode":"managed","type":"test_resource","name":"source","change":{"actions":["update"]}},
    {"address":"module.child.test_resource.inner","module_address":"module.child","mode":"managed","type":"test_resource","name":"inner","change":{"actions":["update"]}},
    {"address":"test_resource.consumer","mode":"managed","type":"test_resource","name":"consumer","change":{"actions":["update"]}}
  ],
  "output_changes":{},
  "configuration":{"root_module":{
    "resources":[
      {"address":"test_resource.source","mode":"managed","type":"test_resource","name":"source","expressions":{}},
      {"address":"test_resource.consumer","mode":"managed","type":"test_resource","name":"consumer","expressions":{"value":{"references":["module.child.result","module.child"]}}}
    ],
    "module_calls":{"child":{
      "expressions":{"source_id":{"references":["test_resource.source.id","test_resource.source"]}},
      "module":{
        "resources":[
          {"address":"test_resource.inner","mode":"managed","type":"test_resource","name":"inner","expressions":{"value":{"references":["var.source_id"]}}}
        ],
        "module_calls":{},
        "outputs":{"result":{"expression":{"references":["test_resource.inner.id","test_resource.inner"]}}}
      }
    }},
    "outputs":{}
  }}
}`
	plan, err := ParseReader(strings.NewReader(planJSON), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}

	assertGraphEdge(t, changeSet.Graph, "test_resource.source", "module.child.test_resource.inner", ir.EdgeModuleInput, ir.ConfidenceExact)
	assertGraphEdge(t, changeSet.Graph, "module.child.test_resource.inner", "test_resource.consumer", ir.EdgeModuleOutput, ir.ConfidenceExact)

	transitive := changeSet.Graph.TransitiveDependents(ir.NodeID("test_resource.source"))
	if len(transitive) != 2 || transitive[0] != "module.child.test_resource.inner" || transitive[1] != "test_resource.consumer" {
		t.Fatalf("unexpected module transitive dependents: %#v", transitive)
	}
}

func TestNormalizeDefaultsMissingCompleteForTerraform17(t *testing.T) {
	plan, err := ParseReader(strings.NewReader(`{
  "format_version":"1.0",
  "terraform_version":"1.7.5",
  "applyable":true,
  "errored":false,
  "resource_changes":[],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}
	if !changeSet.Plan.Complete {
		t.Fatal("Terraform 1.7 plan without complete metadata must default to complete")
	}
}

func TestNormalizePreservesExplicitIncomplete(t *testing.T) {
	plan, err := ParseReader(strings.NewReader(`{
  "format_version":"1.0",
  "terraform_version":"1.7.5",
  "applyable":true,
  "complete":false,
  "errored":false,
  "resource_changes":[],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}
	if changeSet.Plan.Complete {
		t.Fatal("explicit incomplete metadata must be preserved")
	}
}

func TestNormalizeKeepsMissingCompleteConservativeForTerraform18(t *testing.T) {
	plan, err := ParseReader(strings.NewReader(`{
  "format_version":"1.0",
  "terraform_version":"1.8.0",
  "applyable":true,
  "errored":false,
  "resource_changes":[],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatal(err)
	}
	changeSet, err := Normalize(plan, ir.EngineTerraform, ir.RedactionStandard)
	if err != nil {
		t.Fatal(err)
	}
	if changeSet.Plan.Complete {
		t.Fatal("missing complete metadata on Terraform 1.8+ must remain conservative")
	}
}

func TestParseRejectsUnsupportedMajorVersion(t *testing.T) {
	_, err := ParseReader(strings.NewReader(`{"format_version":"2.0"}`), DefaultMaxPlanBytes)
	if err == nil || !strings.Contains(err.Error(), "only major version 1") {
		t.Fatalf("expected unsupported major version error, got %v", err)
	}
}

func assertGraphEdge(t *testing.T, graph ir.DependencyGraph, from, to string, kind ir.EdgeKind, confidence ir.EvidenceConfidence) {
	t.Helper()
	for _, edge := range graph.Edges {
		if edge.From == ir.NodeID(from) && edge.To == ir.NodeID(to) && edge.Kind == kind {
			if edge.Confidence != confidence {
				t.Fatalf("edge %s -> %s (%s) confidence = %s, want %s", from, to, kind, edge.Confidence, confidence)
			}
			return
		}
	}
	t.Fatalf("missing edge %s -> %s (%s): %#v", from, to, kind, graph.Edges)
}
