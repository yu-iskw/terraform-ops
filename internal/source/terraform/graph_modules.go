// Copyright 2026 yu-iskw
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

package terraform

import (
	"sort"
	"strings"

	"github.com/yu/terraform-ops/internal/ir"
)

type graphSource struct {
	id         ir.NodeID
	confidence ir.EvidenceConfidence
}

type configResourceContext struct {
	modulePath string
	address    string
	resource   ConfigResource
}

func buildDependencyGraphWithModules(configuration Configuration, resources []ir.ResourceChange) ir.DependencyGraph {
	// Preserve the existing resource-level graph as the baseline, then augment
	// it with module-aware edges. This keeps the established depends_on and
	// expression-reference behavior while adding module input/output traversal.
	graph := buildDependencyGraph(configuration, resources)
	changed := make(map[string]ir.NodeID, len(resources))
	for _, resource := range resources {
		id := ir.NodeID(resource.Address)
		changed[string(resource.Address)] = id
	}

	contexts := flattenConfigResourceContexts(configuration.RootModule)
	baseToChanges := make(map[string][]ir.NodeID)
	for _, context := range contexts {
		for address, id := range changed {
			if stripAddressIndexes(address) == stripAddressIndexes(context.address) {
				baseToChanges[context.address] = append(baseToChanges[context.address], id)
			}
		}
		baseToChanges[context.address] = uniqueNodeIDs(baseToChanges[context.address])
	}

	edges := make(map[string]ir.Edge, len(graph.Edges))
	for _, edge := range graph.Edges {
		addEdge(edges, edge.From, edge.To, edge.Kind, edge.Confidence)
	}

	// Evaluate resource-level references with module context. Configuration
	// expressions inside child modules use addresses relative to that module.
	for _, context := range contexts {
		targets := baseToChanges[context.address]
		if len(targets) == 0 {
			continue
		}
		for _, dependency := range context.resource.DependsOn {
			if strings.HasPrefix(stripAddressIndexes(dependency), "module.") {
				childPath := referencedChildModulePath(context.modulePath, dependency)
				for _, source := range changedSourcesUnderModule(childPath, changed, ir.ConfidenceExact) {
					addSourcesToTargets(edges, []graphSource{source}, targets, ir.EdgeExplicitDependsOn)
				}
				continue
			}
			reference := qualifyResourceReference(context.modulePath, dependency)
			for _, source := range resolveReference(reference, changed, baseToChanges) {
				for _, target := range targets {
					addEdge(edges, source, target, ir.EdgeExplicitDependsOn, ir.ConfidenceExact)
				}
			}
		}
		for _, rawExpression := range context.resource.Expressions {
			for _, reference := range extractReferences(rawExpression) {
				normalized := stripAddressIndexes(reference)
				if strings.HasPrefix(normalized, "var.") || strings.HasPrefix(normalized, "module.") {
					continue
				}
				reference = qualifyResourceReference(context.modulePath, reference)
				for _, source := range resolveReference(reference, changed, baseToChanges) {
					for _, target := range targets {
						addEdge(edges, source, target, ir.EdgeExpressionRef, ir.ConfidenceExact)
					}
				}
			}
		}
	}

	addModuleBoundaryEdges(
		configuration.RootModule,
		"",
		nil,
		changed,
		baseToChanges,
		edges,
	)

	graph.Edges = graph.Edges[:0]
	for _, edge := range edges {
		graph.Edges = append(graph.Edges, edge)
	}
	return graph
}

