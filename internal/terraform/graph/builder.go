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
	"sort"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
	"github.com/yu/terraform-ops/internal/ir"
)

// Builder projects the normalized dependency graph into the compatibility graph
// renderer model. Dependency discovery is performed once by the source
// normalizer and is never reimplemented here.
type Builder struct{}

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) BuildGraph(changeSet *ir.ChangeSet, opts core.GraphOptions) (*core.GraphData, error) {
	if changeSet == nil {
		return nil, &core.ValidationError{Field: "change_set", Message: "must not be nil"}
	}

	resources := make(map[string]ir.ResourceChange, len(changeSet.Resources))
	for _, resource := range changeSet.Resources {
		resources[string(resource.Address)] = resource
	}
	outputs := make(map[string]ir.OutputChange, len(changeSet.Outputs))
	for _, output := range changeSet.Outputs {
		outputs["output."+output.Name] = output
	}

	graphData := &core.GraphData{}
	included := make(map[ir.NodeID]string, len(changeSet.Graph.Nodes))
	for _, node := range changeSet.Graph.Nodes {
		view, ok := projectNode(node, resources, outputs, opts)
		if !ok {
			continue
		}
		graphData.Nodes = append(graphData.Nodes, view)
		included[node.ID] = view.ID
	}

	// Multiple pieces of normalized evidence can connect the same pair of
	// resources. Diagram renderers need one visual edge, while ChangeSet retains
	// every evidence kind/confidence separately.
	edgeSet := make(map[string]core.GraphEdge)
	for _, edge := range changeSet.Graph.Edges {
		from, fromOK := included[edge.From]
		to, toOK := included[edge.To]
		if !fromOK || !toOK || from == to {
			continue
		}
		key := from + "\x00" + to
		edgeSet[key] = core.GraphEdge{From: from, To: to}
	}
	for _, edge := range edgeSet {
		graphData.Edges = append(graphData.Edges, edge)
	}

	sort.Slice(graphData.Nodes, func(i, j int) bool { return graphData.Nodes[i].Address < graphData.Nodes[j].Address })
	sort.Slice(graphData.Edges, func(i, j int) bool {
		if graphData.Edges[i].From != graphData.Edges[j].From {
			return graphData.Edges[i].From < graphData.Edges[j].From
		}
		return graphData.Edges[i].To < graphData.Edges[j].To
	})
	return graphData, nil
}

func projectNode(
	node ir.Node,
	resources map[string]ir.ResourceChange,
	outputs map[string]ir.OutputChange,
	opts core.GraphOptions,
) (core.GraphNode, bool) {
	address := string(node.Address)
	kind := node.Kind
	if kind == "" {
		// Compatibility for hand-built ChangeSet fixtures created before typed
		// graph nodes were introduced.
		switch {
		case strings.HasPrefix(address, "output."):
			kind = ir.NodeKindOutput
		case strings.HasPrefix(address, "var."):
			kind = ir.NodeKindVariable
		default:
			if resource, ok := resources[address]; ok && resource.Mode == ir.ResourceModeData {
				kind = ir.NodeKindData
			} else {
				kind = ir.NodeKindResource
			}
		}
	}

	view := core.GraphNode{ID: sanitizeID(string(node.ID)), Address: address}
	switch kind {
	case ir.NodeKindResource, ir.NodeKindData:
		resource, ok := resources[address]
		if !ok {
			return core.GraphNode{}, false
		}
		if kind == ir.NodeKindData && opts.NoDataSources {
			return core.GraphNode{}, false
		}
		module := ""
		if resource.ModuleAddress != nil {
			module = string(*resource.ModuleAddress)
		}
		if opts.NoModules && module != "" {
			return core.GraphNode{}, false
		}
		view.Type = resource.Type
		view.Name = resource.Name
		view.Module = module
		view.Provider = extractProviderFromType(resource.Type)
		view.Actions = append([]string(nil), resource.Action.Raw...)
		view.Sensitive = len(resource.SensitivePaths) > 0
	case ir.NodeKindOutput:
		if opts.NoOutputs {
			return core.GraphNode{}, false
		}
		output, ok := outputs[address]
		if !ok {
			return core.GraphNode{}, false
		}
		view.Type = string(core.NodeTypeOutput)
		view.Name = output.Name
		view.Actions = append([]string(nil), output.Action.Raw...)
		view.Sensitive = len(output.SensitivePaths) > 0
	case ir.NodeKindVariable:
		if opts.NoVariables {
			return core.GraphNode{}, false
		}
		view.Type = string(core.NodeTypeVariable)
		view.Name = strings.TrimPrefix(address, "var.")
		view.Actions = []string{"no-op"}
	default:
		return core.GraphNode{}, false
	}
	return view, true
}

func sanitizeID(id string) string {
	replacements := map[string]string{
		".": "_",
		"-": "_",
		"[": "_",
		"]": "_",
		"(": "_",
		")": "_",
		" ": "_",
	}
	result := id
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}
	return result
}

func isResourceType(s string) bool {
	return strings.Contains(s, "_")
}

func extractProviderFromType(resourceType string) string {
	provider, _, ok := strings.Cut(resourceType, "_")
	if !ok {
		return ""
	}
	return provider
}
