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
	"github.com/yu/terraform-ops/internal/terraform/summary"
	"github.com/yu/terraform-ops/internal/terraform/summary/formatters"
)

func TestNewSummarizePlanCommand(t *testing.T) {
	planSummarizer := summary.NewSummarizer()
	formatterFactory := formatters.NewFactory()
	cmd := NewSummarizePlanCommand(planSummarizer, formatterFactory)
	assert.NotNil(t, cmd)
	assert.Equal(t, planSummarizer, cmd.planSummarizer)
	assert.Equal(t, formatterFactory, cmd.formatterFactory)
}

func TestDefaultSummarizePlanCommand(t *testing.T) {
	cmd := DefaultSummarizePlanCommand()
	assert.NotNil(t, cmd)
	assert.NotNil(t, cmd.planSummarizer)
	assert.NotNil(t, cmd.formatterFactory)
}

func TestSummarizePlanCommandCommand(t *testing.T) {
	command := DefaultSummarizePlanCommand().Command()
	assert.Equal(t, "summarize-plan <PLAN_FILE>", command.Use)
	assert.Equal(t, "Generate a human-readable summary of Terraform plan changes", command.Short)
	assert.NoError(t, command.Args(command, []string{"plan.json"}))
	assert.Error(t, command.Args(command, nil))
	assert.Error(t, command.Args(command, []string{"one", "two"}))
}

func TestIsValidSummaryFormat(t *testing.T) {
	for _, format := range []core.SummaryFormat{core.FormatText, core.FormatJSON, core.FormatMarkdown, core.FormatTable, core.FormatPlan} {
		assert.True(t, isValidSummaryFormat(format))
	}
	assert.False(t, isValidSummaryFormat("invalid"))
}

func TestIsValidSummaryGrouping(t *testing.T) {
	for _, grouping := range []core.SummaryGrouping{core.GroupByAction, core.GroupByModule, core.GroupByProvider, core.GroupByResourceType} {
		assert.True(t, isValidSummaryGrouping(grouping))
	}
	assert.False(t, isValidSummaryGrouping("invalid"))
}

func TestShouldUseColor(t *testing.T) {
	assert.True(t, shouldUseColor(core.ColorAlways))
	assert.False(t, shouldUseColor(core.ColorNever))
	assert.False(t, shouldUseColor("invalid"))
}
