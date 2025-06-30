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

// TestAnalyzeDependencies_EmptyPlan tests dependency analysis with an empty plan
func TestAnalyzeDependencies_EmptyPlan(t *testing.T) {
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
	edges, err := builder.analyzeDependencies(plan, opts)

	assert.NoError(t, err)
	assert.Empty(t, edges)
}

func TestIsResourceReference(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		expected bool
	}{
		{
			name:     "resource reference",
			ref:      "random_id.test_id",
			expected: true,
		},
		{
			name:     "resource reference with attribute",
			ref:      "random_id.test_id.hex",
			expected: true,
		},
		{
			name:     "module reference",
			ref:      "module.network.google_compute_network.main",
			expected: true,
		},
		{
			name:     "local reference",
			ref:      "local.test_tag",
			expected: true,
		},
		{
			name:     "variable reference",
			ref:      "var.environment",
			expected: true,
		},
		{
			name:     "invalid reference",
			ref:      "invalid",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isResourceReference(tt.ref, core.GraphOptions{})
			assert.Equal(t, tt.expected, result, "Expected %s to be %v, got %v", tt.ref, tt.expected, result)
		})
	}
}

func TestAnalyzeDependencies_DataResourceOutputInterdependencies(t *testing.T) {
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			// Referenced VPC resource
			{
				Address:       "aws_vpc.main",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_vpc",
				Name:          "main",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"cidr_block": "10.0.0.0/16"},
				},
			},
			// Referenced data.aws_region.current
			{
				Address:       "data.aws_region.current",
				ModuleAddress: "",
				Mode:          "data",
				Type:          "aws_region",
				Name:          "current",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  map[string]interface{}{"name": "us-west-2"},
					After:   map[string]interface{}{"name": "us-west-2"},
				},
			},
			// Data source that depends on a resource
			{
				Address:       "data.aws_subnet.selected",
				ModuleAddress: "",
				Mode:          "data",
				Type:          "aws_subnet",
				Name:          "selected",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  map[string]interface{}{"id": "subnet-123"},
					After:   map[string]interface{}{"id": "subnet-123"},
				},
			},
			// Resource that depends on a data source
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
			// Another resource that depends on the first resource
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
			// Data source that depends on another data source
			{
				Address:       "data.aws_availability_zones.available",
				ModuleAddress: "",
				Mode:          "data",
				Type:          "aws_availability_zones",
				Name:          "available",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  map[string]interface{}{"names": []string{"us-west-2a"}},
					After:   map[string]interface{}{"names": []string{"us-west-2a"}},
				},
			},
		},
		OutputChanges: map[string]core.OutputChange{
			"instance_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "i-1234567890abcdef0",
				},
			},
			"subnet_info": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"id": "subnet-123"},
				},
			},
		},
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources: []core.ConfigurationResource{
					// VPC resource configuration
					{
						Address: "aws_vpc.main",
						Mode:    "managed",
						Type:    "aws_vpc",
						Name:    "main",
						Expressions: map[string]interface{}{
							"cidr_block": map[string]interface{}{
								"constant_value": "10.0.0.0/16",
							},
						},
					},
					// Data source configuration with dependency on resource
					{
						Address: "data.aws_subnet.selected",
						Mode:    "data",
						Type:    "aws_subnet",
						Name:    "selected",
						Expressions: map[string]interface{}{
							"vpc_id": map[string]interface{}{
								"references": []interface{}{
									"aws_vpc.main.id",
								},
							},
						},
					},
					// Resource configuration with dependency on data source
					{
						Address: "aws_instance.web",
						Mode:    "managed",
						Type:    "aws_instance",
						Name:    "web",
						Expressions: map[string]interface{}{
							"subnet_id": map[string]interface{}{
								"references": []interface{}{
									"data.aws_subnet.selected.id",
								},
							},
							"availability_zone": map[string]interface{}{
								"references": []interface{}{
									"data.aws_availability_zones.available.names",
								},
							},
						},
					},
					// Resource configuration with dependency on another resource
					{
						Address: "aws_security_group.web",
						Mode:    "managed",
						Type:    "aws_security_group",
						Name:    "web",
						Expressions: map[string]interface{}{
							"vpc_id": map[string]interface{}{
								"references": []interface{}{
									"aws_instance.web.vpc_id",
								},
							},
						},
					},
					// Data source configuration with dependency on another data source
					{
						Address: "data.aws_availability_zones.available",
						Mode:    "data",
						Type:    "aws_availability_zones",
						Name:    "available",
						Expressions: map[string]interface{}{
							"state": map[string]interface{}{
								"references": []interface{}{
									"data.aws_region.current.name",
								},
							},
						},
					},
					// Region data source configuration
					{
						Address: "data.aws_region.current",
						Mode:    "data",
						Type:    "aws_region",
						Name:    "current",
						Expressions: map[string]interface{}{
							"name": map[string]interface{}{
								"constant_value": "us-west-2",
							},
						},
					},
				},
				Outputs: map[string]core.OutputConfig{
					"instance_id": {
						Expression: map[string]interface{}{
							"references": []interface{}{
								"aws_instance.web.id",
							},
						},
					},
					"subnet_info": {
						Expression: map[string]interface{}{
							"references": []interface{}{
								"data.aws_subnet.selected.id",
							},
						},
					},
				},
				ModuleCalls: make(map[string]core.ModuleCall),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Variables:  make(map[string]core.Variable),
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
		Verbose:       true,
	}

	builder := NewBuilder()
	edges, err := builder.analyzeDependencies(plan, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, edges, "Should have dependency edges")

	// Convert edges to a map for easier testing
	edgeMap := make(map[string][]string)
	for _, edge := range edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge.To)
	}

	// Test data source -> resource dependency (resource depends on data source)
	assert.Contains(t, edgeMap, "data_aws_subnet_selected", "Data source should be a dependency")
	assert.Contains(t, edgeMap["data_aws_subnet_selected"], "aws_instance_web", "Data source should point to dependent resource")

	// Test resource -> resource dependency (security group depends on instance)
	assert.Contains(t, edgeMap, "aws_instance_web", "Instance should be a dependency")
	assert.Contains(t, edgeMap["aws_instance_web"], "aws_security_group_web", "Instance should point to dependent security group")

	// Test data source -> resource dependency (instance depends on availability zones)
	assert.Contains(t, edgeMap, "data_aws_availability_zones_available", "Availability zones should be a dependency")
	assert.Contains(t, edgeMap["data_aws_availability_zones_available"], "aws_instance_web", "Availability zones should point to dependent instance")

	// Test resource -> output dependency (output depends on resource)
	assert.Contains(t, edgeMap, "aws_instance_web", "Instance should be a dependency")
	assert.Contains(t, edgeMap["aws_instance_web"], "output_instance_id", "Instance should point to dependent output")

	// Test data source -> output dependency (output depends on data source)
	assert.Contains(t, edgeMap, "data_aws_subnet_selected", "Data source should be a dependency")
	assert.Contains(t, edgeMap["data_aws_subnet_selected"], "output_subnet_info", "Data source should point to dependent output")
}

