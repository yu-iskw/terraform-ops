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
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/yu/terraform-ops/internal/analysis"
	"github.com/yu/terraform-ops/internal/version"
)

func TestAnalyzeCommandJSONDoesNotLeakSensitiveCanary(t *testing.T) {
	const canary = "TFOPS_COMMAND_CANARY_a91d3f"
	input := `{
  "format_version":"1.0",
  "applyable":true,
  "complete":true,
  "errored":false,
  "variables":{"secret":{"value":"` + canary + `"}},
  "resource_changes":[{
    "address":"test_resource.example",
    "mode":"managed",
    "type":"test_resource",
    "name":"example",
    "change":{
      "actions":["update"],
      "before":{"secret":"` + canary + `"},
      "after":{"secret":"` + canary + `"},
      "before_sensitive":{"secret":true},
      "after_sensitive":{"secret":true}
    }
  }],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(input), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:      "json",
		engine:      "terraform",
		redaction:   "standard",
		failOn:      "none",
		maxPlanSize: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(stdout.String(), canary) {
		t.Fatal("analysis output leaked canary")
	}
	if !strings.Contains(stdout.String(), `"rule_id": "TFOPS-SENSITIVE-MUTATION"`) {
		t.Fatal("expected sensitive mutation finding")
	}
}

func TestAnalyzeCommandReportsBuildVersion(t *testing.T) {
	originalVersion := version.Version
	version.Version = "v9.9.9-test"
	defer func() { version.Version = originalVersion }()

	input := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "errored":false,
  "resource_changes":[],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(input), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:      "json",
		engine:      "terraform",
		redaction:   "standard",
		failOn:      "none",
		maxPlanSize: 1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stdout.String(), `"version": "v9.9.9-test"`) {
		t.Fatalf("analysis report did not contain build version: %s", stdout.String())
	}
}

func TestAnalyzeCommandFailOnReturnsErrorAfterRendering(t *testing.T) {
	input := `{
  "format_version":"1.0",
  "applyable":true,
  "complete":true,
  "errored":false,
  "resource_changes":[{
    "address":"test_resource.example",
    "mode":"managed",
    "type":"test_resource",
    "name":"example",
    "change":{"actions":["delete"]}
  }],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(input), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:      "text",
		engine:      "terraform",
		redaction:   "standard",
		failOn:      "medium",
		maxPlanSize: 1 << 20,
	})
	if err == nil {
		t.Fatal("expected threshold error")
	}
	if stdout.Len() == 0 {
		t.Fatal("expected report to be rendered before threshold error")
	}
}
