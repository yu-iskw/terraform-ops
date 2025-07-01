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

package summary

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yu/terraform-ops/internal/core"
)

func TestNewSummarizer(t *testing.T) {
	summarizer := NewSummarizer()
	assert.NotNil(t, summarizer)
}

func TestSummarizePlan_EmptyPlan(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Applicable:      true,
		Complete:        true,
		Errored:         false,
	}

	opts := core.SummaryOptions{}
	summarizer := NewSummarizer()

	summary, err := summarizer.SummarizePlan(plan, opts)
	require.NoError(t, err)
	assert.NotNil(t, summary)

	// Check plan info
	assert.Equal(t, "1.0", summary.PlanInfo.FormatVersion)
	assert.True(t, summary.PlanInfo.Applicable)
	assert.True(t, summary.PlanInfo.Complete)
	assert.False(t, summary.PlanInfo.Errored)

	// Check statistics
	assert.Equal(t, 0, summary.Statistics.TotalChanges)
	assert.Empty(t, summary.Statistics.ActionBreakdown)
	assert.Empty(t, summary.Statistics.ProviderBreakdown)
	assert.Empty(t, summary.Statistics.ResourceBreakdown)
	assert.Empty(t, summary.Statistics.ModuleBreakdown)

	// Check changes
	assert.Empty(t, summary.Changes.Create)
	assert.Empty(t, summary.Changes.Update)
	assert.Empty(t, summary.Changes.Delete)
	assert.Empty(t, summary.Changes.Replace)
	assert.Empty(t, summary.Changes.NoOp)

	// Check outputs
	assert.Empty(t, summary.Outputs)
}

func TestSummarizePlan_WithResourceChanges(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "aws_instance.web",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_instance",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After: map[string]interface{}{
						"instance_type": "t3.micro",
						"ami":           "ami-12345678",
					},
				},
			},
			{
				Address:       "aws_security_group.web",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_security_group",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"update"},
					Before: map[string]interface{}{
						"name": "old-sg",
					},
					After: map[string]interface{}{
						"name": "new-sg",
					},
				},
			},
			{
				Address:       "module.database.aws_instance.db",
				ModuleAddress: "module.database",
				Mode:          "managed",
				Type:          "aws_instance",
				Name:          "db",
				Change: core.Change{
					Actions: []string{"delete"},
					Before: map[string]interface{}{
						"instance_type": "t2.small",
					},
					After: nil,
				},
			},
		},
		OutputChanges: make(map[string]core.OutputChange),
		Variables:     make(map[string]core.Variable),
		Applicable:    true,
		Complete:      true,
		Errored:       false,
	}

	opts := core.SummaryOptions{}
	summarizer := NewSummarizer()

	summary, err := summarizer.SummarizePlan(plan, opts)
	require.NoError(t, err)
	assert.NotNil(t, summary)

	// Check statistics
	assert.Equal(t, 3, summary.Statistics.TotalChanges)
	assert.Equal(t, 1, summary.Statistics.ActionBreakdown["create"])
	assert.Equal(t, 1, summary.Statistics.ActionBreakdown["update"])
	assert.Equal(t, 1, summary.Statistics.ActionBreakdown["delete"])
	assert.Equal(t, 3, summary.Statistics.ProviderBreakdown["aws"])
	assert.Equal(t, 2, summary.Statistics.ResourceBreakdown["aws_instance"])
	assert.Equal(t, 1, summary.Statistics.ResourceBreakdown["aws_security_group"])
	assert.Equal(t, 2, summary.Statistics.ModuleBreakdown["root"])
	assert.Equal(t, 1, summary.Statistics.ModuleBreakdown["module.database"])

	// Check changes
	assert.Len(t, summary.Changes.Create, 1)
	assert.Len(t, summary.Changes.Update, 1)
	assert.Len(t, summary.Changes.Delete, 1)
	assert.Empty(t, summary.Changes.Replace)
	assert.Empty(t, summary.Changes.NoOp)

	// Check create resource
	createResource := summary.Changes.Create[0]
	assert.Equal(t, "aws_instance.web", createResource.Address)
	assert.Equal(t, "aws_instance", createResource.Type)
	assert.Equal(t, "web", createResource.Name)
	assert.Equal(t, "aws", createResource.Provider)
	assert.Equal(t, []string{"create"}, createResource.Actions)
	assert.False(t, createResource.Sensitive)

	// Check update resource
	updateResource := summary.Changes.Update[0]
	assert.Equal(t, "aws_security_group.web", updateResource.Address)
	assert.Equal(t, "aws_security_group", updateResource.Type)
	assert.Equal(t, "web", updateResource.Name)
	assert.Equal(t, "aws", updateResource.Provider)
	assert.Equal(t, []string{"update"}, updateResource.Actions)
	assert.False(t, updateResource.Sensitive)

	// Check delete resource
	deleteResource := summary.Changes.Delete[0]
	assert.Equal(t, "module.database.aws_instance.db", deleteResource.Address)
	assert.Equal(t, "module.database", deleteResource.ModuleAddress)
	assert.Equal(t, "aws_instance", deleteResource.Type)
	assert.Equal(t, "db", deleteResource.Name)
	assert.Equal(t, "aws", deleteResource.Provider)
	assert.Equal(t, []string{"delete"}, deleteResource.Actions)
	assert.False(t, deleteResource.Sensitive)
}