func TestAnalyzeDependencies_ComplexModuleDependencies(t *testing.T) {
	// Test dependencies within and between modules
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			// Root module resource
			{
				Address:       "aws_vpc.main",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_vpc",
				Name:          "main",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"cidr_block": "10.0.0.0/16"},
				},
			},
			// Module resource (network)
			{
				Address:       "module.network.aws_subnet.public",
				ModuleAddress: "module.network",
				Mode:          "managed",
				Type:          "aws_subnet",
				Name:          "public",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"vpc_id": "aws_vpc.main.id"},
				},
			},
			// App module resource
			{
				Address:       "module.app.aws_instance.web",
				ModuleAddress: "module.app",
				Mode:          "managed",
				Type:          "aws_instance",
				Name:          "web",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"subnet_id": "module.network.aws_subnet.public.id"},
				},
			},
			// App module data source
			{
				Address:       "module.app.data.aws_subnet.selected",
				ModuleAddress: "module.app",
				Mode:          "data",
				Type:          "aws_subnet",
				Name:          "selected",
				Change: core.Change{
					Actions: []string{"read"},
					Before:  nil,
					After:   map[string]interface{}{"id": "module.network.aws_subnet.public.id"},
				},
			},
		},
		OutputChanges: map[string]core.OutputChange{
			"vpc_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "vpc-1234567890abcdef0",
				},
			},
			"subnet_id": {
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   "subnet-1234567890abcdef0",
				},
			},
		},
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources: []core.ConfigurationResource{
					{
						Address: "aws_vpc.main",
						Type:    "aws_vpc",
						Name:    "main",
						Expressions: map[string]interface{}{
							"cidr_block": map[string]interface{}{
								"constant_value": "10.0.0.0/16",
							},
						},
					},
				},
				ModuleCalls: map[string]core.ModuleCall{
					"network": {
						Source: "./modules/network",
						Expressions: map[string]interface{}{
							"vpc_id": map[string]interface{}{
								"references": []interface{}{"aws_vpc.main.id"},
							},
						},
						Module: &core.ModuleConfig{
							Resources: []core.ConfigurationResource{
								{
									Address: "module.network.aws_subnet.public",
									Type:    "aws_subnet",
									Name:    "public",
									Expressions: map[string]interface{}{
										"vpc_id": map[string]interface{}{
											"references": []interface{}{"aws_vpc.main.id"},
										},
									},
								},
							},
							Outputs: map[string]interface{}{},
						},
					},
					"app": {
						Source: "./modules/app",
						Expressions: map[string]interface{}{
							"subnet_id": map[string]interface{}{
								"references": []interface{}{"module.network.aws_subnet.public.id"},
							},
						},
						Module: &core.ModuleConfig{
							Resources: []core.ConfigurationResource{
								{
									Address: "module.app.aws_instance.web",
									Type:    "aws_instance",
									Name:    "web",
									Expressions: map[string]interface{}{
										"subnet_id": map[string]interface{}{
											"references": []interface{}{"module.network.aws_subnet.public.id"},
										},
									},
								},
								{
									Address: "module.app.data.aws_subnet.selected",
									Type:    "aws_subnet",
									Name:    "selected",
									Mode:    "data",
									Expressions: map[string]interface{}{
										"id": map[string]interface{}{
											"references": []interface{}{"module.network.aws_subnet.public.id"},
										},
									},
								},
							},
							Outputs: map[string]interface{}{},
						},
					},
				},
				Outputs: map[string]core.OutputConfig{
					"vpc_id": {
						Expression: map[string]interface{}{
							"references": []interface{}{"aws_vpc.main.id"},
						},
					},
					"subnet_id": {
						Expression: map[string]interface{}{
							"references": []interface{}{"module.network.aws_subnet.public.id"},
						},
					},
				},
			},
		},
		Variables:  make(map[string]core.Variable),
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
		Verbose:       true,
	}

	builder := NewBuilder()
	edges, err := builder.analyzeDependencies(plan, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, edges, "Should have dependency edges")

	// Convert edges to a map for easier testing
	edgeMap := make(map[string][]string)
	for _, edge := range edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge.To)
	}

	// Test root resource -> module dependency (module depends on root resource)
	assert.Contains(t, edgeMap, "aws_vpc_main", "Root resource should be a dependency")
	assert.Contains(t, edgeMap["aws_vpc_main"], "module_network_aws_subnet_public", "Root resource should point to dependent module")

	// Test module -> module dependency (app module depends on network module)
	assert.Contains(t, edgeMap, "module_network_aws_subnet_public", "Network module should be a dependency")
	assert.Contains(t, edgeMap["module_network_aws_subnet_public"], "module_app_aws_instance_web", "Network module should point to dependent app module")

	// Test module data source dependency (module data source depends on module resource)
	assert.Contains(t, edgeMap, "module_network_aws_subnet_public", "Network module should be a dependency")
	assert.Contains(t, edgeMap["module_network_aws_subnet_public"], "module_app_data_aws_subnet_selected", "Network module should point to dependent data source")

	// Test root resource -> output dependency (output depends on root resource)
	assert.Contains(t, edgeMap, "aws_vpc_main", "Root resource should be a dependency")
	assert.Contains(t, edgeMap["aws_vpc_main"], "output_vpc_id", "Root resource should point to dependent output")

	// Test module -> output dependency (output depends on module resource)
	assert.Contains(t, edgeMap, "module_network_aws_subnet_public", "Network module should be a dependency")
	assert.Contains(t, edgeMap["module_network_aws_subnet_public"], "output_subnet_id", "Network module should point to dependent output")
}

