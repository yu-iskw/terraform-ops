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
	"github.com/yu/terraform-ops/internal/core"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	assert.NotNil(t, builder)
}

func TestBuildGraph_EmptyPlan(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Empty(t, graphData.Nodes)
	assert.Empty(t, graphData.Edges)
}

func TestBuildGraph_SimpleResources(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "aws_instance.web",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_instance",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"instance_type": "t3.micro"},
				},
			},
			{
				Address:       "aws_security_group.web",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_security_group",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"name": "web-sg"},
				},
			},
		},
		OutputChanges: make(map[string]core.OutputChange),
		Variables:     make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Len(t, graphData.Nodes, 2)
	assert.Empty(t, graphData.Edges) // No dependencies defined

	// Check first node
	assert.Equal(t, "aws_instance_web", graphData.Nodes[0].ID)
	assert.Equal(t, "aws_instance.web", graphData.Nodes[0].Address)
	assert.Equal(t, string(core.NodeTypeResource), graphData.Nodes[0].Type)
	assert.Equal(t, "web", graphData.Nodes[0].Name)
	assert.Equal(t, "aws", graphData.Nodes[0].Provider)
}

func TestBuildGraph_WithDataSources(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "data.aws_ami.ubuntu",
				ModuleAddress: "",
				Mode:          "data",
				Type:          "aws_ami",
				Name:          "ubuntu",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  nil,
					After:   map[string]interface{}{"most_recent": true},
				},
			},
		},
		OutputChanges: make(map[string]core.OutputChange),
		Variables:     make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Len(t, graphData.Nodes, 1)

	assert.Equal(t, "data_aws_ami_ubuntu", graphData.Nodes[0].ID)
	assert.Equal(t, "data.aws_ami.ubuntu", graphData.Nodes[0].Address)
	assert.Equal(t, string(core.NodeTypeData), graphData.Nodes[0].Type)
	assert.Equal(t, "ubuntu", graphData.Nodes[0].Name)
	assert.Equal(t, "", graphData.Nodes[0].Module)
	assert.Equal(t, "aws", graphData.Nodes[0].Provider)
	assert.Equal(t, []string{"read"}, graphData.Nodes[0].Actions)
}

func TestBuildGraph_NoDataSources(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "data.aws_ami.ubuntu",
				ModuleAddress: "",
				Mode:          "data",
				Type:          "aws_ami",
				Name:          "ubuntu",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  nil,
					After:   map[string]interface{}{"most_recent": true},
				},
			},
		},
		OutputChanges: make(map[string]core.OutputChange),
		Variables:     make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: true, // Exclude data sources
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Empty(t, graphData.Nodes) // Data source should be excluded
	assert.Empty(t, graphData.Edges)
}

func TestBuildGraph_WithOutputs(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges: map[string]core.OutputChange{
			"instance_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "i-1234567890abcdef0",
				},
			},
		},
		Variables: make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Len(t, graphData.Nodes, 1)

	assert.Equal(t, "output_instance_id", graphData.Nodes[0].ID)
	assert.Equal(t, "output.instance_id", graphData.Nodes[0].Address)
	assert.Equal(t, string(core.NodeTypeOutput), graphData.Nodes[0].Type)
	assert.Equal(t, "instance_id", graphData.Nodes[0].Name)
	assert.Equal(t, "", graphData.Nodes[0].Module)
	assert.Equal(t, []string{"create"}, graphData.Nodes[0].Actions)
}

func TestBuildGraph_NoOutputs(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges: map[string]core.OutputChange{
			"instance_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "i-1234567890abcdef0",
				},
			},
		},
		Variables: make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     true, // Exclude outputs
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Empty(t, graphData.Nodes) // Output should be excluded
	assert.Empty(t, graphData.Edges)
}

func TestBuildGraph_WithVariables(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables: map[string]core.Variable{
			"region": {Value: "us-west-2"},
			"zone":   {Value: "us-west-2a"},
		},
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Len(t, graphData.Nodes, 2)

	// Check variables are added
	var varNames []string
	for _, node := range graphData.Nodes {
		if node.Type == "variable" {
			varNames = append(varNames, node.Name)
		}
	}
	assert.Contains(t, varNames, "region")
	assert.Contains(t, varNames, "zone")
}

