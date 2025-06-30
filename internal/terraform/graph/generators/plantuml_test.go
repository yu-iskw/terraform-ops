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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yu/terraform-ops/internal/core"
)

func TestNewPlantUMLGenerator(t *testing.T) {
	generator := NewPlantUMLGenerator()
	assert.NotNil(t, generator)
}

func TestPlantUMLGenerate_EmptyGraph(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the basic PlantUML structure
	assert.Contains(t, output, "@startuml")
	assert.Contains(t, output, "@enduml")
	assert.Contains(t, output, "!theme plain")
	assert.Contains(t, output, "skinparam backgroundColor white")
}

func TestPlantUMLGenerate_SingleResource(t *testing.T) {
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
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the node definition
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "create")
}

func TestPlantUMLGenerate_MultipleResources(t *testing.T) {
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
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains both node definitions
	assert.Contains(t, output, "aws_instance_web")
	assert.Contains(t, output, "aws_security_group_web")
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "aws_security_group.web")
}

func TestPlantUMLGenerate_WithEdges(t *testing.T) {
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
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the edge definition
	assert.Contains(t, output, "aws_instance_web --> aws_security_group_web")
}

func TestPlantUMLGenerate_WithModules(t *testing.T) {
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
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains module packages
	assert.Contains(t, output, "package \"root\"")
	assert.Contains(t, output, "package \"module.app\"")
	assert.Contains(t, output, "module.app")
}

