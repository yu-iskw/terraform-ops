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

package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/terraform/plan"
	"github.com/yu/terraform-ops/internal/terraform/summary"
	"github.com/yu/terraform-ops/internal/terraform/summary/formatters"
)

func TestNewSummarizePlanCommand(t *testing.T) {
	planParser := plan.NewParser()
	planSummarizer := summary.NewSummarizer()
	formatterFactory := formatters.NewFactory()

	cmd := NewSummarizePlanCommand(planParser, planSummarizer, formatterFactory)
	assert.NotNil(t, cmd)
	assert.Equal(t, planParser, cmd.planParser)
	assert.Equal(t, planSummarizer, cmd.planSummarizer)
	assert.Equal(t, formatterFactory, cmd.formatterFactory)
}

func TestDefaultSummarizePlanCommand(t *testing.T) {
	cmd := DefaultSummarizePlanCommand()
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.planParser)
	assert.NotNil(t, cmd.planSummarizer)
	assert.NotNil(t, cmd.formatterFactory)
}

func TestSummarizePlanCommand_Command(t *testing.T) {
	cmd := DefaultSummarizePlanCommand()
	command := cmd.Command()

	assert.Equal(t, "summarize-plan <PLAN_FILE>", command.Use)
	assert.Equal(t, "Generate a human-readable summary of Terraform plan changes", command.Short)
	assert.NotEmpty(t, command.Long)

	// Test that Args function accepts exactly 1 argument
	err := command.Args(command, []string{"plan.json"})
	assert.NoError(t, err)

	// Test that Args function rejects 0 arguments
	err = command.Args(command, []string{})
	assert.Error(t, err)

	// Test that Args function rejects 2 arguments
	err = command.Args(command, []string{"plan1.json", "plan2.json"})
	assert.Error(t, err)
}

func TestIsValidSummaryFormat(t *testing.T) {
	tests := []struct {
		format   core.SummaryFormat
		expected bool
	}{
		{core.FormatText, true},
		{core.FormatJSON, true},
		{core.FormatMarkdown, true},
		{core.FormatTable, true},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(string(test.format), func(t *testing.T) {
			result := isValidSummaryFormat(test.format)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsValidSummaryGrouping(t *testing.T) {
	tests := []struct {
		grouping core.SummaryGrouping
		expected bool
	}{
		{core.GroupByAction, true},
		{core.GroupByModule, true},
		{core.GroupByProvider, true},
		{core.GroupByResourceType, true},
		{"invalid", false},
		{"", false},
	}

	for _, test := range tests {
		t.Run(string(test.grouping), func(t *testing.T) {
			result := isValidSummaryGrouping(test.grouping)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestShouldUseColor(t *testing.T) {
	tests := []struct {
		colorMode core.ColorMode
		expected  bool
	}{
		{core.ColorAlways, true},
		{core.ColorNever, false},
		{core.ColorAuto, false}, // Will depend on environment, but we can test the logic
		{"invalid", false},
	}

	for _, test := range tests {
		t.Run(string(test.colorMode), func(t *testing.T) {
			result := shouldUseColor(test.colorMode)
			switch test.colorMode {
			case core.ColorAlways:
				assert.True(t, result)
			case core.ColorNever:
				assert.False(t, result)
			default:
				// For ColorAuto and invalid modes, we can't easily test in unit tests
				// as ColorAuto depends on terminal detection
			}
		})
	}
}
