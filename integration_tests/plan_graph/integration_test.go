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
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain ensures the plan JSON files are generated before running tests
func TestMain(m *testing.M) {
	// Generate the web-app plan JSON if it doesn't exist
	if _, err := os.Stat("web-app-plan.json"); os.IsNotExist(err) {
		fmt.Println("Generating web-app plan JSON for tests...")
		cmd := exec.Command("make", "web-app-plan-json")
		cmd.Dir = "."
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to generate web-app plan JSON: %v\n", err)
			os.Exit(1)
		}
	}

	// Generate the simple-random plan JSON if it doesn't exist
	if _, err := os.Stat("simple-random-plan.json"); os.IsNotExist(err) {
		fmt.Println("Generating simple-random plan JSON for tests...")
		cmd := exec.Command("make", "simple-random-plan-json")
		cmd.Dir = "."
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to generate simple-random plan JSON: %v\n", err)
			os.Exit(1)
		}
	}

	// Run the tests
	os.Exit(m.Run())
}

// TestPlanGraphCommandWebApp tests the plan-graph command with web-app workspace
func TestPlanGraphCommandWebApp(t *testing.T) {
	// Test with the dynamically generated web-app plan file
	planFile := "web-app-plan.json"

	// Ensure plan file exists
	_, err := os.Stat(planFile)
	require.NoError(t, err, "Web-app plan file should exist: %s", planFile)

	// Test all supported formats
	formats := []string{"graphviz", "mermaid", "plantuml"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("WebApp_Format_%s", format), func(t *testing.T) {
			testPlanGraphFormat(t, planFile, format, "web-app")
		})
	}
}

// TestPlanGraphCommandSimpleRandom tests the plan-graph command with simple-random workspace
func TestPlanGraphCommandSimpleRandom(t *testing.T) {
	// Test with the dynamically generated simple-random plan file
	planFile := "simple-random-plan.json"

	// Ensure plan file exists
	_, err := os.Stat(planFile)
	require.NoError(t, err, "Simple-random plan file should exist: %s", planFile)

	// Test all supported formats
	formats := []string{"graphviz", "mermaid", "plantuml"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("SimpleRandom_Format_%s", format), func(t *testing.T) {
			testPlanGraphFormat(t, planFile, format, "simple-random")
		})
	}
}

// TestPlanGraphCommandWithOutputFileWebApp tests the plan-graph command with output file for web-app
func TestPlanGraphCommandWithOutputFileWebApp(t *testing.T) {
	planFile := "web-app-plan.json"
	outputFile := "test_web_app_output.dot"

	// Clean up after test
	defer func() {
		if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove output file: %v", err)
		}
	}()

	cmd := exec.Command("../build/terraform-ops", "plan-graph",
		"--format", "graphviz",
		"--output", outputFile,
		planFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

	// Verify output file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "Output file should be created")

	// Read and validate the output file content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	validateGraphvizOutput(t, string(content), "web-app")
}

// TestPlanGraphCommandWithOutputFileSimpleRandom tests the plan-graph command with output file for simple-random
func TestPlanGraphCommandWithOutputFileSimpleRandom(t *testing.T) {
	planFile := "simple-random-plan.json"
	outputFile := "test_simple_random_output.dot"

	// Clean up after test
	defer func() {
		if err := os.Remove(outputFile); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove output file: %v", err)
		}
	}()

	cmd := exec.Command("../build/terraform-ops", "plan-graph",
		"--format", "graphviz",
		"--output", outputFile,
		planFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

	// Verify output file was created
	_, err = os.Stat(outputFile)
	require.NoError(t, err, "Output file should be created")

	// Read and validate the output file content
	content, err := os.ReadFile(outputFile)
	require.NoError(t, err)

	validateGraphvizOutput(t, string(content), "simple-random")
}

// TestPlanGraphCommandWithGroupingWebApp tests different grouping strategies for web-app
func TestPlanGraphCommandWithGroupingWebApp(t *testing.T) {
	planFile := "web-app-plan.json"
	groupingStrategies := []string{"module", "action", "resource_type"}

	for _, grouping := range groupingStrategies {
		t.Run(fmt.Sprintf("WebApp_GroupBy_%s", grouping), func(t *testing.T) {
			cmd := exec.Command("../build/terraform-ops", "plan-graph",
				"--format", "graphviz",
				"--group-by", grouping,
				planFile)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

			output := stdout.String()
			assert.NotEmpty(t, output, "Should generate output")

			// Basic validation that output contains expected elements
			assert.Contains(t, output, "digraph terraform_plan")
		})
	}
}

