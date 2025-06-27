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

package plan_graph

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParsePlanFile(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name: "valid plan file",
			jsonData: `{
				"format_version": "1.0",
				"resource_changes": [
					{
						"address": "aws_instance.web",
						"module_address": "",
						"mode": "managed",
						"type": "aws_instance",
						"name": "web",
						"change": {
							"actions": ["create"],
							"before_sensitive": {},
							"after_sensitive": {}
						}
					}
				]
			}`,
			wantErr: false,
		},
		{
			name:     "invalid JSON",
			jsonData: `{ invalid json }`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpfile, err := os.CreateTemp("", "plan-*.json")
			require.NoError(t, err)
			defer func() {
				if removeErr := os.Remove(tmpfile.Name()); removeErr != nil {
					t.Logf("Failed to remove temporary file: %v", removeErr)
				}
			}()

			// Write test data
			_, err = tmpfile.WriteString(tt.jsonData)
			require.NoError(t, err)
			if err := tmpfile.Close(); err != nil {
				t.Fatalf("Failed to close temporary file: %v", err)
			}

			// Test parsing
			plan, err := ParsePlanFile(tmpfile.Name())
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, plan)
			assert.Equal(t, "1.0", plan.FormatVersion)
			assert.Len(t, plan.ResourceChanges, 1)
		})
	}
}

func TestBuildGraphData(t *testing.T) {
	plan := &TerraformPlan{
		FormatVersion: "1.0",
		ResourceChanges: []ResourceChange{
			{
				Address:       "aws_instance.web",
				ModuleAddress: "",
				Type:          "aws_instance",
				Name:          "web",
				Change: Change{
					Actions:        []string{"create"},
					AfterSensitive: map[string]interface{}{},
				},
			},
			{
				Address:       "module.db.aws_instance.database",
				ModuleAddress: "module.db",
				Type:          "aws_instance",
				Name:          "database",
				Change: Change{
					Actions: []string{"update"},
					AfterSensitive: map[string]interface{}{
						"password": true,
					},
				},
			},
		},
	}

	opts := Options{
		Format: FormatGraphviz,
	}

	graphData, err := BuildGraphData(plan, opts)
	require.NoError(t, err)

	assert.Len(t, graphData.Nodes, 2)
	assert.Len(t, graphData.Edges, 0) // No edges in this simple case

	// Check first node
	assert.Equal(t, "aws_instance_web", graphData.Nodes[0].ID)
	assert.Equal(t, "aws_instance.web", graphData.Nodes[0].Address)
	assert.Equal(t, "aws_instance", graphData.Nodes[0].Type)
	assert.Equal(t, "web", graphData.Nodes[0].Name)
	assert.Equal(t, "", graphData.Nodes[0].Module)
	assert.Equal(t, []string{"create"}, graphData.Nodes[0].Actions)
	assert.False(t, graphData.Nodes[0].Sensitive)

	// Check second node
	assert.Equal(t, "module_db_aws_instance_database", graphData.Nodes[1].ID)
	assert.Equal(t, "module.db.aws_instance.database", graphData.Nodes[1].Address)
	assert.Equal(t, "module.db", graphData.Nodes[1].Module)
	assert.Equal(t, []string{"update"}, graphData.Nodes[1].Actions)
	assert.True(t, graphData.Nodes[1].Sensitive)
}

