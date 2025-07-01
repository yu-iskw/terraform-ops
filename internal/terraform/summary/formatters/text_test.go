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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yu/terraform-ops/internal/core"
)

func TestNewTextFormatter(t *testing.T) {
	formatter := NewTextFormatter(true)
	assert.NotNil(t, formatter)
	assert.True(t, formatter.useColor)

	formatter = NewTextFormatter(false)
	assert.NotNil(t, formatter)
	assert.False(t, formatter.useColor)
}

func TestTextFormatter_Format_EmptyPlan(t *testing.T) {
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
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that header is present
	assert.Contains(t, output, "Terraform Plan Summary")
	assert.Contains(t, output, "Plan Status: ✅ Applicable")
	assert.Contains(t, output, "Format Version: 1.0")
	assert.Contains(t, output, "Complete: true")

	// Check that statistics section is present
	assert.Contains(t, output, "📊 Statistics")
	assert.Contains(t, output, "Total Changes: 0")

	// Check that resource changes section is present
	assert.Contains(t, output, "🔄 Resource Changes")
}

func TestTextFormatter_Format_WithResourceChanges(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges: 3,
			ActionBreakdown: map[string]int{
				"create": 1,
				"update": 1,
				"delete": 1,
			},
			ProviderBreakdown: map[string]int{
				"aws": 3,
			},
			ResourceBreakdown: map[string]int{
				"aws_instance":       2,
				"aws_security_group": 1,
			},
			ModuleBreakdown: map[string]int{
				"root":            2,
				"module.database": 1,
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
					Sensitive:     false,
				},
			},
			Delete: []core.ResourceSummary{
				{
					Address:       "aws_instance.db",
					ModuleAddress: "module.database",
					Type:          "aws_instance",
					Name:          "db",
					Provider:      "aws",
					Actions:       []string{"delete"},
					Sensitive:     false,
				},
			},
		},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{}
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that all action groups are present
	assert.Contains(t, output, "➕ Create (1)")
	assert.Contains(t, output, "🔄 Update (1)")
	assert.Contains(t, output, "❌ Delete (1)")

	// Check that resources are listed
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "aws_security_group.web")
	assert.Contains(t, output, "aws_instance.db")

	// Check statistics
	assert.Contains(t, output, "Total Changes: 3")
	assert.Contains(t, output, "➕ create: 1")
	assert.Contains(t, output, "🔄 update: 1")
	assert.Contains(t, output, "❌ delete: 1")
	assert.Contains(t, output, "🏢 aws: 3")
}

func TestTextFormatter_Format_WithSensitiveResources(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges:      1,
			ActionBreakdown:   make(map[string]int),
			ProviderBreakdown: make(map[string]int),
			ResourceBreakdown: make(map[string]int),
			ModuleBreakdown:   make(map[string]int),
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
					Sensitive:     true,
				},
			},
		},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{}
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that sensitive indicator is present
	assert.Contains(t, output, "🔒 Contains sensitive values")
}

func TestTextFormatter_Format_WithOutputChanges(t *testing.T) {
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
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that output changes section is present
	assert.Contains(t, output, "📤 Output Changes")
	assert.Contains(t, output, "instance_id")
	assert.Contains(t, output, "secret_value")
	assert.Contains(t, output, "🔒 Sensitive value")
	assert.Contains(t, output, "Value: i-1234567890abcdef0")
}

func TestTextFormatter_Format_WithDetails(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges:      1,
			ActionBreakdown:   make(map[string]int),
			ProviderBreakdown: make(map[string]int),
			ResourceBreakdown: make(map[string]int),
			ModuleBreakdown:   make(map[string]int),
		},
		Changes: core.Changes{
			Update: []core.ResourceSummary{
				{
					Address:       "aws_instance.web",
					ModuleAddress: "",
					Type:          "aws_instance",
					Name:          "web",
					Provider:      "aws",
					Actions:       []string{"update"},
					Sensitive:     false,
					KeyChanges: map[string]interface{}{
						"instance_type": map[string]interface{}{
							"from": "t2.micro",
							"to":   "t3.micro",
						},
					},
				},
			},
		},
		Outputs: []core.OutputSummary{},
	}

	opts := core.SummaryOptions{
		ShowDetails: true,
	}
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that detailed changes are shown
	assert.Contains(t, output, "Changes:")
	assert.Contains(t, output, "instance_type: t2.micro → t3.micro")
}

func TestTextFormatter_GetActionIcon(t *testing.T) {
	formatter := NewTextFormatter(false)

	tests := []struct {
		action   string
		expected string
	}{
		{"create", "➕"},
		{"update", "🔄"},
		{"delete", "❌"},
		{"replace", "🔄"},
		{"no-op", "➖"},
		{"unknown", "❓"},
	}

	for _, test := range tests {
		t.Run(test.action, func(t *testing.T) {
			result := formatter.getActionIcon(test.action)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTextFormatter_Format_ErroredPlan(t *testing.T) {
	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.0",
			Applicable:    false,
			Complete:      false,
			Errored:       true,
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
	formatter := NewTextFormatter(false)

	output, err := formatter.Format(summary, opts)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that error status is shown
	assert.Contains(t, output, "💥 Errored")
}
