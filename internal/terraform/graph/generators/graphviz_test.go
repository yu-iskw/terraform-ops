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

package generators

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yu/terraform-ops/internal/core"
)

func TestNewGraphvizGenerator(t *testing.T) {
	generator := NewGraphvizGenerator()
	assert.NotNil(t, generator)
}

func TestGenerate_EmptyGraph(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the basic Graphviz structure
	assert.Contains(t, output, "digraph terraform_plan")
	assert.Contains(t, output, "rankdir=TB")
	assert.Contains(t, output, "node [shape=box, style=filled, fontname=\"Arial\"]")
	assert.Contains(t, output, "edge [fontname=\"Arial\"]")
}

func TestGenerate_SingleResource(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the node definition
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "lightgreen") // Color for create action
}

func TestGenerate_MultipleResources(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "aws_security_group_web",
				Address:   "aws_security_group.web",
				Type:      "aws_security_group",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains both node definitions
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "aws_security_group_web")
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "aws_security_group.web")
}

func TestGenerate_WithEdges(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "aws_security_group_web",
				Address:   "aws_security_group.web",
				Type:      "aws_security_group",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{
			{
				From: "aws_instance_web",
				To:   "aws_security_group_web",
			},
		},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the edge definition
	assert.Contains(t, output, "aws_instance_web -> aws_security_group_web")
}

func TestGenerate_WithModules(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "module_app_aws_instance_web",
				Address:   "module.app.aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "module.app",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains module clusters
	assert.Contains(t, output, "subgraph cluster_")
	assert.Contains(t, output, "root")
	assert.Contains(t, output, "module.app")
}

func TestGenerate_DifferentActionTypes(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_create",
				Address:   "aws_instance.create",
				Type:      "aws_instance",
				Name:      "create",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "aws_instance_update",
				Address:   "aws_instance.update",
				Type:      "aws_instance",
				Name:      "update",
				Module:    "",
				Actions:   []string{"update"},
				Sensitive: false,
			},
			{
				ID:        "aws_instance_delete",
				Address:   "aws_instance.delete",
				Type:      "aws_instance",
				Name:      "delete",
				Module:    "",
				Actions:   []string{"delete"},
				Sensitive: false,
			},
			{
				ID:        "aws_instance_replace",
				Address:   "aws_instance.replace",
				Type:      "aws_instance",
				Name:      "replace",
				Module:    "",
				Actions:   []string{"create", "delete"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that different actions have different colors
	assert.Contains(t, output, "lightgreen")  // create
	assert.Contains(t, output, "lightyellow") // update
	assert.Contains(t, output, "lightcoral")  // delete
	assert.Contains(t, output, "orange")      // replace
}

func TestGenerate_DifferentNodeTypes(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "data_aws_ami_ubuntu",
				Address:   "data.aws_ami.ubuntu",
				Type:      "aws_ami",
				Name:      "ubuntu",
				Module:    "",
				Actions:   []string{"read"},
				Sensitive: false,
			},
			{
				ID:        "output_instance_id",
				Address:   "output.instance_id",
				Type:      "output",
				Name:      "instance_id",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "var_region",
				Address:   "var.region",
				Type:      "variable",
				Name:      "region",
				Module:    "",
				Actions:   []string{"no-op"},
				Sensitive: false,
			},
			{
				ID:        "local_common_tags",
				Address:   "local.common_tags",
				Type:      "local",
				Name:      "common_tags",
				Module:    "",
				Actions:   []string{"no-op"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that different node types have different colors and shapes
	assert.Contains(t, output, "lightsteelblue") // output (inverted house)
	assert.Contains(t, output, "lightyellow")    // variable (cylinder)
	assert.Contains(t, output, "lightpink")      // local (octagon)
}

func TestGenerate_ComplexGraph(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_vpc_main",
				Address:   "aws_vpc.main",
				Type:      "aws_vpc",
				Name:      "main",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "aws_subnet_public",
				Address:   "aws_subnet.public",
				Type:      "aws_subnet",
				Name:      "public",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "output_vpc_id",
				Address:   "output.vpc_id",
				Type:      "output",
				Name:      "vpc_id",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{
			{
				From: "aws_subnet_public",
				To:   "aws_vpc_main",
			},
			{
				From: "aws_instance_web",
				To:   "aws_subnet_public",
			},
			{
				From: "output_vpc_id",
				To:   "aws_vpc_main",
			},
		},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that all nodes are present
	assert.Contains(t, output, "aws_vpc_main")
	assert.Contains(t, output, "aws_subnet_public")
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "output_vpc_id")

	// Check that all edges are present
	assert.Contains(t, output, "aws_subnet_public -> aws_vpc_main")
	assert.Contains(t, output, "aws_instance_web -> aws_subnet_public")
	assert.Contains(t, output, "output_vpc_id -> aws_vpc_main")

	// Check that the output is valid Graphviz DOT format
	assert.True(t, strings.HasPrefix(output, "digraph terraform_plan"))
	assert.True(t, strings.HasSuffix(output, "}\n"))
}

func TestGenerate_WithSensitiveData(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      "aws_instance",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: true,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that the node is present (sensitivity is handled in the node data)
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "aws_instance.web")
}

func TestGenerate_ShapeTypes(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_web",
				Address:   "aws_instance.web",
				Type:      string(core.NodeTypeResource),
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "data_aws_ami_ubuntu",
				Address:   "data.aws_ami.ubuntu",
				Type:      string(core.NodeTypeData),
				Name:      "ubuntu",
				Module:    "",
				Actions:   []string{"read"},
				Sensitive: false,
			},
			{
				ID:        "output_instance_id",
				Address:   "output.instance_id",
				Type:      string(core.NodeTypeOutput),
				Name:      "instance_id",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "var_region",
				Address:   "var.region",
				Type:      string(core.NodeTypeVariable),
				Name:      "region",
				Module:    "",
				Actions:   []string{"no-op"},
				Sensitive: false,
			},
			{
				ID:        "local_common_tags",
				Address:   "local.common_tags",
				Type:      string(core.NodeTypeLocal),
				Name:      "common_tags",
				Module:    "",
				Actions:   []string{"no-op"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatGraphviz,
		Verbose: false,
	}

	generator := NewGraphvizGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that the correct shapes are used for each node type
	assert.Contains(t, output, "shape=house")    // Resource should use house shape
	assert.Contains(t, output, "shape=diamond")  // Data source should use diamond shape
	assert.Contains(t, output, "shape=invhouse") // Output should use inverted house shape
	assert.Contains(t, output, "shape=cylinder") // Variable should use cylinder shape
	assert.Contains(t, output, "shape=octagon")  // Local should use octagon shape
}
