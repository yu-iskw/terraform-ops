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
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSummarizePlanCommand(t *testing.T) {

	tests := []struct {
		name          string
		args          []string
		planFile      string
		expectedFile  string
		shouldFail    bool
		expectedError string
	}{
		{
			name:         "text format",
			args:         []string{"summarize-plan", "--format", "text"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/text.txt",
		},
		{
			name:         "json format",
			args:         []string{"summarize-plan", "--format", "json"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/json.json",
		},
		{
			name:         "markdown format",
			args:         []string{"summarize-plan", "--format", "markdown"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/markdown.md",
		},
		{
			name:         "table format",
			args:         []string{"summarize-plan", "--format", "table"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/table.txt",
		},
		{
			name:         "plan format",
			args:         []string{"summarize-plan", "--format", "plan"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/plan.txt",
		},
		{
			name:         "plan format with details",
			args:         []string{"summarize-plan", "--format", "plan", "--show-details"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/plan-with-details.txt",
		},
		{
			name:         "group by provider",
			args:         []string{"summarize-plan", "--group-by", "provider"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/group-by-provider.txt",
		},
		{
			name:         "group by module",
			args:         []string{"summarize-plan", "--group-by", "module"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/group-by-module.txt",
		},
		{
			name:         "compact output",
			args:         []string{"summarize-plan", "--compact"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/compact.txt",
		},
		{
			name:         "no sensitive data",
			args:         []string{"summarize-plan", "--no-sensitive"},
			planFile:     "workspaces/simple/plan.json",
			expectedFile: "expected/no-sensitive.txt",
		},
		{
			name:          "invalid plan file",
			args:          []string{"summarize-plan"},
			planFile:      "non-existent-plan.json",
			shouldFail:    true,
			expectedError: "failed to parse plan file",
		},
		{
			name:          "invalid format",
			args:          []string{"summarize-plan", "--format", "invalid"},
			planFile:      "workspaces/simple/plan.json",
			shouldFail:    true,
			expectedError: "unsupported format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Prepare command arguments
			args := append(tt.args, tt.planFile)
			cmd := exec.Command("../../build/terraform-ops", args...)
			cmd.Dir = "."

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tt.shouldFail {
				// Test should fail
				assert.Error(t, err, "Expected command to fail")
				if tt.expectedError != "" {
					errorOutput := stderr.String()
					assert.Contains(t, errorOutput, tt.expectedError, "Expected error message")
				}
				return
			}

			// Test should succeed
			require.NoError(t, err, "Command should succeed")
			actualOutput := stdout.String()

			// Map test name to format key
			formatKey := map[string]string{
				"text format":              "text_format",
				"json format":              "json_format",
				"markdown format":          "markdown_format",
				"table format":             "table_format",
				"plan format":              "plan_format",
				"plan format with details": "plan_format_with_details",
				"group by provider":        "group_by_provider",
				"group by module":          "group_by_module",
				"compact output":           "compact_output",
				"no sensitive data":        "no_sensitive_data",
			}
			key := formatKey[tt.name]
			checkOutputContains(t, actualOutput, key)
		})
	}
}

func TestSummarizePlanCommandWithOutputFile(t *testing.T) {

	outputFile := "test-output.txt"
	defer func() {
		_ = os.Remove(outputFile)
	}()

	// Run command with output file
	cmd := exec.Command("../../build/terraform-ops", "summarize-plan", "--format", "text", "--output", outputFile, "workspaces/simple/plan.json")
	cmd.Dir = "."

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "Command should succeed")

	// Check that output file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "Output file should be created")

	// Read the output file
	actualContent, err := os.ReadFile(outputFile)
	require.NoError(t, err, "Should be able to read output file")

	// Check that the output file contains key elements
	checkOutputContains(t, string(actualContent), "text_format")
}

func TestSummarizePlanCommandHelp(t *testing.T) {

	// Test help output
	cmd := exec.Command("../../build/terraform-ops", "summarize-plan", "--help")
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "Help command should work")

	output := stdout.String()

	// Check that help includes all formats
	expectedFormats := []string{
		"text: Human-readable console output",
		"json: Machine-readable structured data",
		"markdown: GitHub-compatible markdown format",
		"table: Tabular format for easy parsing",
		"plan: Terraform plan-like output format",
	}

	for _, format := range expectedFormats {
		assert.Contains(t, output, format, "Help should mention %s", format)
	}

	// Check that help includes the plan format example
	assert.Contains(t, output, "terraform-ops summarize-plan --format plan plan.json", "Help should include plan format example")
}

// checkOutputContains verifies that the output contains expected key elements
func checkOutputContains(t *testing.T, output, testName string) {
	// Format-specific elements
	formatElements := map[string][]string{
		"text_format": {
			"Terraform Plan Summary",
			"Plan Status:",
			"Format Version:",
			"Total Changes:",
			"📊 Statistics",
			"🔄 Resource Changes",
			"📤 Output Changes",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
		"json_format": {
			"\"plan_info\":",
			"\"statistics\":",
			"\"changes\":",
			"\"random_id.test_id\"",
			"\"random_string.test_string\"",
			"\"random_password.test_password\"",
			"\"random_pet.test_pet\"",
			"\"module.myrandom.random_integer.test_integer\"",
			"\"module.myrandom.random_string.test_string\"",
		},
		"markdown_format": {
			"# Terraform Plan Summary",
			"**Plan Status:**",
			"**Format Version:**",
			"**Total Changes:**",
			"## 📊 Statistics",
			"## 🔄 Resource Changes",
			"**random_id.test_id**",
			"**random_string.test_string**",
			"**random_password.test_password**",
			"**random_pet.test_pet**",
			"**module.myrandom.random_integer.test_integer**",
			"**module.myrandom.random_string.test_string**",
		},
		"table_format": {
			"## Statistics",
			"| Action | Count |",
			"| Provider | Count |",
			"| Module | Count |",
			"| Address | Type | Provider | Module | Sensitive |",
			"| Name | Actions | Sensitive | Value |",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
		"plan_format": {
			"Terraform will perform the following actions:",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
			"Plan:",
		},
		"plan_format_with_details": {
			"Terraform will perform the following actions:",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
			"Plan:",
		},
		"group_by_provider": {
			"Terraform Plan Summary",
			"Plan Status:",
			"Format Version:",
			"Total Changes:",
			"📊 Statistics",
			"🔄 Resource Changes",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
		"group_by_module": {
			"Terraform Plan Summary",
			"Plan Status:",
			"Format Version:",
			"Total Changes:",
			"📊 Statistics",
			"🔄 Resource Changes",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
		"compact_output": {
			"Terraform Plan Summary",
			"Plan Status:",
			"Format Version:",
			"Total Changes:",
			"📊 Statistics",
			"🔄 Resource Changes",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
		"no_sensitive_data": {
			"Terraform Plan Summary",
			"Plan Status:",
			"Format Version:",
			"Total Changes:",
			"📊 Statistics",
			"🔄 Resource Changes",
			"random_id.test_id",
			"random_string.test_string",
			"random_password.test_password",
			"random_pet.test_pet",
			"module.myrandom.random_integer.test_integer",
			"module.myrandom.random_string.test_string",
		},
	}

	// Check format-specific elements
	if elements, ok := formatElements[testName]; ok {
		for _, element := range elements {
			assert.Contains(t, output, element, "Output should contain %s for %s", element, testName)
		}
	} else {
		t.Errorf("No expected elements defined for test: %s", testName)
	}
}
