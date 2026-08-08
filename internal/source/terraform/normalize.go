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
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/yu/terraform-ops/internal/ir"
)

func Normalize(plan *Plan, engine ir.Engine, mode ir.RedactionMode) (*ir.ChangeSet, error) {
	if plan == nil {
		return nil, fmt.Errorf("plan is nil")
	}
	if engine == "" {
		engine = ir.EngineUnknown
	}
	if mode == "" {
		mode = ir.RedactionStandard
	}

	out := &ir.ChangeSet{
		SchemaVersion: ir.SchemaVersion,
		Source: ir.SourceMetadata{
			Engine:            engine,
			EngineVersion:     plan.TerraformVersion,
			PlanFormatVersion: plan.FormatVersion,
		},
		Plan: ir.PlanMetadata{
			Applyable: plan.Applyable,
			Complete:  planComplete(plan),
			Errored:   plan.Errored,
		},
		Redaction: ir.RedactionSummary{
			Mode:                  mode,
			VariableValuesRemoved: len(plan.Variables),
		},
	}

	for _, raw := range plan.ResourceChanges {
		change, count, strictCount, err := normalizeResourceChange(raw, mode)
		if err != nil {
			return nil, fmt.Errorf("normalize resource %q: %w", raw.Address, err)
		}
		out.Resources = append(out.Resources, change)
		out.Redaction.TerraformSensitivePaths += count
		out.Redaction.StrictValuesRemoved += strictCount
	}

	for name, raw := range plan.OutputChanges {
		output, count, strictCount, err := normalizeOutputChange(name, raw, mode)
		if err != nil {
			return nil, fmt.Errorf("normalize output %q: %w", name, err)
		}
		out.Outputs = append(out.Outputs, output)
		out.Redaction.TerraformSensitivePaths += count
		out.Redaction.StrictValuesRemoved += strictCount
	}

	for _, check := range plan.Checks {
		out.Checks = append(out.Checks, normalizeCheck(check, mode)...)
	}

	for _, relevant := range plan.RelevantAttributes {
		path, err := parsePathRaw(relevant.Attribute)
		if err != nil {
			return nil, fmt.Errorf("normalize relevant attribute %q: %w", relevant.Resource, err)
		}
		out.Relevant = append(out.Relevant, ir.RelevantAttribute{
			Resource: ir.Address(relevant.Resource),
			Path:     path,
		})
	}

	relevantByResource := make(map[ir.Address][]ir.RelevantAttribute)
	for _, relevant := range out.Relevant {
		relevantByResource[relevant.Resource] = append(relevantByResource[relevant.Resource], relevant)
	}
	for _, raw := range plan.ResourceDrift {
		change, count, strictCount, err := normalizeResourceChange(raw, mode)
		if err != nil {
			return nil, fmt.Errorf("normalize drift resource %q: %w", raw.Address, err)
		}
		out.Redaction.TerraformSensitivePaths += count
		out.Redaction.StrictValuesRemoved += strictCount
		out.Drift = append(out.Drift, ir.DriftChange{
			Resource: change,
			Relevant: relevantByResource[change.Address],
		})
	}

	out.Graph = buildDependencyGraphWithModules(
		plan.Configuration,
		out.Resources,
		out.Outputs,
		variableNames(plan.Variables),
	)
	out.Sort()
	return out, nil
}

