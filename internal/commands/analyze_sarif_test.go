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
	writeAnalyzeSource(t, root)

	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(analyzeSARIFPlanJSON), &stdout)
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
	assertAnalyzeSARIF(t, stdout.String(), "main.tf")
}

func TestAnalyzeCommandSARIFUsesRepositoryRelativeSourceRoot(t *testing.T) {
	repositoryRoot := t.TempDir()
	workspaceRoot := filepath.Join(repositoryRoot, "infra", "prod")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeAnalyzeSource(t, workspaceRoot)

	var stdout bytes.Buffer
	cmd := NewAnalyzeCommand(analysis.DefaultRegistry(), strings.NewReader(analyzeSARIFPlanJSON), &stdout)
	err := cmd.run(context.Background(), "-", analyzeOptions{
		format:        "sarif",
		engine:        "terraform",
		redaction:     "strict",
		failOn:        "none",
		workspaceRoot: workspaceRoot,
		sourceRoot:    repositoryRoot,
		maxPlanSize:   1 << 20,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertAnalyzeSARIF(t, stdout.String(), "infra/prod/main.tf")
}

func writeAnalyzeSource(t *testing.T, root string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, "main.tf"), []byte(`resource "terraform_data" "example" {
  input = "hello"
}
`), 0o600); err != nil {
		t.Fatal(err)
	}
}

func assertAnalyzeSARIF(t *testing.T, output, uri string) {
	t.Helper()
	if !strings.Contains(output, `"version": "2.1.0"`) {
		t.Fatalf("missing SARIF version: %s", output)
	}
	if !strings.Contains(output, `"ruleId": "TFOPS-LIFECYCLE-DELETE"`) {
		t.Fatalf("missing lifecycle finding: %s", output)
	}
	if !strings.Contains(output, `"uri": "`+uri+`"`) {
		t.Fatalf("missing source URI %q: %s", uri, output)
	}
}

const analyzeSARIFPlanJSON = `{
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
