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

package compatibility

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/yu/terraform-ops/internal/ir"
	terraformsource "github.com/yu/terraform-ops/internal/source/terraform"
)

const compatibilityCanary = "compatibility-secret-canary"

// TestSemanticEquivalence proves equality only for semantics represented by the
// overlapping Terraform/OpenTofu JSON contracts. Producer-specific source
// metadata, format-minor additions, and Terraform-only applyable/complete
// metadata are intentionally not part of this comparison.
func TestSemanticEquivalence(t *testing.T) {
	artifactDir := os.Getenv("COMPAT_ARTIFACT_DIR")
	if artifactDir == "" {
		t.Skip("COMPAT_ARTIFACT_DIR is not set")
	}
	paths, err := filepath.Glob(filepath.Join(artifactDir, "*.plan.json"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(paths)
	if len(paths) < 2 {
		t.Fatalf("found %d compatibility artifacts, want at least two", len(paths))
	}

	var baseline semanticView
	var baselineName string
	seenTerraform := false
	seenOpenTofu := false
	for _, path := range paths {
		name := filepath.Base(path)
		engine := engineForArtifact(t, name)
		if engine == ir.EngineTerraform {
			seenTerraform = true
		} else if engine == ir.EngineOpenTofu {
			seenOpenTofu = true
		}

		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(raw), compatibilityCanary) {
			t.Fatalf("%s does not contain the raw secret canary; redaction test would be meaningless", name)
		}
		plan, err := terraformsource.ParseReader(strings.NewReader(string(raw)), terraformsource.DefaultMaxPlanBytes)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		changeSet, err := terraformsource.Normalize(plan, engine, ir.RedactionStrict)
		if err != nil {
			t.Fatalf("normalize %s: %v", name, err)
		}
		sanitized, err := json.Marshal(changeSet)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(sanitized), compatibilityCanary) {
			t.Fatalf("normalized %s leaked the secret canary", name)
		}

		view := overlappingSemantics(changeSet)
		if baselineName == "" {
			baselineName = name
			baseline = view
			continue
		}
		if !reflect.DeepEqual(baseline, view) {
			baselineJSON, _ := json.MarshalIndent(baseline, "", "  ")
			viewJSON, _ := json.MarshalIndent(view, "", "  ")
			t.Fatalf("semantic mismatch between %s and %s\n--- baseline ---\n%s\n--- candidate ---\n%s", baselineName, name, baselineJSON, viewJSON)
		}
	}
	if !seenTerraform || !seenOpenTofu {
		t.Fatalf("artifacts must include both engines: terraform=%v opentofu=%v", seenTerraform, seenOpenTofu)
	}
}

type semanticView struct {
	SchemaVersion   string                 `json:"schema_version"`
	PlanErrored     bool                   `json:"plan_errored"`
	PlanFormatMajor string                 `json:"plan_format_major"`
	Resources       []ir.ResourceChange    `json:"resources,omitempty"`
	Outputs         []ir.OutputChange      `json:"outputs,omitempty"`
	Checks          []ir.CheckResult       `json:"checks,omitempty"`
	Drift           []ir.DriftChange       `json:"drift,omitempty"`
	Relevant        []ir.RelevantAttribute `json:"relevant,omitempty"`
	Graph           ir.DependencyGraph     `json:"graph"`
	Redaction       ir.RedactionSummary    `json:"redaction"`
}

func overlappingSemantics(changeSet *ir.ChangeSet) semanticView {
	return semanticView{
		SchemaVersion:   changeSet.SchemaVersion,
		PlanErrored:     changeSet.Plan.Errored,
		PlanFormatMajor: formatMajor(changeSet.Source.PlanFormatVersion),
		Resources:       changeSet.Resources,
		Outputs:         changeSet.Outputs,
		Checks:          changeSet.Checks,
		Drift:           changeSet.Drift,
		Relevant:        changeSet.Relevant,
		Graph:           changeSet.Graph,
		Redaction:       changeSet.Redaction,
	}
}

func formatMajor(version string) string {
	major, _, _ := strings.Cut(version, ".")
	return major
}

func engineForArtifact(t *testing.T, name string) ir.Engine {
	t.Helper()
	switch {
	case strings.HasPrefix(name, "terraform-"):
		return ir.EngineTerraform
	case strings.HasPrefix(name, "opentofu-"):
		return ir.EngineOpenTofu
	default:
		t.Fatalf("cannot infer engine from artifact name %q", name)
		return ir.EngineUnknown
	}
}