// TestPlanGraphCommandWithGroupingSimpleRandom tests different grouping strategies for simple-random
func TestPlanGraphCommandWithGroupingSimpleRandom(t *testing.T) {
	planFile := "simple-random-plan.json"
	groupingStrategies := []string{"module", "action", "resource_type"}

	for _, grouping := range groupingStrategies {
		t.Run(fmt.Sprintf("SimpleRandom_GroupBy_%s", grouping), func(t *testing.T) {
			cmd := exec.Command("../build/terraform-ops", "plan-graph",
				"--format", "graphviz",
				"--group-by", grouping,
				planFile)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

			output := stdout.String()
			assert.NotEmpty(t, output, "Should generate output")

			// Basic validation that output contains expected elements
			assert.Contains(t, output, "digraph terraform_plan")
		})
	}
}

// TestPlanGraphCommandWithOptionsWebApp tests various command line options for web-app
func TestPlanGraphCommandWithOptionsWebApp(t *testing.T) {
	planFile := "web-app-plan.json"
	testPlanGraphCommandWithOptions(t, planFile, "web-app")
}

// TestPlanGraphCommandWithOptionsSimpleRandom tests various command line options for simple-random
func TestPlanGraphCommandWithOptionsSimpleRandom(t *testing.T) {
	planFile := "simple-random-plan.json"
	testPlanGraphCommandWithOptions(t, planFile, "simple-random")
}

// testPlanGraphCommandWithOptions is a helper function to test command line options
func testPlanGraphCommandWithOptions(t *testing.T, planFile, workspace string) {
	testCases := []struct {
		name     string
		args     []string
		validate func(t *testing.T, output string)
	}{
		{
			name: "Compact_Output",
			args: []string{"--compact"},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "digraph terraform_plan")
			},
		},
		{
			name: "Verbose_Output",
			args: []string{"--verbose"},
			validate: func(t *testing.T, output string) {
				assert.Contains(t, output, "digraph terraform_plan")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", workspace, tc.name), func(t *testing.T) {
			args := append([]string{"plan-graph", "--format", "graphviz"}, tc.args...)
			args = append(args, planFile)

			cmd := exec.Command("../build/terraform-ops", args...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

			output := stdout.String()
			tc.validate(t, output)
		})
	}
}