func TestSummarizePlan_WithOutputChanges(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges: map[string]core.OutputChange{
			"instance_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "i-1234567890abcdef0",
				},
			},
			"secret_value": {
				Change: core.Change{
					Actions: []string{"update"},
					Before:  "old-secret",
					After:   "new-secret",
					AfterSensitive: map[string]interface{}{
						"value": true,
					},
				},
			},
		},
		Variables:  make(map[string]core.Variable),
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.SummaryOptions{}
	summarizer := NewSummarizer()

	summary, err := summarizer.SummarizePlan(plan, opts)
	require.NoError(t, err)
	assert.NotNil(t, summary)

	// Check outputs
	assert.Len(t, summary.Outputs, 2)

	// Find outputs by name since order is not guaranteed
	var instanceID, secretValue *core.OutputSummary
	for i := range summary.Outputs {
		switch summary.Outputs[i].Name {
		case "instance_id":
			instanceID = &summary.Outputs[i]
		case "secret_value":
			secretValue = &summary.Outputs[i]
		}
	}

	// Check non-sensitive output
	assert.NotNil(t, instanceID)
	assert.Equal(t, "instance_id", instanceID.Name)
	assert.Equal(t, []string{"create"}, instanceID.Actions)
	assert.False(t, instanceID.Sensitive)
	assert.Equal(t, "i-1234567890abcdef0", instanceID.Value)

	// Check sensitive output
	assert.NotNil(t, secretValue)
	assert.Equal(t, "secret_value", secretValue.Name)
	assert.Equal(t, []string{"update"}, secretValue.Actions)
	assert.True(t, secretValue.Sensitive)
	assert.Nil(t, secretValue.Value) // Should be nil for sensitive values
}

func TestSummarizePlan_WithReplaceAction(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "aws_instance.web",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_instance",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"delete", "create"},
					Before: map[string]interface{}{
						"instance_type": "t2.micro",
					},
					After: map[string]interface{}{
						"instance_type": "t3.micro",
					},
				},
			},
		},
		OutputChanges: make(map[string]core.OutputChange),
		Variables:     make(map[string]core.Variable),
		Applicable:    true,
		Complete:      true,
		Errored:       false,
	}

	opts := core.SummaryOptions{}
	summarizer := NewSummarizer()

	summary, err := summarizer.SummarizePlan(plan, opts)
	require.NoError(t, err)
	assert.NotNil(t, summary)

	// Check that replace action is correctly identified
	assert.Len(t, summary.Changes.Replace, 1)
	assert.Empty(t, summary.Changes.Create)
	assert.Empty(t, summary.Changes.Update)
	assert.Empty(t, summary.Changes.Delete)

	replaceResource := summary.Changes.Replace[0]
	assert.Equal(t, "aws_instance.web", replaceResource.Address)
	assert.Equal(t, []string{"delete", "create"}, replaceResource.Actions)
}