func normalizeResourceChange(raw ResourceChange, mode ir.RedactionMode) (ir.ResourceChange, int, int, error) {
	before, beforePaths, beforeStrict, err := sanitizeValue(raw.Change.Before, raw.Change.BeforeSensitive, mode)
	if err != nil {
		return ir.ResourceChange{}, 0, 0, fmt.Errorf("sanitize before value: %w", err)
	}
	after, afterPaths, afterStrict, err := sanitizeValue(raw.Change.After, raw.Change.AfterSensitive, mode)
	if err != nil {
		return ir.ResourceChange{}, 0, 0, fmt.Errorf("sanitize after value: %w", err)
	}
	unknown, err := collectMaskPaths(raw.Change.AfterUnknown)
	if err != nil {
		return ir.ResourceChange{}, 0, 0, fmt.Errorf("collect unknown paths: %w", err)
	}
	replacePaths := make([]ir.AttributePath, 0, len(raw.Change.ReplacePaths))
	for _, replacePath := range raw.Change.ReplacePaths {
		path, err := parsePathRaw(replacePath)
		if err != nil {
			return ir.ResourceChange{}, 0, 0, fmt.Errorf("parse replacement path: %w", err)
		}
		replacePaths = append(replacePaths, path)
	}

	change := ir.ResourceChange{
		Address:        ir.Address(raw.Address),
		Mode:           ir.ResourceMode(raw.Mode),
		Type:           raw.Type,
		Name:           raw.Name,
		DeposedKey:     raw.Deposed,
		Action:         ir.NormalizeAction(raw.Change.Actions),
		ActionReason:   raw.ActionReason,
		ReplacePaths:   replacePaths,
		Before:         before,
		After:          after,
		UnknownPaths:   uniquePaths(unknown),
		SensitivePaths: uniquePaths(append(beforePaths, afterPaths...)),
	}
	if raw.PreviousAddress != "" {
		addr := ir.Address(raw.PreviousAddress)
		change.PreviousAddress = &addr
	}
	if raw.ModuleAddress != "" {
		addr := ir.Address(raw.ModuleAddress)
		change.ModuleAddress = &addr
	}
	if len(raw.Index) > 0 {
		index, err := decodeJSON(raw.Index)
		if err != nil {
			return ir.ResourceChange{}, 0, 0, fmt.Errorf("decode instance index: %w", err)
		}
		change.Index = fmt.Sprint(index)
	}
	if raw.Change.Importing != nil {
		change.Import = &ir.ImportInfo{ID: raw.Change.Importing.ID, Unknown: raw.Change.Importing.Unknown}
	}
	return change, len(change.SensitivePaths), beforeStrict + afterStrict, nil
}

func normalizeOutputChange(name string, raw OutputChange, mode ir.RedactionMode) (ir.OutputChange, int, int, error) {
	before, beforePaths, beforeStrict, err := sanitizeValue(raw.Change.Before, raw.Change.BeforeSensitive, mode)
	if err != nil {
		return ir.OutputChange{}, 0, 0, fmt.Errorf("sanitize before value: %w", err)
	}
	after, afterPaths, afterStrict, err := sanitizeValue(raw.Change.After, raw.Change.AfterSensitive, mode)
	if err != nil {
		return ir.OutputChange{}, 0, 0, fmt.Errorf("sanitize after value: %w", err)
	}
	unknown, err := collectMaskPaths(raw.Change.AfterUnknown)
	if err != nil {
		return ir.OutputChange{}, 0, 0, err
	}
	sensitive := uniquePaths(append(beforePaths, afterPaths...))
	return ir.OutputChange{
		Name:           name,
		Action:         ir.NormalizeAction(raw.Change.Actions),
		Before:         before,
		After:          after,
		UnknownPaths:   uniquePaths(unknown),
		SensitivePaths: sensitive,
	}, len(sensitive), beforeStrict + afterStrict, nil
}

func normalizeCheck(raw Check, mode ir.RedactionMode) []ir.CheckResult {
	if len(raw.Instances) == 0 {
		return []ir.CheckResult{{
			Address: raw.Address.ToDisplay,
			Kind:    raw.Address.Kind,
			Status:  ir.CheckStatus(raw.Status),
		}}
	}
	results := make([]ir.CheckResult, 0, len(raw.Instances))
	for _, instance := range raw.Instances {
		result := ir.CheckResult{
			Address: instance.Address.ToDisplay,
			Kind:    raw.Address.Kind,
			Status:  ir.CheckStatus(instance.Status),
		}
		if result.Address == "" {
			result.Address = raw.Address.ToDisplay
		}
		if mode != ir.RedactionStrict {
			for _, problem := range instance.Problems {
				result.Problems = append(result.Problems, ir.CheckProblem{Message: problem.Message})
			}
		}
		results = append(results, result)
	}
	return results
}

func sanitizeValue(valueRaw, maskRaw json.RawMessage, mode ir.RedactionMode) (ir.SafeValue, []ir.AttributePath, int, error) {
	paths, err := collectMaskPaths(maskRaw)
	if err != nil {
		return ir.SafeValue{}, nil, 0, err
	}
	if mode == ir.RedactionStrict {
		if len(valueRaw) == 0 || string(valueRaw) == "null" {
			return ir.SafeValue{}, paths, 0, nil
		}
		return ir.SafeValue{Redacted: true}, paths, 1, nil
	}
	value, err := decodeJSON(valueRaw)
	if err != nil {
		return ir.SafeValue{}, nil, 0, err
	}
	mask, err := decodeJSON(maskRaw)
	if err != nil {
		return ir.SafeValue{}, nil, 0, err
	}
	redacted, fullyRedacted := applyMask(value, mask)
	return ir.SafeValue{Value: redacted, Redacted: fullyRedacted}, paths, 0, nil
}