func TestBuildGraph_WithLocals(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion:   "1.0",
		ResourceChanges: []core.ResourceChange{},
		OutputChanges:   make(map[string]core.OutputChange),
		Variables:       make(map[string]core.Variable),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources:   []core.ConfigurationResource{},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals: map[string]core.LocalConfig{
					"common_tags": {Expression: map[string]interface{}{}},
					"env":         {Expression: map[string]interface{}{}},
				},
			},
		},
		Applicable: true,
		Complete:   true,
		Errored:    false,
	}

	opts := core.GraphOptions{
		Format:        core.FormatGraphviz,
		NoDataSources: false,
		NoOutputs:     false,
		NoVariables:   false,
		NoLocals:      false,
		Verbose:       false,
	}

	builder := NewBuilder()
	graphData, err := builder.BuildGraph(plan, opts)

	assert.NoError(t, err)
	assert.NotNil(t, graphData)
	assert.Len(t, graphData.Nodes, 2)

	// Check locals are added
	var localNames []string
	for _, node := range graphData.Nodes {
		if node.Type == "local" {
			localNames = append(localNames, node.Name)
		}
	}
	assert.Contains(t, localNames, "common_tags")
	assert.Contains(t, localNames, "env")
}

func TestSanitizeID(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"aws_instance.web", "aws_instance_web"},
		{"module.app.aws_instance.web", "module_app_aws_instance_web"},
		{"data.aws_ami.ubuntu", "data_aws_ami_ubuntu"},
		{"output.instance_id", "output_instance_id"},
		{"var.region", "var_region"},
		{"local.common_tags", "local_common_tags"},
		{"aws_instance.web[0]", "aws_instance_web_0_"},
		{"aws_instance.web[\"prod\"]", "aws_instance_web_\"prod\"_"},
		{"aws_instance.web(prod)", "aws_instance_web_prod_"},
		{"aws instance web", "aws_instance_web"},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := sanitizeID(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestHasSensitiveValues(t *testing.T) {
	// Test with nil
	assert.False(t, hasSensitiveValues(nil))

	// Test with false
	assert.False(t, hasSensitiveValues(false))

	// Test with true
	assert.False(t, hasSensitiveValues(true))

	// Test with map
	sensitiveMap := map[string]interface{}{
		"password": true,
		"secret":   false,
	}
	assert.False(t, hasSensitiveValues(sensitiveMap))
}

func TestIsResourceType(t *testing.T) {
	testCases := []struct {
		input    string
		expected bool
	}{
		{"aws_instance", true},
		{"google_compute_instance", true},
		{"azurerm_virtual_machine", true},
		{"kubernetes_pod", true},
		{"docker_container", true},
		{"null_resource", true},
		{"random_string", true},
		{"local_file", true},
		{"template_file", true},
		{"archive_file", true},
		{"external", false}, // Single word, not a resource type
		{"http", false},     // Single word, not a resource type
		{"tls_private_key", true},
		{"time_static", true},
		{"custom_provider_resource", true}, // Any provider_resource pattern
		{"unknown_type", true},             // Any provider_resource pattern
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isResourceType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestExtractProviderFromType(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"aws_instance", "aws"},
		{"google_compute_instance", "google"},
		{"azurerm_virtual_machine", "azurerm"},
		{"kubernetes_pod", "kubernetes"},
		{"docker_container", "docker"},
		{"null_resource", "null"},
		{"random_string", "random"},
		{"local_file", "local"},
		{"template_file", "template"},
		{"archive_file", "archive"},
		{"tls_private_key", "tls"},
		{"time_static", "time"},
		{"custom_provider_resource", "custom"},
		{"unknown_type", "unknown"},
		{"external", ""}, // Single word, no provider
		{"http", ""},     // Single word, no provider
		{"", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := extractProviderFromType(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}
