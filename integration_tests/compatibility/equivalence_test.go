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

// TestSemanticEquivalence proves equality for change semantics represented by
// the overlapping Terraform/OpenTofu JSON contracts. Producer-specific source
// metadata, format-minor additions, Terraform-only applyable/complete metadata,
// value payloads, and sensitivity-mask propagation are intentionally not part of
// equality. Sensitivity masks are producer evidence rather than change intent;
// their cardinality differs across supported engines even when the resulting
// resource actions and dependency semantics are identical.
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
		assertStrictRedaction(t, name, changeSet)

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

func assertStrictRedaction(t *testing.T, name string, changeSet *ir.ChangeSet) {
	t.Helper()
	if changeSet.Redaction.Mode != ir.RedactionStrict {
		t.Fatalf("%s redaction mode = %q, want strict", name, changeSet.Redaction.Mode)
	}
	if changeSet.Redaction.VariableValuesRemoved != 1 {
		t.Fatalf("%s variable values removed = %d, want 1", name, changeSet.Redaction.VariableValuesRemoved)
	}
	if changeSet.Redaction.StrictValuesRemoved == 0 {
		t.Fatalf("%s did not remove any values in strict mode", name)
	}
	if len(changeSet.Resources) != 3 {
		t.Fatalf("%s resources = %d, want 3", name, len(changeSet.Resources))
	}
}

type semanticView struct {
	SchemaVersion   string              `json:"schema_version"`
	PlanErrored     bool                `json:"plan_errored"`
	PlanFormatMajor string              `json:"plan_format_major"`
	Resources       []resourceSemantic  `json:"resources,omitempty"`
	Outputs         []outputSemantic    `json:"outputs,omitempty"`
	Checks          []ir.CheckResult    `json:"checks,omitempty"`
	Drift           []resourceSemantic  `json:"drift,omitempty"`
	Relevant        []relevantSemantic  `json:"relevant,omitempty"`
	Graph           ir.DependencyGraph  `json:"graph"`
}

type resourceSemantic struct {
	Address         string         `json:"address"`
	PreviousAddress string         `json:"previous_address,omitempty"`
	ModuleAddress   string         `json:"module_address,omitempty"`
	Mode            ir.ResourceMode `json:"mode"`
	Type            string         `json:"type"`
	Name            string         `json:"name"`
	Index           string         `json:"index,omitempty"`
	DeposedKey      string         `json:"deposed_key,omitempty"`
	Action          ir.Action      `json:"action"`
	ActionReason    string         `json:"action_reason,omitempty"`
	ReplacePaths    []string       `json:"replace_paths,omitempty"`
	UnknownPaths    []string       `json:"unknown_paths,omitempty"`
	Import          *ir.ImportInfo `json:"import,omitempty"`
}

type outputSemantic struct {
	Name         string    `json:"name"`
	Action       ir.Action `json:"action"`
	UnknownPaths []string  `json:"unknown_paths,omitempty"`
}

type relevantSemantic struct {
	Resource string `json:"resource"`
	Path     string `json:"path"`
}

func overlappingSemantics(changeSet *ir.ChangeSet) semanticView {
	resources := make([]resourceSemantic, 0, len(changeSet.Resources))
	for _, resource := range changeSet.Resources {
		resources = append(resources, projectResource(resource))
	}
	outputs := make([]outputSemantic, 0, len(changeSet.Outputs))
	for _, output := range changeSet.Outputs {
		outputs = append(outputs, outputSemantic{
			Name:         output.Name,
			Action:       output.Action,
			UnknownPaths: pathStrings(output.UnknownPaths),
		})
	}
	drift := make([]resourceSemantic, 0, len(changeSet.Drift))
	for _, item := range changeSet.Drift {
		drift = append(drift, projectResource(item.Resource))
	}
	relevant := make([]relevantSemantic, 0, len(changeSet.Relevant))
	for _, item := range changeSet.Relevant {
		relevant = append(relevant, relevantSemantic{Resource: string(item.Resource), Path: item.Path.String()})
	}

	return semanticView{
		SchemaVersion:   changeSet.SchemaVersion,
		PlanErrored:     changeSet.Plan.Errored,
		PlanFormatMajor: formatMajor(changeSet.Source.PlanFormatVersion),
		Resources:       resources,
		Outputs:         outputs,
		Checks:          changeSet.Checks,
		Drift:           drift,
		Relevant:        relevant,
		Graph:           changeSet.Graph,
	}
}

func projectResource(resource ir.ResourceChange) resourceSemantic {
	return resourceSemantic{
		Address:         string(resource.Address),
		PreviousAddress: addressString(resource.PreviousAddress),
		ModuleAddress:   addressString(resource.ModuleAddress),
		Mode:            resource.Mode,
		Type:            resource.Type,
		Name:            resource.Name,
		Index:           resource.Index,
		DeposedKey:      resource.DeposedKey,
		Action:          resource.Action,
		ActionReason:    resource.ActionReason,
		ReplacePaths:    pathStrings(resource.ReplacePaths),
		UnknownPaths:    pathStrings(resource.UnknownPaths),
		Import:          resource.Import,
	}
}

func addressString(address *ir.Address) string {
	if address == nil {
		return ""
	}
	return string(*address)
}

func pathStrings(paths []ir.AttributePath) []string {
	if len(paths) == 0 {
		return nil
	}
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, path.String())
	}
	return out
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