func TestExtractProvider(t *testing.T) {
	summarizer := NewSummarizer()

	tests := []struct {
		address  string
		expected string
	}{
		{"aws_instance.web", "aws"},
		{"google_compute_instance.web", "google"},
		{"azurerm_virtual_machine.web", "azurerm"},
		{"module.database.aws_instance.db", "aws"},
		{"module.network.module.subnet.aws_subnet.private", "aws"},
		{"random_string.password", "random"},
		{"unknown", "unknown"},
	}

	for _, test := range tests {
		t.Run(test.address, func(t *testing.T) {
			result := summarizer.extractProvider(test.address)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestHasSensitiveValues(t *testing.T) {
	summarizer := NewSummarizer()

	tests := []struct {
		name     string
		change   core.Change
		expected bool
	}{
		{
			name: "no sensitive values",
			change: core.Change{
				AfterSensitive: map[string]interface{}{
					"name": false,
				},
			},
			expected: false,
		},
		{
			name: "has sensitive values",
			change: core.Change{
				AfterSensitive: map[string]interface{}{
					"password": true,
					"name":     false,
				},
			},
			expected: true,
		},
		{
			name: "before sensitive values",
			change: core.Change{
				BeforeSensitive: map[string]interface{}{
					"secret": true,
				},
			},
			expected: true,
		},
		{
			name:     "nil sensitive values",
			change:   core.Change{},
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := summarizer.hasSensitiveValues(test.change)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestGetPrimaryAction(t *testing.T) {
	summarizer := NewSummarizer()

	tests := []struct {
		actions  []string
		expected string
	}{
		{[]string{"create"}, "create"},
		{[]string{"update"}, "update"},
		{[]string{"delete"}, "delete"},
		{[]string{"delete", "create"}, "replace"},
		{[]string{"create", "delete"}, "replace"},
		{[]string{}, "no-op"},
		{[]string{"no-op"}, "no-op"},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.actions), func(t *testing.T) {
			result := summarizer.getPrimaryAction(test.actions)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestExtractKeyChanges(t *testing.T) {
	summarizer := NewSummarizer()

	// Test create scenario
	createChange := core.ResourceChange{
		Change: core.Change{
			Before: nil,
			After: map[string]interface{}{
				"instance_type": "t3.micro",
				"ami":           "ami-12345678",
			},
		},
	}

	keyChanges := summarizer.extractKeyChanges(createChange)
	assert.Len(t, keyChanges, 2)
	assert.Equal(t, map[string]interface{}{"from": nil, "to": "t3.micro"}, keyChanges["instance_type"])
	assert.Equal(t, map[string]interface{}{"from": nil, "to": "ami-12345678"}, keyChanges["ami"])

	// Test update scenario
	updateChange := core.ResourceChange{
		Change: core.Change{
			Before: map[string]interface{}{
				"instance_type": "t2.micro",
				"ami":           "ami-12345678",
			},
			After: map[string]interface{}{
				"instance_type": "t3.micro",
				"ami":           "ami-12345678",
			},
		},
	}

	keyChanges = summarizer.extractKeyChanges(updateChange)
	assert.Len(t, keyChanges, 1)
	assert.Equal(t, map[string]interface{}{"from": "t2.micro", "to": "t3.micro"}, keyChanges["instance_type"])

	// Test delete scenario
	deleteChange := core.ResourceChange{
		Change: core.Change{
			Before: map[string]interface{}{
				"instance_type": "t3.micro",
				"ami":           "ami-12345678",
			},
			After: nil,
		},
	}

	keyChanges = summarizer.extractKeyChanges(deleteChange)
	assert.Len(t, keyChanges, 2)
	assert.Equal(t, map[string]interface{}{"from": "t3.micro", "to": nil}, keyChanges["instance_type"])
	assert.Equal(t, map[string]interface{}{"from": "ami-12345678", "to": nil}, keyChanges["ami"])
}