func TestAnalyzeDependencies_ExplicitDependsOn(t *testing.T) {
	// Test explicit depends_on dependencies
	plan := &core.TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []core.ResourceChange{
			{
				Address:       "aws_vpc.main",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_vpc",
				Name:          "main",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"cidr_block": "10.0.0.0/16"},
				},
			},
			{
				Address:       "aws_subnet.public",
				ModuleAddress: "",
				Mode:          "managed",
				Type:          "aws_subnet",
				Name:          "public",
				Change: core.Change{
					Actions: []string{"create"},
					Before:  nil,
					After:   map[string]interface{}{"cidr_block": "10.0.1.0/24"},
				},
			},
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
		},
		OutputChanges: make(map[string]core.OutputChange),
		Configuration: core.Configuration{
			RootModule: core.RootModule{
				Resources: []core.ConfigurationResource{
					{
						Address: "aws_vpc.main",
						Mode:    "managed",
						Type:    "aws_vpc",
						Name:    "main",
						Expressions: map[string]interface{}{
							"cidr_block": map[string]interface{}{
								"constant_value": "10.0.0.0/16",
							},
						},
					},
					{
						Address: "aws_subnet.public",
						Mode:    "managed",
						Type:    "aws_subnet",
						Name:    "public",
						Expressions: map[string]interface{}{
							"vpc_id": map[string]interface{}{
								"references": []interface{}{
									"aws_vpc.main.id",
								},
							},
						},
					},
					{
						Address: "aws_instance.web",
						Mode:    "managed",
						Type:    "aws_instance",
						Name:    "web",
						DependsOn: []string{
							"aws_vpc.main",
							"aws_subnet.public",
						},
						Expressions: map[string]interface{}{
							"subnet_id": map[string]interface{}{
								"references": []interface{}{
									"aws_subnet.public.id",
								},
							},
						},
					},
				},
				ModuleCalls: make(map[string]core.ModuleCall),
				Outputs:     make(map[string]core.OutputConfig),
				Variables:   make(map[string]core.VariableConfig),
				Locals:      make(map[string]core.LocalConfig),
			},
		},
		Variables:  make(map[string]core.Variable),
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
	edges, err := builder.analyzeDependencies(plan, opts)

	assert.NoError(t, err)
	assert.NotEmpty(t, edges, "Should have dependency edges")

	// Convert edges to a map for easier testing
	edgeMap := make(map[string][]string)
	for _, edge := range edges {
		edgeMap[edge.From] = append(edgeMap[edge.From], edge.To)
	}

	// Test implicit dependency (subnet depends on vpc)
	assert.Contains(t, edgeMap, "aws_vpc_main", "VPC should be a dependency")
	assert.Contains(t, edgeMap["aws_vpc_main"], "aws_subnet_public", "VPC should point to dependent subnet")

	// Test explicit depends_on dependency (instance depends on vpc)
	assert.Contains(t, edgeMap, "aws_vpc_main", "VPC should be a dependency")
	assert.Contains(t, edgeMap["aws_vpc_main"], "aws_instance_web", "VPC should point to dependent instance via depends_on")

	// Test explicit depends_on dependency (instance depends on subnet)
	assert.Contains(t, edgeMap, "aws_subnet_public", "Subnet should be a dependency")
	assert.Contains(t, edgeMap["aws_subnet_public"], "aws_instance_web", "Subnet should point to dependent instance via depends_on")

	// Test implicit dependency (instance depends on subnet via reference)
	// This should already be covered by the depends_on test above, but let's verify
	assert.Contains(t, edgeMap["aws_subnet_public"], "aws_instance_web", "Subnet should point to dependent instance via reference")
}