// TestPlanGraphCommandErrorHandling tests error cases
func TestPlanGraphCommandErrorHandling(t *testing.T) {
	testCases := []struct {
		name        string
		args        []string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Missing_Plan_File",
			args:        []string{"plan-graph"},
			expectError: true,
			errorMsg:    "accepts 1 arg(s), received 0",
		},
		{
			name:        "Invalid_Format",
			args:        []string{"plan-graph", "--format", "invalid", "web-app-plan.json"},
			expectError: true,
			errorMsg:    "unsupported format",
		},
		{
			name:        "Invalid_Grouping",
			args:        []string{"plan-graph", "--group-by", "invalid", "web-app-plan.json"},
			expectError: true,
			errorMsg:    "unsupported grouping",
		},
		{
			name:        "Non_Existent_Plan_File",
			args:        []string{"plan-graph", "non-existent-plan.json"},
			expectError: true,
			errorMsg:    "failed to open plan file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("../build/terraform-ops", tc.args...)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			if tc.expectError {
				assert.Error(t, err)
				output := stderr.String()
				if output == "" {
					output = stdout.String()
				}
				assert.Contains(t, output, tc.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestPlanGraphCommandHelp tests the help command
func TestPlanGraphCommandHelp(t *testing.T) {
	cmd := exec.Command("../build/terraform-ops", "plan-graph", "--help")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "plan-graph help command failed: %s", stderr.String())

	output := stdout.String()
	assert.Contains(t, output, "Generate a visual graph representation")
	assert.Contains(t, output, "Supported output formats")
	assert.Contains(t, output, "graphviz")
	assert.Contains(t, output, "mermaid")
	assert.Contains(t, output, "plantuml")
}

// TestPlanGraphVisualizationToolsWebApp tests that generated graphs can be processed by visualization tools for web-app
func TestPlanGraphVisualizationToolsWebApp(t *testing.T) {
	planFile := "web-app-plan.json"
	testPlanGraphVisualizationTools(t, planFile, "web-app")
}

// TestPlanGraphVisualizationToolsSimpleRandom tests that generated graphs can be processed by visualization tools for simple-random
func TestPlanGraphVisualizationToolsSimpleRandom(t *testing.T) {
	planFile := "simple-random-plan.json"
	testPlanGraphVisualizationTools(t, planFile, "simple-random")
}

// testPlanGraphVisualizationTools is a helper function to test visualization tools
func testPlanGraphVisualizationTools(t *testing.T, planFile, workspace string) {
	testCases := []struct {
		name     string
		format   string
		testTool func(t *testing.T, output string)
	}{
		{
			name:   "Graphviz_Validation",
			format: "graphviz",
			testTool: func(t *testing.T, output string) {
				testGraphvizValidation(t, output)
			},
		},
		{
			name:   "Mermaid_Validation",
			format: "mermaid",
			testTool: func(t *testing.T, output string) {
				testMermaidValidation(t, output)
			},
		},
		{
			name:   "PlantUML_Validation",
			format: "plantuml",
			testTool: func(t *testing.T, output string) {
				testPlantUMLValidation(t, output)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s_%s", workspace, tc.name), func(t *testing.T) {
			cmd := exec.Command("../build/terraform-ops", "plan-graph",
				"--format", tc.format,
				planFile)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

			output := stdout.String()
			assert.NotEmpty(t, output, "Should generate output")

			// Test with visualization tool
			tc.testTool(t, output)
		})
	}
}

// testPlanGraphFormat tests a specific format
func testPlanGraphFormat(t *testing.T, planFile, format, workspace string) {
	cmd := exec.Command("../build/terraform-ops", "plan-graph",
		"--format", format,
		planFile)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

	output := stdout.String()
	assert.NotEmpty(t, output, "Should generate output")

	// Validate format-specific output
	switch format {
	case "graphviz":
		validateGraphvizOutput(t, output, workspace)
	case "mermaid":
		validateMermaidOutput(t, output, workspace)
	case "plantuml":
		validatePlantUMLOutput(t, output, workspace)
	}
}

// validateGraphvizOutput validates Graphviz format output
func validateGraphvizOutput(t *testing.T, output, workspace string) {
	assert.Contains(t, output, "digraph terraform_plan")
	assert.Contains(t, output, "rankdir=TB")
	assert.Contains(t, output, "subgraph cluster_")

	switch workspace {
	case "web-app":
		// Should contain expected resources from web-app plan (GCP)
		assert.Contains(t, output, "google_compute_network.main")
		assert.Contains(t, output, "google_compute_subnetwork.public")
		assert.Contains(t, output, "google_compute_firewall.web")
		assert.Contains(t, output, "google_compute_instance.web")
		assert.Contains(t, output, "google_sql_database_instance.main")
		assert.Contains(t, output, "google_sql_database.app")
		assert.Contains(t, output, "google_sql_user.app")

		// Should contain random resources
		assert.Contains(t, output, "random_id.deployment_id")
		assert.Contains(t, output, "random_string.session_token")
		assert.Contains(t, output, "random_password.app_secret")
		assert.Contains(t, output, "random_uuid.correlation_id")

		// Should contain module grouping
		assert.Contains(t, output, "root")
		assert.Contains(t, output, "module.app")
		assert.Contains(t, output, "module.network")
		assert.Contains(t, output, "module.app.module.database")
	case "simple-random":
		// Should contain expected resources from simple-random plan
		assert.Contains(t, output, "random_id.test_id")
		assert.Contains(t, output, "random_string.test_string")
		assert.Contains(t, output, "random_password.test_password")
		assert.Contains(t, output, "random_uuid.test_uuid")
		assert.Contains(t, output, "random_integer.test_integer")
		assert.Contains(t, output, "random_pet.test_pet")

		// Should contain root module grouping
		assert.Contains(t, output, "root")
	}

	// Should contain action types
	assert.Contains(t, output, "[CREATE]")
}

// validateMermaidOutput validates Mermaid format output
func validateMermaidOutput(t *testing.T, output, workspace string) {
	assert.Contains(t, output, "graph TB")
	assert.Contains(t, output, "subgraph")
	assert.Contains(t, output, "end")

	switch workspace {
	case "web-app":
		// Should contain expected resources from web-app plan (GCP)
		assert.Contains(t, output, "google_compute_network.main")
		assert.Contains(t, output, "google_compute_subnetwork.public")
		assert.Contains(t, output, "google_compute_firewall.web")
		assert.Contains(t, output, "google_compute_instance.web")
		assert.Contains(t, output, "google_sql_database_instance.main")
		assert.Contains(t, output, "google_sql_database.app")
		assert.Contains(t, output, "google_sql_user.app")

		// Should contain random resources
		assert.Contains(t, output, "random_id.deployment_id")
		assert.Contains(t, output, "random_string.session_token")
		assert.Contains(t, output, "random_password.app_secret")
		assert.Contains(t, output, "random_uuid.correlation_id")
	case "simple-random":
		// Should contain expected resources from simple-random plan
		assert.Contains(t, output, "random_id.test_id")
		assert.Contains(t, output, "random_string.test_string")
		assert.Contains(t, output, "random_password.test_password")
		assert.Contains(t, output, "random_uuid.test_uuid")
		assert.Contains(t, output, "random_integer.test_integer")
		assert.Contains(t, output, "random_pet.test_pet")
	}

	// Should contain action types
	assert.Contains(t, output, "[CREATE]")
}

// validatePlantUMLOutput validates PlantUML format output
func validatePlantUMLOutput(t *testing.T, output, workspace string) {
	assert.Contains(t, output, "@startuml")
	assert.Contains(t, output, "@enduml")
	assert.Contains(t, output, "package")

	switch workspace {
	case "web-app":
		// Should contain expected resources from web-app plan (GCP)
		assert.Contains(t, output, "google_compute_network.main")
		assert.Contains(t, output, "google_compute_subnetwork.public")
		assert.Contains(t, output, "google_compute_firewall.web")
		assert.Contains(t, output, "google_compute_instance.web")
		assert.Contains(t, output, "google_sql_database_instance.main")
		assert.Contains(t, output, "google_sql_database.app")
		assert.Contains(t, output, "google_sql_user.app")

		// Should contain random resources
		assert.Contains(t, output, "random_id.deployment_id")
		assert.Contains(t, output, "random_string.session_token")
		assert.Contains(t, output, "random_password.app_secret")
		assert.Contains(t, output, "random_uuid.correlation_id")
	case "simple-random":
		// Should contain expected resources from simple-random plan
		assert.Contains(t, output, "random_id.test_id")
		assert.Contains(t, output, "random_string.test_string")
		assert.Contains(t, output, "random_password.test_password")
		assert.Contains(t, output, "random_uuid.test_uuid")
		assert.Contains(t, output, "random_integer.test_integer")
		assert.Contains(t, output, "random_pet.test_pet")
	}

	// Should contain action types
	assert.Contains(t, output, "[CREATE]")
}

// testGraphvizValidation tests that Graphviz output can be processed by dot command
func testGraphvizValidation(t *testing.T, output string) {
	// Check if dot command is available
	_, err := exec.LookPath("dot")
	if err != nil {
		t.Skip("Graphviz dot command not available, skipping validation")
	}

	// Create temporary DOT file
	tmpFile, err := os.CreateTemp("", "test-graph-*.dot")
	require.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	_, err = tmpFile.WriteString(output)
	require.NoError(t, err)
	err = tmpFile.Close()
	require.NoError(t, err)

	// Test that dot can parse the file
	cmd := exec.Command("dot", "-Tsvg", "-o", "/dev/null", tmpFile.Name())
	err = cmd.Run()
	assert.NoError(t, err, "Graphviz dot command should be able to parse the generated DOT file")
}

// testMermaidValidation tests that Mermaid output follows the correct syntax
func testMermaidValidation(t *testing.T, output string) {
	// Basic Mermaid syntax validation
	lines := strings.Split(output, "\n")

	// Should start with graph declaration
	assert.True(t, strings.HasPrefix(strings.TrimSpace(lines[0]), "graph"),
		"Mermaid output should start with graph declaration")

	// Should contain subgraph declarations
	hasSubgraph := false
	for _, line := range lines {
		if strings.Contains(line, "subgraph") {
			hasSubgraph = true
			break
		}
	}
	assert.True(t, hasSubgraph, "Mermaid output should contain subgraph declarations")

	// Should contain end statements
	hasEnd := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "end" {
			hasEnd = true
			break
		}
	}
	assert.True(t, hasEnd, "Mermaid output should contain end statements")
}

// testPlantUMLValidation tests that PlantUML output follows the correct syntax
func testPlantUMLValidation(t *testing.T, output string) {
	// Basic PlantUML syntax validation
	lines := strings.Split(output, "\n")

	// Should start with @startuml
	hasStartUml := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "@startuml" {
			hasStartUml = true
			break
		}
	}
	assert.True(t, hasStartUml, "PlantUML output should start with @startuml")

	// Should end with @enduml
	hasEndUml := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "@enduml" {
			hasEndUml = true
			break
		}
	}
	assert.True(t, hasEndUml, "PlantUML output should end with @enduml")

	// Should contain package declarations
	hasPackage := false
	for _, line := range lines {
		if strings.Contains(line, "package") {
			hasPackage = true
			break
		}
	}
	assert.True(t, hasPackage, "PlantUML output should contain package declarations")
}