func TestGetActionType(t *testing.T) {
	tests := []struct {
		name    string
		actions []string
		want    ActionType
	}{
		{
			name:    "create action",
			actions: []string{"create"},
			want:    ActionCreate,
		},
		{
			name:    "update action",
			actions: []string{"update"},
			want:    ActionUpdate,
		},
		{
			name:    "delete action",
			actions: []string{"delete"},
			want:    ActionDelete,
		},
		{
			name:    "replace action",
			actions: []string{"delete", "create"},
			want:    ActionReplace,
		},
		{
			name:    "no-op action",
			actions: []string{"no-op"},
			want:    ActionNoOp,
		},
		{
			name:    "empty actions",
			actions: []string{},
			want:    ActionNoOp,
		},
		{
			name:    "unknown action",
			actions: []string{"unknown"},
			want:    ActionNoOp,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getActionType(tt.actions)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetActionColor(t *testing.T) {
	tests := []struct {
		name       string
		actionType ActionType
		want       string
	}{
		{
			name:       "create color",
			actionType: ActionCreate,
			want:       "lightgreen",
		},
		{
			name:       "update color",
			actionType: ActionUpdate,
			want:       "yellow",
		},
		{
			name:       "delete color",
			actionType: ActionDelete,
			want:       "lightcoral",
		},
		{
			name:       "replace color",
			actionType: ActionReplace,
			want:       "orange",
		},
		{
			name:       "no-op color",
			actionType: ActionNoOp,
			want:       "lightgrey",
		},
		{
			name:       "unknown color",
			actionType: "unknown",
			want:       "white",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getActionColor(tt.actionType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want string
	}{
		{
			name: "simple resource",
			id:   "aws_instance.web",
			want: "aws_instance_web",
		},
		{
			name: "module resource",
			id:   "module.db.aws_instance.database",
			want: "module_db_aws_instance_database",
		},
		{
			name: "resource with brackets",
			id:   "aws_instance.web[0]",
			want: "aws_instance_web_0_",
		},
		{
			name: "resource with spaces",
			id:   "aws instance web",
			want: "aws_instance_web",
		},
		{
			name: "resource with parentheses",
			id:   "aws_instance.web(test)",
			want: "aws_instance_web_test_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeID(tt.id)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGenerateGraph(t *testing.T) {
	graphData := &GraphData{
		Nodes: []GraphNode{
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
				ID:        "module_db_aws_instance_database",
				Address:   "module.db.aws_instance.database",
				Type:      "aws_instance",
				Name:      "database",
				Module:    "module.db",
				Actions:   []string{"update"},
				Sensitive: true,
			},
		},
		Edges: []GraphEdge{},
	}

	tests := []struct {
		name    string
		format  GraphFormat
		wantErr bool
	}{
		{
			name:    "graphviz format",
			format:  FormatGraphviz,
			wantErr: false,
		},
		{
			name:    "mermaid format",
			format:  FormatMermaid,
			wantErr: false,
		},
		{
			name:    "plantuml format",
			format:  FormatPlantUML,
			wantErr: false,
		},
		{
			name:    "unsupported format",
			format:  "unsupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{Format: tt.format}
			output, err := GenerateGraph(graphData, opts)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, output)

			// Check that output contains expected elements
			switch tt.format {
			case FormatGraphviz:
				assert.Contains(t, output, "digraph terraform_plan")
				assert.Contains(t, output, "aws_instance_web")
				assert.Contains(t, output, "module_db_aws_instance_database")
			case FormatMermaid:
				assert.Contains(t, output, "graph TB")
				assert.Contains(t, output, "aws_instance_web")
				assert.Contains(t, output, "module_db_aws_instance_database")
			case FormatPlantUML:
				assert.Contains(t, output, "@startuml")
				assert.Contains(t, output, "aws_instance_web")
				assert.Contains(t, output, "module_db_aws_instance_database")
			}
		})
	}
}

func TestGroupNodesByModule(t *testing.T) {
	nodes := []GraphNode{
		{
			ID:     "aws_instance_web",
			Module: "",
		},
		{
			ID:     "aws_security_group_web",
			Module: "",
		},
		{
			ID:     "module_db_aws_instance_database",
			Module: "module.db",
		},
		{
			ID:     "module_cache_aws_elasticache_cluster",
			Module: "module.cache",
		},
	}

	groups := groupNodesByModule(nodes)

	assert.Len(t, groups, 3)
	assert.Len(t, groups["root"], 2)
	assert.Len(t, groups["module.db"], 1)
	assert.Len(t, groups["module.cache"], 1)
}

func TestHasSensitiveValues(t *testing.T) {
	tests := []struct {
		name      string
		sensitive map[string]interface{}
		want      bool
	}{
		{
			name:      "no sensitive values",
			sensitive: map[string]interface{}{},
			want:      false,
		},
		{
			name: "has sensitive values",
			sensitive: map[string]interface{}{
				"password": true,
			},
			want: true,
		},
		{
			name: "multiple sensitive values",
			sensitive: map[string]interface{}{
				"password": true,
				"secret":   true,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasSensitiveValues(tt.sensitive)
			assert.Equal(t, tt.want, got)
		})
	}
}
