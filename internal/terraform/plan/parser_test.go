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

package plan

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yu/terraform-ops/internal/core"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
}

func TestParsePlanFile_ValidPlan(t *testing.T) {
	// Create a temporary file with valid plan JSON
	validPlanJSON := `{
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
					"before": null,
					"after": {"instance_type": "t3.micro"}
				}
			}
		],
		"output_changes": {},
		"configuration": {
			"provider_config": {},
			"root_module": {
				"resources": [],
				"module_calls": {},
				"outputs": {},
				"variables": {},
				"locals": {}
			}
		},
		"variables": {},
		"applicable": true,
		"complete": true,
		"errored": false
	}`

	tmpFile := createTempFile(t, validPlanJSON)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	plan, err := parser.ParsePlanFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "1.0", plan.FormatVersion)
	assert.Len(t, plan.ResourceChanges, 1)
	assert.Equal(t, "aws_instance.web", plan.ResourceChanges[0].Address)
	assert.Equal(t, "aws_instance", plan.ResourceChanges[0].Type)
	assert.Equal(t, "web", plan.ResourceChanges[0].Name)
	assert.Equal(t, []string{"create"}, plan.ResourceChanges[0].Change.Actions)
	assert.True(t, plan.Applicable)
	assert.True(t, plan.Complete)
	assert.False(t, plan.Errored)
}

func TestParsePlanFile_InvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	invalidJSON := `{ invalid json }`

	tmpFile := createTempFile(t, invalidJSON)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	plan, err := parser.ParsePlanFile(tmpFile.Name())

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "failed to parse JSON")
}

func TestParsePlanFile_FileNotFound(t *testing.T) {
	parser := NewParser()
	plan, err := parser.ParsePlanFile("nonexistent_file.json")

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestParsePlanFile_InvalidFormatVersion(t *testing.T) {
	// Create a temporary file with invalid format version
	invalidVersionJSON := `{
		"format_version": "2.0",
		"resource_changes": [],
		"output_changes": {},
		"configuration": {
			"provider_config": {},
			"root_module": {
				"resources": [],
				"module_calls": {},
				"outputs": {},
				"variables": {},
				"locals": {}
			}
		},
		"variables": {},
		"applicable": true,
		"complete": true,
		"errored": false
	}`

	tmpFile := createTempFile(t, invalidVersionJSON)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	plan, err := parser.ParsePlanFile(tmpFile.Name())

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "unsupported format version")
}

func TestParsePlanFile_MissingFormatVersion(t *testing.T) {
	// Create a temporary file with missing format version
	missingVersionJSON := `{
		"resource_changes": [],
		"output_changes": {},
		"configuration": {
			"provider_config": {},
			"root_module": {
				"resources": [],
				"module_calls": {},
				"outputs": {},
				"variables": {},
				"locals": {}
			}
		},
		"variables": {},
		"applicable": true,
		"complete": true,
		"errored": false
	}`

	tmpFile := createTempFile(t, missingVersionJSON)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	plan, err := parser.ParsePlanFile(tmpFile.Name())

	assert.Error(t, err)
	assert.Nil(t, plan)
	assert.Contains(t, err.Error(), "missing format_version")
}

func TestParsePlanFile_ComplexPlan(t *testing.T) {
	// Create a temporary file with a more complex plan
	complexPlanJSON := `{
		"format_version": "1.1",
		"resource_changes": [
			{
				"address": "aws_instance.web",
				"module_address": "",
				"mode": "managed",
				"type": "aws_instance",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"instance_type": "t3.micro"}
				}
			},
			{
				"address": "aws_security_group.web",
				"module_address": "",
				"mode": "managed",
				"type": "aws_security_group",
				"name": "web",
				"change": {
					"actions": ["create"],
					"before": null,
					"after": {"name": "web-sg"}
				}
			}
		],
		"output_changes": {
			"instance_id": {
				"change": {
					"actions": ["create"],
					"before": null,
					"after": "i-1234567890abcdef0"
				}
			}
		},
		"configuration": {
			"provider_config": {},
			"root_module": {
				"resources": [],
				"module_calls": {},
				"outputs": {},
				"variables": {},
				"locals": {}
			}
		},
		"variables": {
			"region": {"value": "us-west-2"}
		},
		"applicable": true,
		"complete": true,
		"errored": false
	}`

	tmpFile := createTempFile(t, complexPlanJSON)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	plan, err := parser.ParsePlanFile(tmpFile.Name())

	assert.NoError(t, err)
	assert.NotNil(t, plan)
	assert.Equal(t, "1.1", plan.FormatVersion)
	assert.Len(t, plan.ResourceChanges, 2)
	assert.Len(t, plan.OutputChanges, 1)
	assert.Len(t, plan.Variables, 1)

	// Check first resource
	assert.Equal(t, "aws_instance.web", plan.ResourceChanges[0].Address)
	assert.Equal(t, "aws_instance", plan.ResourceChanges[0].Type)

	// Check second resource
	assert.Equal(t, "aws_security_group.web", plan.ResourceChanges[1].Address)
	assert.Equal(t, "aws_security_group", plan.ResourceChanges[1].Type)

	// Check output
	assert.Contains(t, plan.OutputChanges, "instance_id")
	assert.Equal(t, []string{"create"}, plan.OutputChanges["instance_id"].Change.Actions)

	// Check variable
	assert.Contains(t, plan.Variables, "region")
	assert.Equal(t, "us-west-2", plan.Variables["region"].Value)
}

func TestValidatePlan_ValidPlan(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Applicable:      true,
		Complete:        true,
		Errored:         false,
	}

	parser := NewParser()
	err := parser.validatePlan(plan)

	assert.NoError(t, err)
}

func TestValidatePlan_InvalidFormatVersion(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "2.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Applicable:      true,
		Complete:        true,
		Errored:         false,
	}

	parser := NewParser()
	err := parser.validatePlan(plan)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported format version")
}

func TestValidatePlan_MissingFormatVersion(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Applicable:      true,
		Complete:        true,
		Errored:         false,
	}

	parser := NewParser()
	err := parser.validatePlan(plan)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing format_version")
}

// Helper function to create a temporary file with content
func createTempFile(t *testing.T, content string) *os.File {
	tmpFile, err := os.CreateTemp("", "test_plan_*.json")
	assert.NoError(t, err)

	_, err = tmpFile.WriteString(content)
	assert.NoError(t, err)

	err = tmpFile.Close()
	assert.NoError(t, err)

	return tmpFile
}
