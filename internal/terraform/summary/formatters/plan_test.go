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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yu/terraform-ops/internal/core"
)

func TestNewPlanFormatter(t *testing.T) {
	formatter := NewPlanFormatter(true)
	assert.NotNil(t, formatter)
	assert.True(t, formatter.useColor)

	formatter = NewPlanFormatter(false)
	assert.NotNil(t, formatter)
	assert.False(t, formatter.useColor)
}

func TestPlanFormatter_Format_EmptyPlan(t *testing.T) {
	formatter := NewPlanFormatter(false)

	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.2",
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
	output, err := formatter.Format(summary, opts)

	assert.NoError(t, err)
	assert.Contains(t, output, "Terraform used the selected providers to generate the following execution plan")
	assert.Contains(t, output, "No changes. Your infrastructure matches the configuration.")
	assert.Contains(t, output, "No changes. No objects need to be destroyed.")
}

func TestPlanFormatter_Format_WithResourceChanges(t *testing.T) {
	formatter := NewPlanFormatter(false)

	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.2",
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
		},
		Changes: core.Changes{
			Create: []core.ResourceSummary{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Name:    "web",
					Actions: []string{"create"},
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
					Address: "aws_security_group.web",
					Type:    "aws_security_group",
					Name:    "web",
					Actions: []string{"update"},
					KeyChanges: map[string]interface{}{
						"name": map[string]interface{}{
							"from": "old-sg",
							"to":   "new-sg",
						},
					},
				},
			},
			Delete: []core.ResourceSummary{
				{
					Address:       "module.database.aws_instance.db",
					ModuleAddress: "module.database",
					Type:          "aws_instance",
					Name:          "db",
					Actions:       []string{"delete"},
					KeyChanges: map[string]interface{}{
						"instance_type": map[string]interface{}{
							"from": "t2.small",
							"to":   nil,
						},
					},
				},
			},
		},
	}

	opts := core.SummaryOptions{ShowDetails: true}
	output, err := formatter.Format(summary, opts)

	assert.NoError(t, err)
	assert.Contains(t, output, "Terraform will perform the following actions:")

	// Check for create action
	assert.Contains(t, output, "# aws_instance.web will be created")
	assert.Contains(t, output, "+ resource \"aws_instance\" \"web\" {")
	assert.Contains(t, output, "+ instance_type = \"t3.micro\"")

	// Check for update action
	assert.Contains(t, output, "# aws_security_group.web will be updated in-place")
	assert.Contains(t, output, "~ resource \"aws_security_group\" \"web\" {")
	assert.Contains(t, output, "~ name = \"old-sg\"")
	assert.Contains(t, output, "        -> \"new-sg\"")

	// Check for delete action
	assert.Contains(t, output, "# module.database.aws_instance.db will be destroyed")
	assert.Contains(t, output, "- resource \"aws_instance\" \"db\" {")
	assert.Contains(t, output, "- instance_type = \"t2.small\"")

	// Check plan summary
	assert.Contains(t, output, "Plan: 1 to add, 1 to change, 1 to destroy.")
}

func TestPlanFormatter_Format_WithReplaceAction(t *testing.T) {
	formatter := NewPlanFormatter(false)

	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.2",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges: 1,
			ActionBreakdown: map[string]int{
				"replace": 1,
			},
		},
		Changes: core.Changes{
			Replace: []core.ResourceSummary{
				{
					Address: "aws_instance.web",
					Type:    "aws_instance",
					Name:    "web",
					Actions: []string{"delete", "create"},
					KeyChanges: map[string]interface{}{
						"instance_type": map[string]interface{}{
							"from": "t2.micro",
							"to":   "t3.micro",
						},
					},
				},
			},
		},
	}

	opts := core.SummaryOptions{ShowDetails: true}
	output, err := formatter.Format(summary, opts)

	assert.NoError(t, err)
	assert.Contains(t, output, "# aws_instance.web must be replaced")
	assert.Contains(t, output, "-/+ resource \"aws_instance\" \"web\" {")
	assert.Contains(t, output, "~ instance_type = \"t2.micro\"")
	assert.Contains(t, output, "        -> \"t3.micro\"")
}