func addModuleBoundaryEdges(
	module Module,
	modulePath string,
	inputs map[string][]graphSource,
	changed map[string]ir.NodeID,
	baseToChanges map[string][]ir.NodeID,
	edges map[string]ir.Edge,
) map[string][]graphSource {
	childOutputs := make(map[string][]graphSource)

	names := make([]string, 0, len(module.ModuleCalls))
	for name := range module.ModuleCalls {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		call := module.ModuleCalls[name]
		childPath := appendModulePath(modulePath, name)
		childInputs := make(map[string][]graphSource, len(call.Expressions))
		for inputName, rawExpression := range call.Expressions {
			var sources []graphSource
			for _, reference := range extractReferences(rawExpression) {
				sources = append(sources, resolveBoundaryReference(
					reference,
					modulePath,
					inputs,
					childOutputs,
					changed,
					baseToChanges,
				)...)
			}
			childInputs[inputName] = uniqueGraphSources(sources)
		}
		if call.Module == nil {
			continue
		}
		for outputRef, sources := range addModuleBoundaryEdges(
			*call.Module,
			childPath,
			childInputs,
			changed,
			baseToChanges,
			edges,
		) {
			childOutputs[outputRef] = uniqueGraphSources(append(childOutputs[outputRef], sources...))
		}
	}

	for _, resource := range module.Resources {
		address := qualifyConfigResourceAddress(modulePath, resource.Address)
		targets := baseToChanges[address]
		if len(targets) == 0 {
			continue
		}
		for _, rawExpression := range resource.Expressions {
			for _, reference := range extractReferences(rawExpression) {
				normalized := stripAddressIndexes(reference)
				switch {
				case strings.HasPrefix(normalized, "var."):
					name := variableReferenceName(normalized)
					addSourcesToTargets(edges, inputs[name], targets, ir.EdgeModuleInput)
				case strings.HasPrefix(normalized, "module."):
					sources := resolveBoundaryReference(
						reference,
						modulePath,
						inputs,
						childOutputs,
						changed,
						baseToChanges,
					)
					addSourcesToTargets(edges, sources, targets, ir.EdgeModuleOutput)
				}
			}
		}
	}

	outputs := make(map[string][]graphSource, len(module.Outputs))
	outputNames := make([]string, 0, len(module.Outputs))
	for name := range module.Outputs {
		outputNames = append(outputNames, name)
	}
	sort.Strings(outputNames)
	for _, name := range outputNames {
		rawOutput := module.Outputs[name]
		var sources []graphSource
		for _, reference := range extractReferences(rawOutput.Expression) {
			sources = append(sources, resolveBoundaryReference(
				reference,
				modulePath,
				inputs,
				childOutputs,
				changed,
				baseToChanges,
			)...)
		}
		outputs[moduleOutputReference(modulePath, name)] = uniqueGraphSources(sources)
	}
	return outputs
}

func resolveBoundaryReference(
	reference string,
	modulePath string,
	inputs map[string][]graphSource,
	childOutputs map[string][]graphSource,
	changed map[string]ir.NodeID,
	baseToChanges map[string][]ir.NodeID,
) []graphSource {
	normalized := stripAddressIndexes(reference)
	if strings.HasPrefix(normalized, "var.") {
		return append([]graphSource(nil), inputs[variableReferenceName(normalized)]...)
	}
	if strings.HasPrefix(normalized, "module.") {
		absolute := absoluteModuleReference(modulePath, normalized)
		if sources := resolveModuleOutputReference(absolute, childOutputs); len(sources) > 0 {
			return sources
		}
		// If Terraform's expression representation does not let us attribute the
		// output to a specific changed resource, prefer conservative propagation
		// over understating blast radius.
		childPath := referencedChildModulePath(modulePath, normalized)
		return changedSourcesUnderModule(childPath, changed, ir.ConfidenceHeuristic)
	}

	qualified := qualifyResourceReference(modulePath, reference)
	ids := resolveReference(qualified, changed, baseToChanges)
	sources := make([]graphSource, 0, len(ids))
	for _, id := range ids {
		sources = append(sources, graphSource{id: id, confidence: ir.ConfidenceExact})
	}
	return sources
}

func resolveModuleOutputReference(reference string, outputs map[string][]graphSource) []graphSource {
	var best string
	for outputRef := range outputs {
		if referenceHasAddressPrefix(reference, outputRef) && len(outputRef) > len(best) {
			best = outputRef
		}
	}
	if best == "" {
		return nil
	}
	return append([]graphSource(nil), outputs[best]...)
}

func addSourcesToTargets(
	edges map[string]ir.Edge,
	sources []graphSource,
	targets []ir.NodeID,
	kind ir.EdgeKind,
) {
	for _, source := range uniqueGraphSources(sources) {
		for _, target := range targets {
			addEdge(edges, source.id, target, kind, source.confidence)
		}
	}
}

func flattenConfigResourceContexts(root Module) []configResourceContext {
	var out []configResourceContext
	var visit func(Module, string)
	visit = func(module Module, modulePath string) {
		for _, resource := range module.Resources {
			out = append(out, configResourceContext{
				modulePath: modulePath,
				address:    qualifyConfigResourceAddress(modulePath, resource.Address),
				resource:   resource,
			})
		}
		names := make([]string, 0, len(module.ModuleCalls))
		for name := range module.ModuleCalls {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			call := module.ModuleCalls[name]
			if call.Module != nil {
				visit(*call.Module, appendModulePath(modulePath, name))
			}
		}
	}
	visit(root, "")
	return out
}

