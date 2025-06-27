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

package plan_graph

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand()
	require.NotNil(t, cmd)
	assert.Equal(t, "plan-graph <PLAN_FILE>", cmd.Use)
	assert.Equal(t, "Generate a visual graph representation of Terraform plan changes", cmd.Short)
	assert.True(t, len(cmd.Long) > 0)
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		name   string
		format GraphFormat
		want   bool
	}{
		{
			name:   "valid graphviz format",
			format: FormatGraphviz,
			want:   true,
		},
		{
			name:   "valid mermaid format",
			format: FormatMermaid,
			want:   true,
		},
		{
			name:   "valid plantuml format",
			format: FormatPlantUML,
			want:   true,
		},
		{
			name:   "invalid format",
			format: "invalid",
			want:   false,
		},
		{
			name:   "empty format",
			format: "",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFormat(tt.format)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsValidGrouping(t *testing.T) {
	tests := []struct {
		name     string
		grouping GroupingStrategy
		want     bool
	}{
		{
			name:     "valid module grouping",
			grouping: GroupByModule,
			want:     true,
		},
		{
			name:     "valid action grouping",
			grouping: GroupByAction,
			want:     true,
		},
		{
			name:     "valid resource type grouping",
			grouping: GroupByResourceType,
			want:     true,
		},
		{
			name:     "invalid grouping",
			grouping: "invalid",
			want:     false,
		},
		{
			name:     "empty grouping",
			grouping: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGrouping(tt.grouping)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRunPlanGraph(t *testing.T) {
	// Create a temporary plan file for testing
	planData := `{
		"format_version": "1.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"module_address": "",
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before_sensitive": {},
					"after_sensitive": {}
				}
			}
		]
	}`

	tmpfile, err := os.CreateTemp("", "test-plan-*.json")
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			t.Logf("Failed to remove temporary file: %v", removeErr)
		}
	}()

	_, err = tmpfile.WriteString(planData)
	require.NoError(t, err)
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	tests := []struct {
		name    string
		opts    Options
		wantErr bool
	}{
		{
			name: "valid options",
			opts: Options{
				Format:  FormatGraphviz,
				GroupBy: GroupByModule,
			},
			wantErr: false,
		},
		{
			name: "invalid format",
			opts: Options{
				Format:  "invalid",
				GroupBy: GroupByModule,
			},
			wantErr: true,
		},
		{
			name: "invalid grouping",
			opts: Options{
				Format:  FormatGraphviz,
				GroupBy: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := runPlanGraph(tmpfile.Name(), tt.opts)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunPlanGraphWithOutputFile(t *testing.T) {
	// Create a temporary plan file for testing
	planData := `{
		"format_version": "1.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"module_address": "",
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before_sensitive": {},
					"after_sensitive": {}
				}
			}
		]
	}`

	tmpfile, err := os.CreateTemp("", "test-plan-*.json")
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			t.Logf("Failed to remove temporary file: %v", removeErr)
		}
	}()

	_, err = tmpfile.WriteString(planData)
	require.NoError(t, err)
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	// Get a unique output file name, but do not pre-create the file
	outputFile, err := os.CreateTemp("", "test-output-*.dot")
	require.NoError(t, err)
	outputFileName := outputFile.Name()
	if err := outputFile.Close(); err != nil {
		t.Fatalf("Failed to close output file: %v", err)
	}
	if removeErr := os.Remove(outputFileName); removeErr != nil {
		t.Logf("Failed to remove pre-created output file: %v", removeErr)
	}
	defer func() {
		if removeErr := os.Remove(outputFileName); removeErr != nil {
			t.Logf("Failed to remove output file: %v", removeErr)
		}
	}()

	opts := Options{
		Format:  FormatGraphviz,
		Output:  outputFileName,
		GroupBy: GroupByModule,
	}

	err = runPlanGraph(tmpfile.Name(), opts)
	require.NoError(t, err)

	// Check that the output file was created and contains content
	content, err := os.ReadFile(outputFileName)
	require.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Contains(t, string(content), "digraph terraform_plan")
}

func TestRunPlanGraphWithNonexistentFile(t *testing.T) {
	opts := Options{
		Format:  FormatGraphviz,
		GroupBy: GroupByModule,
	}

	err := runPlanGraph("nonexistent-file.json", opts)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse plan file")
}

func TestCommandExecution(t *testing.T) {
	cmd := NewCommand()
	require.NotNil(t, cmd)

	// Test that the command has the expected flags
	assert.NotNil(t, cmd.Flags().Lookup("format"))
	assert.NotNil(t, cmd.Flags().Lookup("output"))
	assert.NotNil(t, cmd.Flags().Lookup("group-by"))
	assert.NotNil(t, cmd.Flags().Lookup("no-data-sources"))
	assert.NotNil(t, cmd.Flags().Lookup("no-outputs"))
	assert.NotNil(t, cmd.Flags().Lookup("no-variables"))
	assert.NotNil(t, cmd.Flags().Lookup("no-locals"))
	assert.NotNil(t, cmd.Flags().Lookup("compact"))
	assert.NotNil(t, cmd.Flags().Lookup("verbose"))
}

func TestRunPlanGraphWithOutputs(t *testing.T) {
	// Create a temporary plan file for testing with outputs
	planData := `{
		"format_version": "1.0",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"module_address": "",
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before_sensitive": {},
					"after_sensitive": {}
				}
			}
		],
		"output_changes": {
			"web_ip": {
				"change": {
					"actions": ["create"],
					"before": null,
					"after": "192.168.1.100",
					"before_sensitive": false,
					"after_sensitive": false
				}
			},
			"app_secret": {
				"change": {
					"actions": ["create"],
					"before": null,
					"after": "sensitive-value",
					"before_sensitive": false,
					"after_sensitive": true
				}
			}
		},
		"configuration": {
			"root_module": {
				"outputs": {
					"web_ip": {
						"expression": {
							"references": ["aws_instance.web.public_ip"]
						},
						"sensitive": false
					},
					"app_secret": {
						"expression": {
							"references": ["random_password.secret.result"]
						},
						"sensitive": true
					}
				}
			}
		}
	}`

	tmpfile, err := os.CreateTemp("", "test-plan-outputs-*.json")
	require.NoError(t, err)
	defer func() {
		if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
			t.Logf("Failed to remove temporary file: %v", removeErr)
		}
	}()

	_, err = tmpfile.WriteString(planData)
	require.NoError(t, err)
	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temporary file: %v", err)
	}

	tests := []struct {
		name         string
		noOutputs    bool
		expectOutput bool
	}{
		{
			name:         "without no-outputs flag (default: show outputs)",
			noOutputs:    false,
			expectOutput: true,
		},
		{
			name:         "with no-outputs flag (exclude outputs)",
			noOutputs:    true,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				Format:    FormatMermaid,
				GroupBy:   GroupByModule,
				NoOutputs: tt.noOutputs,
			}

			// Capture output to string by running against a test plan
			plan, err := ParsePlanFile(tmpfile.Name())
			require.NoError(t, err)

			graphData, err := BuildGraphData(plan, opts)
			require.NoError(t, err)

			// Check if output nodes are present
			outputNodes := 0
			for _, node := range graphData.Nodes {
				if node.Type == "output" {
					outputNodes++
				}
			}

			if tt.expectOutput {
				assert.Greater(t, outputNodes, 0, "Expected output nodes when no-outputs is disabled")
				assert.Equal(t, 2, outputNodes, "Expected exactly 2 output nodes")
			} else {
				assert.Equal(t, 0, outputNodes, "Expected no output nodes when no-outputs is enabled")
			}
		})
	}
}
