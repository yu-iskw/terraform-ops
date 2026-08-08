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

package test

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"testing"
)

const integrationCanary = "TFOPS_INTEGRATION_SECRET_CANARY_7f3d91"

func TestAnalyzeCommandRedactsGeneratedPlan(t *testing.T) {
	planPath := "workspaces/simple/plan.json"
	planJSON, err := os.ReadFile(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(planJSON, []byte(integrationCanary)) {
		t.Fatal("generated plan does not contain the canary; the fixture cannot prove redaction")
	}

	cmd := exec.Command(
		"../../build/terraform-ops",
		"analyze",
		"--format", "json",
		"--engine", "terraform",
		planPath,
	)
	cmd.Dir = "."
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("analyze failed: %v\nstderr: %s", err, stderr.String())
	}
	if strings.Contains(stdout.String(), integrationCanary) {
		t.Fatal("analyze output leaked the sensitive integration-test canary")
	}

	var report struct {
		SchemaVersion string `json:"schema_version"`
		Tool          struct {
			Version string `json:"version"`
		} `json:"tool"`
		Source struct {
			Engine            string `json:"engine"`
			PlanFormatVersion string `json:"plan_format_version"`
		} `json:"source"`
		Plan struct {
			Complete bool `json:"complete"`
		} `json:"plan"`
		Changes []struct {
			Address        string   `json:"address"`
			SensitivePaths []string `json:"sensitive_paths"`
			BlastRadius    struct {
				DirectDependents     int `json:"direct_dependents"`
				TransitiveDependents int `json:"transitive_dependents"`
			} `json:"blast_radius"`
		} `json:"changes"`
		Redaction struct {
			SensitivePaths        int `json:"terraform_sensitive_paths"`
			VariableValuesRemoved int `json:"variable_values_removed"`
		} `json:"redaction"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode analysis report: %v\noutput: %s", err, stdout.String())
	}
	if report.SchemaVersion != "1.0" {
		t.Fatalf("unexpected report schema version %q", report.SchemaVersion)
	}
	if report.Tool.Version == "" || report.Tool.Version == "dev" {
		t.Fatalf("expected build-stamped tool version, got %q", report.Tool.Version)
	}
	if report.Source.Engine != "terraform" || report.Source.PlanFormatVersion == "" {
		t.Fatalf("unexpected source metadata: %#v", report.Source)
	}
	if !report.Plan.Complete {
		t.Fatal("generated converged plan was reported incomplete")
	}
	if report.Redaction.VariableValuesRemoved < 1 {
		t.Fatal("expected source variable values to be removed before analysis")
	}
	if report.Redaction.SensitivePaths < 1 {
		t.Fatal("expected at least one Terraform-sensitive path in the generated plan")
	}

	var source, child, consumer *struct {
		Address        string   `json:"address"`
		SensitivePaths []string `json:"sensitive_paths"`
		BlastRadius    struct {
			DirectDependents     int `json:"direct_dependents"`
			TransitiveDependents int `json:"transitive_dependents"`
		} `json:"blast_radius"`
	}
	for i := range report.Changes {
		switch report.Changes[i].Address {
		case "terraform_data.source":
			source = &report.Changes[i]
		case "module.child.terraform_data.child":
			child = &report.Changes[i]
		case "terraform_data.consumer":
			consumer = &report.Changes[i]
		}
	}
	if source == nil || child == nil || consumer == nil {
		t.Fatalf("expected source, child, and consumer changes, got %#v", report.Changes)
	}
	if len(source.SensitivePaths) == 0 {
		t.Fatal("source resource lost sensitive-path evidence")
	}
	if source.BlastRadius.DirectDependents < 1 || source.BlastRadius.TransitiveDependents < 2 {
		t.Fatalf("expected module input to propagate source blast radius, got %#v", source.BlastRadius)
	}
	if child.BlastRadius.DirectDependents < 1 || child.BlastRadius.TransitiveDependents < 1 {
		t.Fatalf("expected module output to propagate child blast radius, got %#v", child.BlastRadius)
	}
}
