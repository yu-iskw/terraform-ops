// Copyright 2025 yu-iskw
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
	"os/exec"
	"testing"

	"bytes"
	"encoding/json"
	"path/filepath"

	"github.com/stretchr/testify/assert"
)

func TestCLIVersion(t *testing.T) {
	// Build the CLI first
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = "/Users/yu/local/src/github/terraform-ops" // Explicitly set working directory
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", buildErr, string(buildOutput))
	}
	t.Logf("Build output: %s", string(buildOutput))

	// Test running the CLI
	cmd := exec.Command("../build/terraform-ops")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI execution failed: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected output from CLI, got empty response")
	}

	t.Logf("CLI output: %s", string(output))
}

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("../build/terraform-ops", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI help command failed: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected help output from CLI, got empty response")
	}

	t.Logf("CLI help output: %s", string(output))
}

func TestShowTerraformCommand(t *testing.T) {
	workspaces := []string{
		"workspaces/simple-providers",
		"workspaces/no-providers",
		"workspaces/gcs-backend",
		"workspaces/s3-backend",
	}
	args := append([]string{"show-terraform"}, workspaces...)
	cmd := exec.Command("../build/terraform-ops", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Logf("Stderr: %s", stderr.String())
		t.Fatalf("show-terraform command failed: %v", err)
	}

	type Result struct {
		Path      string `json:"path"`
		Terraform struct {
			RequiredVersion   string            `json:"required_version"`
			Backend           map[string]any    `json:"backend"`
			RequiredProviders map[string]string `json:"required_providers"`
		} `json:"terraform"`
	}

	var results []Result
	err = json.Unmarshal(stdout.Bytes(), &results)
	if err != nil {
		t.Fatalf("Failed to unmarshal JSON output: %v\nOutput:\n%s", err, stdout.String())
	}

	assert.Len(t, results, len(workspaces))

	// Create a map to easily look up results by workspace name
	resultsByPath := make(map[string]Result)
	for _, res := range results {
		resultsByPath[res.Path] = res
	}

	// Test specific workspace expectations
	for _, ws := range workspaces {
		absPath, err := filepath.Abs(ws)
		assert.NoError(t, err)

		res, exists := resultsByPath[absPath]
		assert.True(t, exists, "Expected workspace %s in results", ws)

		switch ws {
		case "workspaces/simple-providers":
			assert.Equal(t, map[string]string{
				"google": ">=4.83.0,<5.0.0",
				"aws":    "3.0.0",
			}, res.Terraform.RequiredProviders)
			assert.Nil(t, res.Terraform.Backend, "simple-providers should not have backend")

		case "workspaces/no-providers":
			assert.Empty(t, res.Terraform.RequiredProviders)
			assert.Nil(t, res.Terraform.Backend, "no-providers should not have backend")

		case "workspaces/gcs-backend":
			assert.Equal(t, ">= 1.0.0", res.Terraform.RequiredVersion)
			assert.Equal(t, map[string]string{
				"google": "~> 4.0",
			}, res.Terraform.RequiredProviders)
			if assert.NotNil(t, res.Terraform.Backend, "gcs-backend should have backend") {
				assert.Equal(t, "gcs", res.Terraform.Backend["type"])
				config := res.Terraform.Backend["config"].(map[string]any)
				assert.Equal(t, "terraform-state-prod", config["bucket"])
				assert.Equal(t, "terraform/state", config["prefix"])
			}

		case "workspaces/s3-backend":
			assert.Equal(t, ">= 1.0.0", res.Terraform.RequiredVersion)
			assert.Equal(t, map[string]string{
				"aws": "~> 5.0",
			}, res.Terraform.RequiredProviders)
			if assert.NotNil(t, res.Terraform.Backend, "s3-backend should have backend") {
				assert.Equal(t, "s3", res.Terraform.Backend["type"])
				config := res.Terraform.Backend["config"].(map[string]any)
				assert.Equal(t, "terraform-state-prod", config["bucket"])
				assert.Equal(t, "terraform/state.tfstate", config["key"])
				assert.Equal(t, "us-west-2", config["region"])
				assert.Equal(t, "true", config["encrypt"])
			}
		}
	}
}