func applyMask(value, mask any) (any, bool) {
	if sensitive, ok := mask.(bool); ok && sensitive {
		return nil, true
	}
	if mask == nil {
		return value, false
	}

	switch typedMask := mask.(type) {
	case map[string]any:
		valueMap, ok := value.(map[string]any)
		if !ok {
			return value, false
		}
		out := make(map[string]any, len(valueMap))
		for key, item := range valueMap {
			childMask := typedMask[key]
			child, redacted := applyMask(item, childMask)
			if redacted {
				out[key] = "<redacted>"
			} else {
				out[key] = child
			}
		}
		return out, false
	case []any:
		valueSlice, ok := value.([]any)
		if !ok {
			return value, false
		}
		out := make([]any, len(valueSlice))
		for i, item := range valueSlice {
			var childMask any
			if i < len(typedMask) {
				childMask = typedMask[i]
			}
			child, redacted := applyMask(item, childMask)
			if redacted {
				out[i] = "<redacted>"
			} else {
				out[i] = child
			}
		}
		return out, false
	default:
		return value, false
	}
}

func collectMaskPaths(raw json.RawMessage) ([]ir.AttributePath, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	var paths []ir.AttributePath
	walkMask(value, nil, &paths)
	return paths, nil
}

func walkMask(value any, path ir.AttributePath, paths *[]ir.AttributePath) {
	switch typed := value.(type) {
	case bool:
		if typed {
			copyPath := append(ir.AttributePath(nil), path...)
			*paths = append(*paths, copyPath)
		}
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			walkMask(typed[key], appendPath(path, ir.Attribute(key)), paths)
		}
	case []any:
		for i, item := range typed {
			walkMask(item, appendPath(path, ir.Index(strconv.Itoa(i))), paths)
		}
	}
}

func parsePathRaw(raw json.RawMessage) (ir.AttributePath, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return nil, nil
	}
	value, err := decodeJSON(raw)
	if err != nil {
		return nil, err
	}
	return parsePathValue(value)
}

func parsePathValue(value any) (ir.AttributePath, error) {
	switch typed := value.(type) {
	case string:
		return ir.AttributePath{ir.Attribute(typed)}, nil
	case []any:
		path := make(ir.AttributePath, 0, len(typed))
		for _, step := range typed {
			switch v := step.(type) {
			case string:
				path = append(path, ir.Attribute(v))
			case json.Number:
				path = append(path, ir.Index(v.String()))
			case float64:
				path = append(path, ir.Index(strconv.FormatFloat(v, 'f', -1, 64)))
			default:
				return nil, fmt.Errorf("unsupported path step type %T", step)
			}
		}
		return path, nil
	default:
		return nil, fmt.Errorf("unsupported path type %T", value)
	}
}

func appendPath(path ir.AttributePath, step ir.PathStep) ir.AttributePath {
	out := make(ir.AttributePath, len(path), len(path)+1)
	copy(out, path)
	return append(out, step)
}

