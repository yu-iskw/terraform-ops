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

package analysis

import (
	"context"
	"testing"

	"github.com/yu/terraform-ops/internal/ir"
	"github.com/yu/terraform-ops/internal/report"
)

func TestDefaultRegistryProducesEvidenceBackedFindings(t *testing.T) {
	path := ir.AttributePath{ir.Attribute("disk_type")}
	cs := &ir.ChangeSet{
		Plan: ir.PlanMetadata{Applyable: true, Complete: true},
		Resources: []ir.ResourceChange{{
			Address:        "test_resource.db",
			Mode:           ir.ResourceModeManaged,
			Type:           "test_resource",
			Action:         ir.NormalizeAction([]string{"delete", "create"}),
			ActionReason:   "replace_because_cannot_update",
			ReplacePaths:   []ir.AttributePath{path},
			SensitivePaths: []ir.AttributePath{{ir.Attribute("password")}},
			UnknownPaths:   []ir.AttributePath{{ir.Attribute("endpoint")}},
		}},
	}
	findings, err := DefaultRegistry().Analyze(context.Background(), cs)
	if err != nil {
		t.Fatal(err)
	}
	if len(findings) != 3 {
		t.Fatalf("got %d findings: %#v", len(findings), findings)
	}
	if findings[0].RuleID != "TFOPS-LIFECYCLE-REPLACE" || findings[0].Severity != report.SeverityMedium {
		t.Fatalf("unexpected first finding: %#v", findings[0])
	}
	foundReplacePath := false
	for _, evidence := range findings[0].Evidence {
		if evidence.Kind == "replace_path" && evidence.Path == "disk_type" {
			foundReplacePath = true
		}
	}
	if !foundReplacePath {
		t.Fatal("replacement finding omitted replace_path evidence")
	}
}