// TestPlanGraphWithComplexPlan tests with a more complex plan file
func TestPlanGraphWithComplexPlan(t *testing.T) {
	// Create a complex plan file for testing
	complexPlan := createComplexPlanFile(t)
	defer func() {
		if err := os.Remove(complexPlan); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove complex plan file: %v", err)
		}
	}()

	formats := []string{"graphviz", "mermaid", "plantuml"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("Complex_Plan_%s", format), func(t *testing.T) {
			cmd := exec.Command("../build/terraform-ops", "plan-graph",
				"--format", format,
				complexPlan)

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			require.NoError(t, err, "plan-graph command failed: %s", stderr.String())

			output := stdout.String()
			assert.NotEmpty(t, output, "Should generate output")

			// Validate that complex plan elements are present
			assert.Contains(t, output, "google_compute_network.main")
			assert.Contains(t, output, "google_compute_subnetwork.private")
			assert.Contains(t, output, "google_compute_subnetwork.public")
			assert.Contains(t, output, "module.ec2")
			assert.Contains(t, output, "module.rds")
		})
	}
}

// createComplexPlanFile creates a complex plan file for testing
func createComplexPlanFile(t *testing.T) string {
	plan := map[string]interface{}{
		"format_version": "1.0",
		"prior_state": map[string]interface{}{
			"version":           4,
			"terraform_version": "1.5.0",
			"serial":            1,
			"lineage":           "complex-lineage",
			"outputs":           map[string]interface{}{},
			"resources":         []interface{}{},
		},
		"planned_values": map[string]interface{}{
			"root_module": map[string]interface{}{
				"resources": []interface{}{
					map[string]interface{}{
						"address":        "google_compute_network.main",
						"mode":           "managed",
						"type":           "google_compute_network",
						"name":           "main",
						"index":          0,
						"provider_name":  "registry.terraform.io/hashicorp/google",
						"schema_version": 1,
						"values": map[string]interface{}{
							"auto_create_subnetworks": false,
						},
						"sensitive_values": map[string]interface{}{},
					},
					map[string]interface{}{
						"address":        "google_compute_subnetwork.private",
						"mode":           "managed",
						"type":           "google_compute_subnetwork",
						"name":           "private",
						"index":          0,
						"provider_name":  "registry.terraform.io/hashicorp/google",
						"schema_version": 1,
						"values": map[string]interface{}{
							"ip_cidr_range": "10.0.1.0/24",
							"network":       "${google_compute_network.main.self_link}",
						},
						"sensitive_values": map[string]interface{}{},
					},
					map[string]interface{}{
						"address":        "google_compute_subnetwork.public",
						"mode":           "managed",
						"type":           "google_compute_subnetwork",
						"name":           "public",
						"index":          0,
						"provider_name":  "registry.terraform.io/hashicorp/google",
						"schema_version": 1,
						"values": map[string]interface{}{
							"ip_cidr_range": "10.0.2.0/24",
							"network":       "${google_compute_network.main.self_link}",
						},
						"sensitive_values": map[string]interface{}{},
					},
				},
				"child_modules": []interface{}{
					map[string]interface{}{
						"address": "module.ec2",
						"resources": []interface{}{
							map[string]interface{}{
								"address":        "module.ec2.google_compute_instance.app",
								"mode":           "managed",
								"type":           "google_compute_instance",
								"name":           "app",
								"index":          0,
								"provider_name":  "registry.terraform.io/hashicorp/google",
								"schema_version": 1,
								"values": map[string]interface{}{
									"machine_type": "e2-micro",
									"subnetwork":   "${google_compute_subnetwork.private.self_link}",
								},
								"sensitive_values": map[string]interface{}{},
							},
						},
						"child_modules": []interface{}{},
					},
					map[string]interface{}{
						"address": "module.rds",
						"resources": []interface{}{
							map[string]interface{}{
								"address":        "module.rds.google_sql_database_instance.main",
								"mode":           "managed",
								"type":           "google_sql_database_instance",
								"name":           "main",
								"index":          0,
								"provider_name":  "registry.terraform.io/hashicorp/google",
								"schema_version": 1,
								"values": map[string]interface{}{
									"instance_id":      "main",
									"region":           "us-central1",
									"database_version": "POSTGRES_14",
									"settings": map[string]interface{}{
										"tier": "db-f1-micro",
									},
								},
								"sensitive_values": map[string]interface{}{
									"password": true,
								},
							},
						},
						"child_modules": []interface{}{},
					},
				},
			},
		},
		"resource_changes": []interface{}{
			map[string]interface{}{
				"address":        "google_compute_network.main",
				"module_address": "",
				"mode":           "managed",
				"type":           "google_compute_network",
				"name":           "main",
				"index":          0,
				"deposed":        nil,
				"actions":        []string{"create"},
				"before":         nil,
				"after": map[string]interface{}{
					"auto_create_subnetworks": false,
				},
				"after_unknown": map[string]interface{}{
					"id":  true,
					"arn": true,
				},
				"before_sensitive": map[string]interface{}{},
				"after_sensitive":  map[string]interface{}{},
				"replace_paths":    []interface{}{},
				"importing":        nil,
			},
			map[string]interface{}{
				"address":        "google_compute_subnetwork.private",
				"module_address": "",
				"mode":           "managed",
				"type":           "google_compute_subnetwork",
				"name":           "private",
				"index":          0,
				"deposed":        nil,
				"actions":        []string{"create"},
				"before":         nil,
				"after": map[string]interface{}{
					"ip_cidr_range": "10.0.1.0/24",
				},
				"after_unknown": map[string]interface{}{
					"id":  true,
					"arn": true,
				},
				"before_sensitive": map[string]interface{}{},
				"after_sensitive":  map[string]interface{}{},
				"replace_paths":    []interface{}{},
				"importing":        nil,
			},
			map[string]interface{}{
				"address":        "google_compute_subnetwork.public",
				"module_address": "",
				"mode":           "managed",
				"type":           "google_compute_subnetwork",
				"name":           "public",
				"index":          0,
				"deposed":        nil,
				"actions":        []string{"create"},
				"before":         nil,
				"after": map[string]interface{}{
					"ip_cidr_range": "10.0.2.0/24",
				},
				"after_unknown": map[string]interface{}{
					"id":  true,
					"arn": true,
				},
				"before_sensitive": map[string]interface{}{},
				"after_sensitive":  map[string]interface{}{},
				"replace_paths":    []interface{}{},
				"importing":        nil,
			},
			map[string]interface{}{
				"address":        "module.ec2.google_compute_instance.app",
				"module_address": "module.ec2",
				"mode":           "managed",
				"type":           "google_compute_instance",
				"name":           "app",
				"index":          0,
				"deposed":        nil,
				"actions":        []string{"create"},
				"before":         nil,
				"after": map[string]interface{}{
					"machine_type": "e2-micro",
				},
				"after_unknown": map[string]interface{}{
					"id":  true,
					"arn": true,
				},
				"before_sensitive": map[string]interface{}{},
				"after_sensitive":  map[string]interface{}{},
				"replace_paths":    []interface{}{},
				"importing":        nil,
			},
			map[string]interface{}{
				"address":        "module.rds.google_sql_database_instance.main",
				"module_address": "module.rds",
				"mode":           "managed",
				"type":           "google_sql_database_instance",
				"name":           "main",
				"index":          0,
				"deposed":        nil,
				"actions":        []string{"create"},
				"before":         nil,
				"after": map[string]interface{}{
					"instance_id":      "main",
					"region":           "us-central1",
					"database_version": "POSTGRES_14",
					"settings": map[string]interface{}{
						"tier": "db-f1-micro",
					},
				},
				"after_unknown": map[string]interface{}{
					"id":  true,
					"arn": true,
				},
				"before_sensitive": map[string]interface{}{},
				"after_sensitive": map[string]interface{}{
					"password": true,
				},
				"replace_paths": []interface{}{},
				"importing":     nil,
			},
		},
		"configuration": map[string]interface{}{},
	}

	data, err := json.MarshalIndent(plan, "", "  ")
	require.NoError(t, err)

	tmpFile, err := os.CreateTemp("", "complex-plan-*.json")
	require.NoError(t, err)

	_, err = tmpFile.Write(data)
	require.NoError(t, err)

	err = tmpFile.Close()
	require.NoError(t, err)

	return tmpFile.Name()
}