func uniquePaths(paths []ir.AttributePath) []ir.AttributePath {
	seen := make(map[string]ir.AttributePath)
	for _, path := range paths {
		seen[path.String()] = path
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	out := make([]ir.AttributePath, 0, len(keys))
	for _, key := range keys {
		out = append(out, seen[key])
	}
	return out
}

func variableNames(variables map[string]json.RawMessage) []string {
	names := make([]string, 0, len(variables))
	for name := range variables {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func buildDependencyGraph(configuration Configuration, resources []ir.ResourceChange) ir.DependencyGraph {
	graph := ir.DependencyGraph{}
	changed := make(map[string]ir.NodeID, len(resources))
	for _, resource := range resources {
		id := ir.NodeID(resource.Address)
		changed[string(resource.Address)] = id
		kind := ir.NodeKindResource
		if resource.Mode == ir.ResourceModeData {
			kind = ir.NodeKindData
		}
		graph.Nodes = append(graph.Nodes, ir.Node{ID: id, Address: resource.Address, Kind: kind})
	}

	baseToChanges := make(map[string][]ir.NodeID)
	for _, resource := range flattenConfigResources(configuration.RootModule) {
		for addr, id := range changed {
			if addr == resource.Address || strings.HasPrefix(addr, resource.Address+"[") {
				baseToChanges[resource.Address] = append(baseToChanges[resource.Address], id)
			}
		}
	}

	edges := make(map[string]ir.Edge)
	for _, resource := range flattenConfigResources(configuration.RootModule) {
		targets := baseToChanges[resource.Address]
		if len(targets) == 0 {
			continue
		}
		for _, dependency := range resource.DependsOn {
			for _, source := range resolveReference(dependency, changed, baseToChanges) {
				for _, target := range targets {
					addEdge(edges, source, target, ir.EdgeExplicitDependsOn, ir.ConfidenceExact)
				}
			}
		}
		for _, rawExpression := range resource.Expressions {
			for _, reference := range extractReferences(rawExpression) {
				for _, source := range resolveReference(reference, changed, baseToChanges) {
					for _, target := range targets {
						addEdge(edges, source, target, ir.EdgeExpressionRef, ir.ConfidenceExact)
					}
				}
			}
		}
	}

	for _, edge := range edges {
		graph.Edges = append(graph.Edges, edge)
	}
	return graph
}

func addEdge(edges map[string]ir.Edge, from, to ir.NodeID, kind ir.EdgeKind, confidence ir.EvidenceConfidence) {
	if from == "" || to == "" || from == to {
		return
	}
	key := fmt.Sprintf("%s\x00%s\x00%s", from, to, kind)
	edges[key] = ir.Edge{From: from, To: to, Kind: kind, Confidence: confidence}
}

func flattenConfigResources(root Module) []ConfigResource {
	var out []ConfigResource
	var visit func(Module)
	visit = func(module Module) {
		out = append(out, module.Resources...)
		names := make([]string, 0, len(module.ModuleCalls))
		for name := range module.ModuleCalls {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			call := module.ModuleCalls[name]
			if call.Module != nil {
				visit(*call.Module)
			}
		}
	}
	visit(root)
	return out
}

func resolveReference(reference string, changed map[string]ir.NodeID, baseToChanges map[string][]ir.NodeID) []ir.NodeID {
	if id, ok := changed[reference]; ok {
		return []ir.NodeID{id}
	}
	if ids, ok := baseToChanges[reference]; ok {
		return append([]ir.NodeID(nil), ids...)
	}

	// Expression references commonly include an attribute traversal (for example,
	// aws_instance.web.id) while graph nodes are resource-instance addresses.
	// Match only at Terraform address boundaries so similarly-prefixed resources
	// such as foo and foobar are never conflated. Prefer the longest matching
	// address and return its instances when the reference targets an unkeyed
	// resource address.
	type candidate struct {
		address string
		ids     []ir.NodeID
	}
	var best candidate
	for address, id := range changed {
		if referenceHasAddressPrefix(reference, address) && len(address) > len(best.address) {
			best = candidate{address: address, ids: []ir.NodeID{id}}
		}
	}
	for address, ids := range baseToChanges {
		if referenceHasAddressPrefix(reference, address) && len(address) > len(best.address) {
			best = candidate{address: address, ids: append([]ir.NodeID(nil), ids...)}
		}
	}
	return best.ids
}

func referenceHasAddressPrefix(reference, address string) bool {
	if reference == address {
		return true
	}
	if !strings.HasPrefix(reference, address) || len(reference) == len(address) {
		return false
	}
	switch reference[len(address)] {
	case '.', '[':
		return true
	default:
		return false
	}
}

func extractReferences(raw json.RawMessage) []string {
	value, err := decodeJSON(raw)
	if err != nil {
		return nil
	}
	set := make(map[string]struct{})
	walkReferences(value, set)
	out := make([]string, 0, len(set))
	for ref := range set {
		out = append(out, ref)
	}
	sort.Strings(out)
	return out
}

func walkReferences(value any, refs map[string]struct{}) {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			if key == "references" {
				if values, ok := child.([]any); ok {
					for _, value := range values {
						if ref, ok := value.(string); ok {
							refs[ref] = struct{}{}
						}
					}
				}
			}
			walkReferences(child, refs)
		}
	case []any:
		for _, child := range typed {
			walkReferences(child, refs)
		}
	}
}
