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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/ir"
)

func TestSummarizePlanProjectsChangeSet(t *testing.T) {
	module := ir.Address("module.database")
	changeSet := &ir.ChangeSet{
		Source: ir.SourceMetadata{PlanFormatVersion: "1.0"},
		Plan:   ir.PlanMetadata{Applyable: true, Complete: true},
		Resources: []ir.ResourceChange{
			{
				Address: "aws_instance.web",
				Mode:    ir.ResourceModeManaged,
				Type:    "aws_instance",
				Name:    "web",
				Action:  ir.NormalizeAction([]string{"create"}),
				After: ir.SafeValue{Value: map[string]interface{}{
					"instance_type": "t3.micro",
				}},
			},
			{
				Address: "aws_security_group.web",
				Mode:    ir.ResourceModeManaged,
				Type:    "aws_security_group",
				Name:    "web",
				Action:  ir.NormalizeAction([]string{"update"}),
				Before:  ir.SafeValue{Value: map[string]interface{}{"name": "old"}},
				After:   ir.SafeValue{Value: map[string]interface{}{"name": "new"}},
			},
			{
				Address:       "module.database.aws_instance.db",
				ModuleAddress: &module,
				Mode:          ir.ResourceModeManaged,
				Type:          "aws_instance",
				Name:          "db",
				Action:        ir.NormalizeAction([]string{"delete"}),
				Before:        ir.SafeValue{Value: map[string]interface{}{"size": "small"}},
			},
			{
				Address: "aws_instance.replace",
				Mode:    ir.ResourceModeManaged,
				Type:    "aws_instance",
				Name:    "replace",
				Action:  ir.NormalizeAction([]string{"delete", "create"}),
				Before:  ir.SafeValue{Value: map[string]interface{}{"type": "old"}},
				After:   ir.SafeValue{Value: map[string]interface{}{"type": "new"}},
			},
		},
		Outputs: []ir.OutputChange{
			{
				Name:   "instance_id",
				Action: ir.NormalizeAction([]string{"create"}),
				After:  ir.SafeValue{Value: "i-123"},
			},
			{
				Name:           "secret",
				Action:         ir.NormalizeAction([]string{"update"}),
				After:          ir.SafeValue{Redacted: true},
				SensitivePaths: []ir.AttributePath{{}},
			},
		},
	}

	got, err := NewSummarizer().SummarizePlan(changeSet, core.SummaryOptions{ShowDetails: true})
	require.NoError(t, err)

	assert.Equal(t, core.PlanInfo{FormatVersion: "1.0", Applicable: true, Complete: true}, got.PlanInfo)
	assert.Equal(t, 4, got.Statistics.TotalChanges)
	assert.Equal(t, 2, got.Statistics.ActionBreakdown["create"])
	assert.Equal(t, 1, got.Statistics.ActionBreakdown["update"])
	assert.Equal(t, 2, got.Statistics.ActionBreakdown["delete"])
	assert.Equal(t, 4, got.Statistics.ProviderBreakdown["aws"])
	assert.Equal(t, 3, got.Statistics.ModuleBreakdown["root"])
	assert.Equal(t, 1, got.Statistics.ModuleBreakdown["module.database"])

	require.Len(t, got.Changes.Create, 1)
	require.Len(t, got.Changes.Update, 1)
	require.Len(t, got.Changes.Delete, 1)
	require.Len(t, got.Changes.Replace, 1)
	assert.Equal(t, "t3.micro", got.Changes.Create[0].KeyChanges["instance_type"].(map[string]interface{})["to"])
	assert.Equal(t, "new", got.Changes.Update[0].KeyChanges["name"].(map[string]interface{})["to"])

	require.Len(t, got.Outputs, 2)
	assert.Equal(t, "i-123", got.Outputs[0].Value)
	assert.True(t, got.Outputs[1].Sensitive)
	assert.Nil(t, got.Outputs[1].Value)
}

func TestSummarizePlanRejectsNilChangeSet(t *testing.T) {
	_, err := NewSummarizer().SummarizePlan(nil, core.SummaryOptions{})
	require.Error(t, err)
}

func TestExtractKeyChangesDoesNotExposeFullyRedactedValues(t *testing.T) {
	change := ir.ResourceChange{
		Before: ir.SafeValue{Redacted: true},
		After:  ir.SafeValue{Value: map[string]interface{}{"visible": "value"}},
	}
	assert.Nil(t, extractKeyChanges(change))
}

func TestExtractProviderFromType(t *testing.T) {
	assert.Equal(t, "aws", extractProviderFromType("aws_instance"))
	assert.Equal(t, "google", extractProviderFromType("google_compute_instance"))
	assert.Equal(t, "unknown", extractProviderFromType("terraform"))
}