func qualifyConfigResourceAddress(modulePath, address string) string {
	if modulePath == "" {
		return address
	}
	normalized := stripAddressIndexes(address)
	if strings.HasPrefix(normalized, modulePath+".") {
		return address
	}
	return modulePath + "." + address
}

func qualifyResourceReference(modulePath, reference string) string {
	if modulePath == "" || !isRelativeResourceReference(reference) {
		return reference
	}
	normalized := stripAddressIndexes(reference)
	if strings.HasPrefix(normalized, modulePath+".") {
		return reference
	}
	return modulePath + "." + reference
}

func isRelativeResourceReference(reference string) bool {
	normalized := stripAddressIndexes(reference)
	for _, prefix := range []string{
		"var.", "local.", "module.", "path.", "terraform.", "count.", "each.",
	} {
		if strings.HasPrefix(normalized, prefix) {
			return false
		}
	}
	return normalized != "self"
}

func appendModulePath(parent, name string) string {
	child := "module." + name
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func absoluteModuleReference(modulePath, reference string) string {
	if modulePath == "" {
		return reference
	}
	if strings.HasPrefix(reference, modulePath+".") {
		return reference
	}
	return modulePath + "." + reference
}

func referencedChildModulePath(modulePath, reference string) string {
	normalized := stripAddressIndexes(reference)
	if !strings.HasPrefix(normalized, "module.") {
		return ""
	}
	rest := strings.TrimPrefix(normalized, "module.")
	name, _, _ := strings.Cut(rest, ".")
	if name == "" {
		return ""
	}
	return appendModulePath(modulePath, name)
}

func moduleOutputReference(modulePath, outputName string) string {
	if modulePath == "" {
		return outputName
	}
	return modulePath + "." + outputName
}

func variableReferenceName(reference string) string {
	rest := strings.TrimPrefix(reference, "var.")
	name, _, _ := strings.Cut(rest, ".")
	return name
}

func changedSourcesUnderModule(
	modulePath string,
	changed map[string]ir.NodeID,
	confidence ir.EvidenceConfidence,
) []graphSource {
	if modulePath == "" {
		return nil
	}
	var sources []graphSource
	for address, id := range changed {
		normalized := stripAddressIndexes(address)
		if strings.HasPrefix(normalized, modulePath+".") {
			sources = append(sources, graphSource{id: id, confidence: confidence})
		}
	}
	return uniqueGraphSources(sources)
}

func stripAddressIndexes(address string) string {
	var out strings.Builder
	out.Grow(len(address))

	inIndex := false
	inString := false
	escaped := false
	for i := 0; i < len(address); i++ {
		ch := address[i]
		if !inIndex {
			if ch == '[' {
				inIndex = true
				continue
			}
			out.WriteByte(ch)
			continue
		}

		if escaped {
			escaped = false
			continue
		}
		if inString && ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if ch == ']' && !inString {
			inIndex = false
		}
	}
	return out.String()
}

func uniqueGraphSources(sources []graphSource) []graphSource {
	byID := make(map[ir.NodeID]ir.EvidenceConfidence, len(sources))
	for _, source := range sources {
		if source.id == "" {
			continue
		}
		if current, ok := byID[source.id]; !ok || confidenceRank(source.confidence) > confidenceRank(current) {
			byID[source.id] = source.confidence
		}
	}
	ids := make([]ir.NodeID, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	out := make([]graphSource, 0, len(ids))
	for _, id := range ids {
		out = append(out, graphSource{id: id, confidence: byID[id]})
	}
	return out
}

func uniqueNodeIDs(ids []ir.NodeID) []ir.NodeID {
	set := make(map[ir.NodeID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	out := make([]ir.NodeID, 0, len(set))
	for id := range set {
		out = append(out, id)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func confidenceRank(confidence ir.EvidenceConfidence) int {
	switch confidence {
	case ir.ConfidenceExact:
		return 4
	case ir.ConfidenceStrong:
		return 3
	case ir.ConfidenceHeuristic:
		return 2
	case ir.ConfidenceUnknown:
		return 1
	default:
		return 0
	}
}
