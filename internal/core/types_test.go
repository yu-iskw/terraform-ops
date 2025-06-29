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

func TestTerraformPlan_EmptyPlan(t *testing.T) {
	plan := &TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []ResourceChange{},
		OutputChanges:   make(map[string]OutputChange),
		Variables:       make(map[string]Variable),
		Applicable:      true,
		Complete:        true,
		Errored:         false,
	}

	assert.Equal(t, "1.0", plan.FormatVersion)
	assert.Empty(t, plan.ResourceChanges)
	assert.Empty(t, plan.OutputChanges)
	assert.Empty(t, plan.Variables)
	assert.True(t, plan.Applicable)
	assert.True(t, plan.Complete)
	assert.False(t, plan.Errored)
}

func TestResourceChange_CompleteResource(t *testing.T) {
	change := ResourceChange{
		Address:       "aws_instance.web",
		ModuleAddress: "",
		Mode:          "managed",
		Type:          "aws_instance",
		Name:          "web",
		Change: Change{
			Actions: []string{"create"},
			Before:  nil,
			After:   map[string]interface{}{"instance_type": "t3.micro"},
		},
	}

	assert.Equal(t, "aws_instance.web", change.Address)
	assert.Equal(t, "", change.ModuleAddress)
	assert.Equal(t, "managed", change.Mode)
	assert.Equal(t, "aws_instance", change.Type)
	assert.Equal(t, "web", change.Name)
	assert.Equal(t, []string{"create"}, change.Change.Actions)
}

func TestGraphOptions_DefaultValues(t *testing.T) {
	opts := GraphOptions{
		Format:        FormatGraphviz,
		Output:        "",
		GroupBy:       GroupByModule,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Compact:       false,
		Verbose:       false,
	}

	assert.Equal(t, FormatGraphviz, opts.Format)
	assert.Equal(t, "", opts.Output)
	assert.Equal(t, GroupByModule, opts.GroupBy)
	assert.False(t, opts.NoDataSources)
	assert.False(t, opts.NoOutputs)
	assert.False(t, opts.NoVariables)
	assert.False(t, opts.NoLocals)
	assert.False(t, opts.Compact)
	assert.False(t, opts.Verbose)
}

func TestGraphData_EmptyGraph(t *testing.T) {
	graphData := &GraphData{
		Nodes: []GraphNode{},
		Edges: []GraphEdge{},
	}

	assert.Empty(t, graphData.Nodes)
	assert.Empty(t, graphData.Edges)
}

func TestGraphNode_CompleteNode(t *testing.T) {
	node := GraphNode{
		ID:        "aws_instance_web",
		Address:   "aws_instance.web",
		Type:      "aws_instance",
		Name:      "web",
		Module:    "",
		Actions:   []string{"create"},
		Sensitive: false,
	}

	assert.Equal(t, "aws_instance_web", node.ID)
	assert.Equal(t, "aws_instance.web", node.Address)
	assert.Equal(t, "aws_instance", node.Type)
	assert.Equal(t, "web", node.Name)
	assert.Equal(t, "", node.Module)
	assert.Equal(t, []string{"create"}, node.Actions)
	assert.False(t, node.Sensitive)
}

func TestGraphEdge_CompleteEdge(t *testing.T) {
	edge := GraphEdge{
		From: "aws_instance_web",
		To:   "aws_security_group_web",
	}

	assert.Equal(t, "aws_instance_web", edge.From)
	assert.Equal(t, "aws_security_group_web", edge.To)
}

func TestTerraformConfig_CompleteConfig(t *testing.T) {
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
	assert.Equal(t, map[string]string{"aws": "~> 5.0"}, config.RequiredProviders)
	assert.NotNil(t, config.Backend)
	assert.Equal(t, "s3", config.Backend.Type)
	assert.Equal(t, map[string]string{"bucket": "terraform-state"}, config.Backend.Config)
}

func TestBackend_CompleteBackend(t *testing.T) {
	backend := &Backend{
		Type: "gcs",
		Config: map[string]string{
			"bucket": "terraform-state",
			"prefix": "terraform/state",
		},
	}

	assert.Equal(t, "gcs", backend.Type)
	assert.Equal(t, map[string]string{
		"bucket": "terraform-state",
		"prefix": "terraform/state",
	}, backend.Config)
}
