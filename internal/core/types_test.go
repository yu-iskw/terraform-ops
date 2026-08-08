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

package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphFormatConstants(t *testing.T) {
	assert.Equal(t, GraphFormat("graphviz"), FormatGraphviz)
	assert.Equal(t, GraphFormat("mermaid"), FormatMermaid)
	assert.Equal(t, GraphFormat("plantuml"), FormatPlantUML)
}

func TestGroupingStrategyConstants(t *testing.T) {
	assert.Equal(t, GroupingStrategy("module"), GroupByModule)
	assert.Equal(t, GroupingStrategy("action"), GroupByAction)
	assert.Equal(t, GroupingStrategy("resource_type"), GroupByResourceType)
	assert.Equal(t, GroupingStrategy("provider"), GroupByProvider)
}

func TestActionTypeConstants(t *testing.T) {
	assert.Equal(t, ActionType("create"), ActionCreate)
	assert.Equal(t, ActionType("update"), ActionUpdate)
	assert.Equal(t, ActionType("delete"), ActionDelete)
	assert.Equal(t, ActionType("replace"), ActionReplace)
	assert.Equal(t, ActionType("no-op"), ActionNoOp)
}

func TestNodeTypeConstants(t *testing.T) {
	assert.Equal(t, NodeType("resource"), NodeTypeResource)
	assert.Equal(t, NodeType("data"), NodeTypeData)
	assert.Equal(t, NodeType("output"), NodeTypeOutput)
	assert.Equal(t, NodeType("variable"), NodeTypeVariable)
	assert.Equal(t, NodeType("local"), NodeTypeLocal)
}

func TestGraphOptionsDefaultValues(t *testing.T) {
	opts := GraphOptions{Format: FormatGraphviz, GroupBy: GroupByModule}
	assert.Equal(t, FormatGraphviz, opts.Format)
	assert.Equal(t, GroupByModule, opts.GroupBy)
	assert.False(t, opts.NoDataSources)
	assert.False(t, opts.NoOutputs)
	assert.False(t, opts.NoVariables)
	assert.False(t, opts.NoLocals)
	assert.False(t, opts.NoModules)
}

func TestGraphDataAndNode(t *testing.T) {
	graphData := &GraphData{
		Nodes: []GraphNode{{
			ID: "aws_instance_web", Address: "aws_instance.web", Type: "aws_instance",
			Name: "web", Provider: "aws", Actions: []string{"create"},
		}},
		Edges: []GraphEdge{{From: "aws_instance_web", To: "output_id"}},
	}
	assert.Len(t, graphData.Nodes, 1)
	assert.Equal(t, "aws", graphData.Nodes[0].Provider)
	assert.Len(t, graphData.Edges, 1)
}

func TestTerraformConfigRemainsShowTerraformDomain(t *testing.T) {
	config := TerraformConfig{
		Path:              "/path/to/config",
		RequiredVersion:   ">= 1.0.0",
		RequiredProviders: map[string]string{"aws": "~> 5.0"},
		Backend: &Backend{
			Type:   "s3",
			Config: map[string]string{"bucket": "terraform-state"},
		},
	}
	assert.Equal(t, "/path/to/config", config.Path)
	assert.Equal(t, ">= 1.0.0", config.RequiredVersion)
	assert.Equal(t, "s3", config.Backend.Type)
}
