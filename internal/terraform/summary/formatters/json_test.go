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

package formatters

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yu/terraform-ops/internal/core"
)

func TestNewJSONFormatter(t *testing.T) {
	formatter := NewJSONFormatter()
	assert.NotNil(t, formatter)
}

func TestJSONFormatter_Format_EmptyPlan(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges:      0,
			ActionBreakdown:   make(map[string]int),
			ProviderBreakdown: make(map[string]int),
			ResourceBreakdown: make(map[string]int),
			ModuleBreakdown:   make(map[string]int),
		},
		Changes: core.Changes{},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{}
	formatter := NewJSONFormatter()

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Verify it's valid JSON
	var parsedSummary core.PlanSummary
	err = json.Unmarshal([]byte(output), &parsedSummary)
	require.NoError(t, err)

	// Check that the parsed data matches the original
	assert.Equal(t, "1.0", parsedSummary.PlanInfo.FormatVersion)
	assert.True(t, parsedSummary.PlanInfo.Applicable)
	assert.True(t, parsedSummary.PlanInfo.Complete)
	assert.False(t, parsedSummary.PlanInfo.Errored)
	assert.Equal(t, 0, parsedSummary.Statistics.TotalChanges)
}

func TestJSONFormatter_Format_WithResourceChanges(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges: 2,
			ActionBreakdown: map[string]int{
				"create": 1,
				"update": 1,
			},
			ProviderBreakdown: map[string]int{
				"aws": 2,
			},
			ResourceBreakdown: map[string]int{
				"aws_instance":       1,
				"aws_security_group": 1,
			},
			ModuleBreakdown: map[string]int{
				"root": 2,
			},
		},
		Changes: core.Changes{
			Create: []core.ResourceSummary{
				{
					Address:       "aws_instance.web",
					ModuleAddress: "",
					Type:          "aws_instance",
					Name:          "web",
					Provider:      "aws",
					Actions:       []string{"create"},
					Sensitive:     false,
					KeyChanges: map[string]interface{}{
						"instance_type": map[string]interface{}{
							"from": nil,
							"to":   "t3.micro",
						},
					},
				},
			},
			Update: []core.ResourceSummary{
				{
					Address:       "aws_security_group.web",
					ModuleAddress: "",
					Type:          "aws_security_group",
					Name:          "web",
					Provider:      "aws",
					Actions:       []string{"update"},
					Sensitive:     true,
				},
			},
		},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{}
	formatter := NewJSONFormatter()

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Verify it's valid JSON
	var parsedSummary core.PlanSummary
	err = json.Unmarshal([]byte(output), &parsedSummary)
	require.NoError(t, err)

	// Check that the parsed data matches the original
	assert.Equal(t, 2, parsedSummary.Statistics.TotalChanges)
	assert.Equal(t, 1, parsedSummary.Statistics.ActionBreakdown["create"])
	assert.Equal(t, 1, parsedSummary.Statistics.ActionBreakdown["update"])
	assert.Equal(t, 2, parsedSummary.Statistics.ProviderBreakdown["aws"])

	// Check resource changes
	assert.Len(t, parsedSummary.Changes.Create, 1)
	assert.Len(t, parsedSummary.Changes.Update, 1)

	createResource := parsedSummary.Changes.Create[0]
	assert.Equal(t, "aws_instance.web", createResource.Address)
	assert.Equal(t, "aws_instance", createResource.Type)
	assert.Equal(t, "aws", createResource.Provider)
	assert.False(t, createResource.Sensitive)
	assert.Len(t, createResource.KeyChanges, 1)

	updateResource := parsedSummary.Changes.Update[0]
	assert.Equal(t, "aws_security_group.web", updateResource.Address)
	assert.Equal(t, "aws_security_group", updateResource.Type)
	assert.Equal(t, "aws", updateResource.Provider)
	assert.True(t, updateResource.Sensitive)
}

func TestJSONFormatter_Format_WithOutputChanges(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges:      0,
			ActionBreakdown:   make(map[string]int),
			ProviderBreakdown: make(map[string]int),
			ResourceBreakdown: make(map[string]int),
			ModuleBreakdown:   make(map[string]int),
		},
		Changes: core.Changes{},
		Outputs: []core.OutputSummary{
			{
				Name:      "instance_id",
				Actions:   []string{"create"},
				Sensitive: false,
				Value:     "i-1234567890abcdef0",
			},
			{
				Name:      "secret_value",
				Actions:   []string{"update"},
				Sensitive: true,
				Value:     nil,
			},
		},
	}

	opts := core.SummaryOptions{}
	formatter := NewJSONFormatter()

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Verify it's valid JSON
	var parsedSummary core.PlanSummary
	err = json.Unmarshal([]byte(output), &parsedSummary)
	require.NoError(t, err)

	// Check outputs
	assert.Len(t, parsedSummary.Outputs, 2)

	instanceID := parsedSummary.Outputs[0]
	assert.Equal(t, "instance_id", instanceID.Name)
	assert.Equal(t, []string{"create"}, instanceID.Actions)
	assert.False(t, instanceID.Sensitive)
	assert.Equal(t, "i-1234567890abcdef0", instanceID.Value)

	secretValue := parsedSummary.Outputs[1]
	assert.Equal(t, "secret_value", secretValue.Name)
	assert.Equal(t, []string{"update"}, secretValue.Actions)
	assert.True(t, secretValue.Sensitive)
	assert.Nil(t, secretValue.Value)
}

func TestJSONFormatter_Format_Indentation(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges:      0,
			ActionBreakdown:   make(map[string]int),
			ProviderBreakdown: make(map[string]int),
			ResourceBreakdown: make(map[string]int),
			ModuleBreakdown:   make(map[string]int),
		},
		Changes: core.Changes{},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{}
	formatter := NewJSONFormatter()

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that the output is properly indented (contains spaces)
	assert.Contains(t, output, "  ")
	assert.Contains(t, output, "\n")
}
