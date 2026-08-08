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
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/yu/terraform-ops/internal/analysis"
)

func TestAnalyzeCommandSARIFRequiresWorkspaceRoot(t *testing.T) {
	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(`{"format_version":"1.0"}`), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:      "sarif",
		engine:      "terraform",
		redaction:   "strict",
		failOn:      "none",
		maxPlanSize: 1 << 20,
	})
	if err == nil || !strings.Contains(err.Error(), "requires --workspace-root") {
		t.Fatalf("error = %v, want workspace-root requirement", err)
	}
}

func TestAnalyzeCommandSARIFMapsResourceToTerraformSource(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(`resource "terraform_data" "example" {
  input = "hello"
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
	input := `{
  "format_version":"1.0",
  "terraform_version":"1.15.8",
  "applyable":true,
  "complete":true,
  "errored":false,
  "resource_changes":[{
    "address":"terraform_data.example",
    "mode":"managed",
    "type":"terraform_data",
    "name":"example",
    "change":{"actions":["delete"]}
  }],
  "output_changes":{},
  "configuration":{"root_module":{"resources":[],"module_calls":{},"outputs":{}}}
}`
	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(input), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:        "sarif",
		engine:        "terraform",
		redaction:     "strict",
		failOn:        "none",
		workspaceRoot: root,
		maxPlanSize:   1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	output := stdout.String()
	if !strings.Contains(output, `"version": "2.1.0"`) {
		t.Fatalf("missing SARIF version: %s", output)
	}
	if !strings.Contains(output, `"ruleId": "TFOPS-LIFECYCLE-DELETE"`) {
		t.Fatalf("missing lifecycle finding: %s", output)
	}
	if !strings.Contains(output, `"uri": "main.tf"`) {
		t.Fatalf("missing source URI: %s", output)
	}
}