func TestPlanFormatter_Format_WithSensitiveResources(t *testing.T) {
	formatter := NewPlanFormatter(false)

	summary := &core.PlanSummary{
		PlanInfo: core.PlanInfo{
			FormatVersion: "1.2",
			Applicable:    true,
			Complete:      true,
			Errored:       false,
		},
		Statistics: core.Statistics{
			TotalChanges: 1,
			ActionBreakdown: map[string]int{
				"create": 1,
			},
		},
		Changes: core.Changes{
			Create: []core.ResourceSummary{
				{
					Address:   "aws_db_instance.main",
					Type:      "aws_db_instance",
					Name:      "main",
					Actions:   []string{"create"},
					Sensitive: true,
				},
			},
		},
	}

	opts := core.SummaryOptions{}
	output, err := formatter.Format(summary, opts)

	assert.NoError(t, err)
	assert.Contains(t, output, "# aws_db_instance.main will be created")
	assert.Contains(t, output, "+ resource \"aws_db_instance\" \"main\" {")
	assert.Contains(t, output, "# (sensitive value)")
}

func TestPlanFormatter_GetActionSymbolAndColor(t *testing.T) {
	formatter := NewPlanFormatter(false)

	tests := []struct {
		actions        []string
		expectedSymbol string
		expectedColor  string
	}{
		{[]string{"create"}, "  +", "green"},
		{[]string{"update"}, "  ~", "yellow"},
		{[]string{"delete"}, "  -", "red"},
		{[]string{"delete", "create"}, "-/+", "yellow"},
		{[]string{"no-op"}, "   ", ""},
		{[]string{}, "  ", ""},
	}

	for _, test := range tests {
		t.Run(strings.Join(test.actions, ","), func(t *testing.T) {
			symbol, color := formatter.getActionSymbolAndColor(test.actions)
			assert.Equal(t, test.expectedSymbol, symbol)
			assert.Equal(t, test.expectedColor, color)
		})
	}
}

func TestPlanFormatter_FormatValue(t *testing.T) {
	formatter := NewPlanFormatter(false)

	tests := []struct {
		value    interface{}
		expected string
	}{
		{nil, "(null)"},
		{"hello", `"hello"`},
		{true, "true"},
		{false, "false"},
		{123, "123"},
		{map[string]interface{}{"key": "value"}, "{...}"},
		{[]interface{}{"item1", "item2"}, "[...]"},
	}

	for _, test := range tests {
		t.Run(test.expected, func(t *testing.T) {
			result := formatter.formatValue(test.value)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestPlanFormatter_Colorize(t *testing.T) {
	// Test with color enabled
	formatter := NewPlanFormatter(true)

	greenText := formatter.colorize("test", "green")
	assert.Contains(t, greenText, "\033[32m")
	assert.Contains(t, greenText, "\033[0m")
	assert.Contains(t, greenText, "test")

	yellowText := formatter.colorize("test", "yellow")
	assert.Contains(t, yellowText, "\033[33m")

	redText := formatter.colorize("test", "red")
	assert.Contains(t, redText, "\033[31m")

	// Test with color disabled
	formatter = NewPlanFormatter(false)

	plainText := formatter.colorize("test", "green")
	assert.Equal(t, "test", plainText)
	assert.NotContains(t, plainText, "\033[")
}

func TestPlanFormatter_ContainsAction(t *testing.T) {
	formatter := NewPlanFormatter(false)

	actions := []string{"delete", "create"}

	assert.True(t, formatter.containsAction(actions, "delete"))
	assert.True(t, formatter.containsAction(actions, "create"))
	assert.False(t, formatter.containsAction(actions, "update"))
}