func TestPlantUMLGenerate_DifferentActionTypes(t *testing.T) {
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
			{
				ID:        "aws_instance_noop",
				Address:   "aws_instance.noop",
				Type:      "aws_instance",
				Name:      "noop",
				Module:    "",
				Actions:   []string{"no-op"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains all action types
	assert.Contains(t, output, "create")
	assert.Contains(t, output, "update")
	assert.Contains(t, output, "delete")
	assert.Contains(t, output, "replace")
	assert.Contains(t, output, "noop")
}

func TestPlantUMLGenerate_DifferentNodeTypes(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_resource",
				Address:   "aws_instance.resource",
				Type:      "aws_instance",
				Name:      "resource",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "data_aws_ami_data",
				Address:   "data.aws_ami.data",
				Type:      "data",
				Name:      "data",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "output_web_output",
				Address:   "output.web_output",
				Type:      "output",
				Name:      "web_output",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "variable_env_var",
				Address:   "variable.env_var",
				Type:      "variable",
				Name:      "env_var",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "local_computed_local",
				Address:   "local.computed_local",
				Type:      "local",
				Name:      "computed_local",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains all node types
	assert.Contains(t, output, "aws_instance.resource")
	assert.Contains(t, output, "data.aws_ami.data")
	assert.Contains(t, output, "output.web_output")
	assert.Contains(t, output, "variable.env_var")
	assert.Contains(t, output, "local.computed_local")
}

func TestPlantUMLGenerate_ComplexGraph(t *testing.T) {
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
				ID:        "aws_subnet_private",
				Address:   "aws_subnet.private",
				Type:      "aws_subnet",
				Name:      "private",
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
				ID:        "aws_security_group_web",
				Address:   "aws_security_group.web",
				Type:      "aws_security_group",
				Name:      "web",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "module_database_aws_db_instance_main",
				Address:   "module.database.aws_db_instance.main",
				Type:      "aws_db_instance",
				Name:      "main",
				Module:    "module.database",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{
			{
				From: "aws_subnet_private",
				To:   "aws_vpc_main",
			},
			{
				From: "aws_subnet_public",
				To:   "aws_vpc_main",
			},
			{
				From: "aws_instance_web",
				To:   "aws_subnet_public",
			},
			{
				From: "aws_instance_web",
				To:   "aws_security_group_web",
			},
			{
				From: "module_database_aws_db_instance_main",
				To:   "aws_subnet_private",
			},
		},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains all nodes
	assert.Contains(t, output, "aws_vpc.main")
	assert.Contains(t, output, "aws_subnet.private")
	assert.Contains(t, output, "aws_subnet.public")
	assert.Contains(t, output, "aws_instance.web")
	assert.Contains(t, output, "aws_security_group.web")
	assert.Contains(t, output, "module.database.aws_db_instance.main")

	// Check that it contains all edges
	assert.Contains(t, output, "aws_subnet_private --> aws_vpc_main")
	assert.Contains(t, output, "aws_subnet_public --> aws_vpc_main")
	assert.Contains(t, output, "aws_instance_web --> aws_subnet_public")
	assert.Contains(t, output, "aws_instance_web --> aws_security_group_web")
	assert.Contains(t, output, "module_database_aws_db_instance_main --> aws_subnet_private")

	// Check that it contains module packages
	assert.Contains(t, output, "package \"root\"")
	assert.Contains(t, output, "package \"module.database\"")
}

func TestPlantUMLGenerate_WithSensitiveData(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_sensitive",
				Address:   "aws_instance.sensitive",
				Type:      "aws_instance",
				Name:      "sensitive",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: true,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the sensitive node
	assert.Contains(t, output, "aws_instance.sensitive")
	assert.Contains(t, output, "create")
}

func TestPlantUMLGenerate_ShapeTypes(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_resource",
				Address:   "aws_instance.resource",
				Type:      "aws_instance",
				Name:      "resource",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "data_aws_ami_data",
				Address:   "data.aws_ami.data",
				Type:      "data",
				Name:      "data",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "output_web_output",
				Address:   "output.web_output",
				Type:      "output",
				Name:      "web_output",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "variable_env_var",
				Address:   "variable.env_var",
				Type:      "variable",
				Name:      "env_var",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
			{
				ID:        "local_computed_local",
				Address:   "local.computed_local",
				Type:      "local",
				Name:      "computed_local",
				Module:    "",
				Actions:   []string{},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that the output contains proper PlantUML syntax
	assert.Contains(t, output, "@startuml")
	assert.Contains(t, output, "@enduml")
	assert.Contains(t, output, "package \"root\"")
	assert.Contains(t, output, "aws_instance_resource")
	assert.Contains(t, output, "data_aws_ami_data")
	assert.Contains(t, output, "output_web_output")
	assert.Contains(t, output, "variable_env_var")
	assert.Contains(t, output, "local_computed_local")
}

func TestPlantUMLGenerate_ThemeConfiguration(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains the theme configuration
	assert.Contains(t, output, "!theme plain")
	assert.Contains(t, output, "skinparam backgroundColor white")
	assert.Contains(t, output, "skinparam defaultFontName Arial")
}

func TestPlantUMLGenerate_EmptyModuleName(t *testing.T) {
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
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that empty module name is converted to "root"
	assert.Contains(t, output, "package \"root\"")
}

func TestPlantUMLGenerate_MultipleActions(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_replace",
				Address:   "aws_instance.replace",
				Type:      "aws_instance",
				Name:      "replace",
				Module:    "",
				Actions:   []string{"delete", "create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that multiple actions are handled correctly
	assert.Contains(t, output, "replace")
}

func TestPlantUMLGenerate_NodeLabelFormat(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "aws_instance_test",
				Address:   "aws_instance.test",
				Type:      "aws_instance",
				Name:      "test",
				Module:    "",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that the label format is correct (address + action)
	assert.Contains(t, output, "aws_instance.test\\n[create]")
}

func TestPlantUMLGenerate_ComplexModuleStructure(t *testing.T) {
	graphData := &core.GraphData{
		Nodes: []core.GraphNode{
			{
				ID:        "module_web_aws_instance_app",
				Address:   "module.web.aws_instance.app",
				Type:      "aws_instance",
				Name:      "app",
				Module:    "module.web",
				Actions:   []string{"create"},
				Sensitive: false,
			},
			{
				ID:        "module_web_module_api_aws_lambda_function",
				Address:   "module.web.module.api.aws_lambda_function.main",
				Type:      "aws_lambda_function",
				Name:      "main",
				Module:    "module.web.module.api",
				Actions:   []string{"create"},
				Sensitive: false,
			},
		},
		Edges: []core.GraphEdge{},
	}

	opts := core.GraphOptions{
		Format:  core.FormatPlantUML,
		Verbose: false,
	}

	generator := NewPlantUMLGenerator()
	output, err := generator.Generate(graphData, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Check that it contains nested module packages
	assert.Contains(t, output, "package \"module.web\"")
	assert.Contains(t, output, "package \"module.web.module.api\"")
	assert.Contains(t, output, "module.web.aws_instance.app")
	assert.Contains(t, output, "module.web.module.api.aws_lambda_function.main")
}
