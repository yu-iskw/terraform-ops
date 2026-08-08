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

func TestNormalizeCarriesSanitizedSummaryAndGraphMetadata(t *testing.T) {
	const secret = "TFOPS_COMMAND_IR_SECRET_7a91"
	planJSON := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "errored":false,
  "variables":{"prefix":{"value":"` + secret + `"}},
  "resource_changes":[{
    "address":"test_resource.source",
    "mode":"managed",
    "type":"test_resource",
    "name":"source",
    "change":{"actions":["update"],"before":{"name":"before"},"after":{"name":"after"}}
  }],
  "output_changes":{"secret":{
    "change":{
      "actions":["update"],
      "before":"` + secret + `",
      "after":"` + secret + `",
      "before_sensitive":true,
      "after_sensitive":true
    }
  }},
  "configuration":{"root_module":{
    "resources":[{
      "address":"test_resource.source",
      "mode":"managed",
      "type":"test_resource",
      "name":"source",
      "expressions":{"name":{"references":["var.prefix"]}}
    }],
    "module_calls":{},
    "outputs":{"secret":{"expression":{"references":["test_resource.source.id"]}}}
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

	if len(changeSet.Outputs) != 1 {
		t.Fatalf("outputs = %d, want 1", len(changeSet.Outputs))
	}
	output := changeSet.Outputs[0]
	if !output.After.Redacted || output.After.Value != nil || len(output.SensitivePaths) == 0 {
		t.Fatalf("sensitive output was not normalized safely: %#v", output)
	}
	if changeSet.Redaction.VariableValuesRemoved != 1 {
		t.Fatalf("variable values removed = %d, want 1", changeSet.Redaction.VariableValuesRemoved)
	}

	assertGraphNodeKind(t, changeSet.Graph, "var.prefix", ir.NodeKindVariable)
	assertGraphNodeKind(t, changeSet.Graph, "test_resource.source", ir.NodeKindResource)
	assertGraphNodeKind(t, changeSet.Graph, "output.secret", ir.NodeKindOutput)
	assertGraphEdge(t, changeSet.Graph, "var.prefix", "test_resource.source", ir.EdgeVariableReference, ir.ConfidenceExact)
	assertGraphEdge(t, changeSet.Graph, "test_resource.source", "output.secret", ir.EdgeOutputReference, ir.ConfidenceExact)

	encoded, err := json.Marshal(changeSet)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) {
		t.Fatal("normalized ChangeSet leaked variable/output secret")
	}
}

func assertGraphNodeKind(t *testing.T, graph ir.DependencyGraph, id string, want ir.NodeKind) {
	t.Helper()
	for _, node := range graph.Nodes {
		if node.ID == ir.NodeID(id) {
			if node.Kind != want {
				t.Fatalf("node %s kind = %s, want %s", id, node.Kind, want)
			}
			return
		}
	}
	t.Fatalf("missing graph node %s: %#v", id, graph.Nodes)
}
