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

package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/ir"
)

func TestBuildGraphProjectsNormalizedNodesAndEdges(t *testing.T) {
	module := ir.Address("module.child")
	changeSet := &ir.ChangeSet{
		Resources: []ir.ResourceChange{
			{
				Address: "aws_instance.web", Mode: ir.ResourceModeManaged, Type: "aws_instance", Name: "web",
				Action: ir.NormalizeAction([]string{"update"}),
			},
			{
				Address: "data.aws_ami.latest", Mode: ir.ResourceModeData, Type: "aws_ami", Name: "latest",
				Action: ir.NormalizeAction([]string{"read"}),
			},
			{
				Address: "module.child.aws_instance.worker", ModuleAddress: &module, Mode: ir.ResourceModeManaged,
				Type: "aws_instance", Name: "worker", Action: ir.NormalizeAction([]string{"create"}),
			},
		},
		Outputs: []ir.OutputChange{{Name: "id", Action: ir.NormalizeAction([]string{"update"})}},
		Graph: ir.DependencyGraph{
			Nodes: []ir.Node{
				{ID: "var.region", Address: "var.region", Kind: ir.NodeKindVariable},
				{ID: "data.aws_ami.latest", Address: "data.aws_ami.latest", Kind: ir.NodeKindData},
				{ID: "aws_instance.web", Address: "aws_instance.web", Kind: ir.NodeKindResource},
				{ID: "module.child.aws_instance.worker", Address: "module.child.aws_instance.worker", Kind: ir.NodeKindResource},
				{ID: "output.id", Address: "output.id", Kind: ir.NodeKindOutput},
			},
			Edges: []ir.Edge{
				{From: "var.region", To: "aws_instance.web", Kind: ir.EdgeVariableReference, Confidence: ir.ConfidenceExact},
				{From: "data.aws_ami.latest", To: "aws_instance.web", Kind: ir.EdgeExpressionRef, Confidence: ir.ConfidenceExact},
				{From: "aws_instance.web", To: "module.child.aws_instance.worker", Kind: ir.EdgeModuleInput, Confidence: ir.ConfidenceExact},
				{From: "aws_instance.web", To: "module.child.aws_instance.worker", Kind: ir.EdgeExpressionRef, Confidence: ir.ConfidenceStrong},
				{From: "module.child.aws_instance.worker", To: "output.id", Kind: ir.EdgeOutputReference, Confidence: ir.ConfidenceExact},
			},
		},
	}

	got, err := NewBuilder().BuildGraph(changeSet, core.GraphOptions{})
	require.NoError(t, err)
	require.Len(t, got.Nodes, 5)
	// Two normalized evidence edges connect web -> worker, but a renderer needs
	// one visual edge for that pair.
	require.Len(t, got.Edges, 4)

	byAddress := make(map[string]core.GraphNode)
	for _, node := range got.Nodes {
		byAddress[node.Address] = node
	}
	assert.Equal(t, "aws", byAddress["aws_instance.web"].Provider)
	assert.Equal(t, []string{"update"}, byAddress["aws_instance.web"].Actions)
	assert.Equal(t, string(core.NodeTypeVariable), byAddress["var.region"].Type)
	assert.Equal(t, string(core.NodeTypeOutput), byAddress["output.id"].Type)
	assert.Equal(t, "module.child", byAddress["module.child.aws_instance.worker"].Module)
}

func TestBuildGraphFiltersNormalizedNodeKinds(t *testing.T) {
	module := ir.Address("module.child")
	changeSet := &ir.ChangeSet{
		Resources: []ir.ResourceChange{
			{Address: "data.test.lookup", Mode: ir.ResourceModeData, Type: "test_lookup", Name: "lookup"},
			{Address: "module.child.test_resource.item", ModuleAddress: &module, Mode: ir.ResourceModeManaged, Type: "test_resource", Name: "item"},
		},
		Outputs: []ir.OutputChange{{Name: "result"}},
		Graph: ir.DependencyGraph{
			Nodes: []ir.Node{
				{ID: "var.input", Address: "var.input", Kind: ir.NodeKindVariable},
				{ID: "data.test.lookup", Address: "data.test.lookup", Kind: ir.NodeKindData},
				{ID: "module.child.test_resource.item", Address: "module.child.test_resource.item", Kind: ir.NodeKindResource},
				{ID: "output.result", Address: "output.result", Kind: ir.NodeKindOutput},
			},
			Edges: []ir.Edge{{From: "var.input", To: "module.child.test_resource.item", Kind: ir.EdgeModuleInput}},
		},
	}

	got, err := NewBuilder().BuildGraph(changeSet, core.GraphOptions{
		NoDataSources: true,
		NoOutputs:     true,
		NoVariables:   true,
		NoModules:     true,
	})
	require.NoError(t, err)
	assert.Empty(t, got.Nodes)
	assert.Empty(t, got.Edges)
}

func TestBuildGraphRejectsNilChangeSet(t *testing.T) {
	_, err := NewBuilder().BuildGraph(nil, core.GraphOptions{})
	require.Error(t, err)
}

func TestSanitizeID(t *testing.T) {
	assert.Equal(t, "module_child_aws_instance_web_0_", sanitizeID("module.child.aws-instance.web[0]"))
}
