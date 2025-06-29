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
