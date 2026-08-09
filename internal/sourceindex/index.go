// Copyright 2026 yu-iskw
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sourceindex

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// Location identifies an exact source range in a Terraform configuration file.
// Paths are artifact-root-relative and use slash separators so they are directly
// usable as SARIF artifact URIs.
type Location struct {
	Path        string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
}

// Index maps static Terraform resource addresses to source ranges. Dynamic
// instance keys are intentionally normalized at lookup time because all
// instances of a resource declaration share the same source block.
type Index struct {
	workspaceRoot string
	artifactRoot  string
	resources     map[string]Location
}

// Build indexes .tf files under workspaceRoot and emits locations relative to
// that same root.
func Build(workspaceRoot string) (*Index, error) {
	return BuildWithArtifactRoot(workspaceRoot, workspaceRoot)
}

// BuildWithArtifactRoot indexes .tf files under workspaceRoot while emitting
// source paths relative to artifactRoot. artifactRoot must contain
// workspaceRoot after canonical symlink resolution. This lets GitHub Actions
// constrain module traversal to one Terraform workspace while producing SARIF
// URIs relative to the repository checkout root.
func BuildWithArtifactRoot(workspaceRoot, artifactRoot string) (*Index, error) {
	workspace, err := canonicalDirectory(workspaceRoot, "workspace root")
	if err != nil {
		return nil, err
	}
	artifact, err := canonicalDirectory(artifactRoot, "artifact root")
	if err != nil {
		return nil, err
	}
	if !pathWithin(artifact, workspace) {
		return nil, fmt.Errorf("workspace root %q is outside artifact root %q", workspaceRoot, artifactRoot)
	}

	idx := &Index{
		workspaceRoot: workspace,
		artifactRoot:  artifact,
		resources:     map[string]Location{},
	}
	if err := idx.walkModule(workspace, "", map[string]bool{}); err != nil {
		return nil, err
	}
	return idx, nil
}

func canonicalDirectory(path, label string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("%s is empty", label)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve %s: %w", label, err)
	}
	canonical, err := filepath.EvalSymlinks(absolute)
	if err != nil {
		return "", fmt.Errorf("resolve %s symlinks: %w", label, err)
	}
	info, err := os.Stat(canonical)
	if err != nil {
		return "", fmt.Errorf("stat %s: %w", label, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s %q is not a directory", label, path)
	}
	return canonical, nil
}

func pathWithin(root, candidate string) bool {
	rel, err := filepath.Rel(root, candidate)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

// Resolve returns the declaration range for a resource or data-source instance.
// Instance keys in both module and resource addresses are stripped before lookup.
func (i *Index) Resolve(address string) (Location, bool) {
	if i == nil {
		return Location{}, false
	}
	location, ok := i.resources[staticAddress(address)]
	return location, ok
}

func (i *Index) walkModule(dir, modulePrefix string, stack map[string]bool) error {
	key := dir + "\x00" + modulePrefix
	if stack[key] {
		return nil
	}
	stack[key] = true
	defer delete(stack, key)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read module directory %q: %w", dir, err)
	}
	sort.Slice(entries, func(a, b int) bool { return entries[a].Name() < entries[b].Name() })

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tf") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		if err := i.indexFile(path, modulePrefix, stack); err != nil {
			return err
		}
	}
	return nil
}

func (i *Index) indexFile(path, modulePrefix string, stack map[string]bool) error {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(path)
	if diags.HasErrors() {
		return fmt.Errorf("parse Terraform source %q: %s", path, diags.Error())
	}
	if file == nil || file.Body == nil {
		return nil
	}

	schema := &hcl.BodySchema{Blocks: []hcl.BlockHeaderSchema{
		{Type: "resource", LabelNames: []string{"type", "name"}},
		{Type: "data", LabelNames: []string{"type", "name"}},
		{Type: "module", LabelNames: []string{"name"}},
	}}
	content, _, diags := file.Body.PartialContent(schema)
	if diags.HasErrors() {
		return fmt.Errorf("index Terraform source %q: %s", path, diags.Error())
	}

	for _, block := range content.Blocks {
		switch block.Type {
		case "resource":
			if len(block.Labels) != 2 {
				continue
			}
			i.resources[modulePrefix+block.Labels[0]+"."+block.Labels[1]] = i.location(block.DefRange)
		case "data":
			if len(block.Labels) != 2 {
				continue
			}
			i.resources[modulePrefix+"data."+block.Labels[0]+"."+block.Labels[1]] = i.location(block.DefRange)
		case "module":
			if len(block.Labels) != 1 {
				continue
			}
			source, ok := localModuleSource(block.Body)
			if !ok {
				continue
			}
			child, ok, err := i.resolveLocalModule(filepath.Dir(path), source)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			prefix := modulePrefix + "module." + block.Labels[0] + "."
			if err := i.walkModule(child, prefix, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

func localModuleSource(body hcl.Body) (string, bool) {
	content, _, diags := body.PartialContent(&hcl.BodySchema{Attributes: []hcl.AttributeSchema{{Name: "source"}}})
	if diags.HasErrors() {
		return "", false
	}
	attr, ok := content.Attributes["source"]
	if !ok {
		return "", false
	}
	value, diags := attr.Expr.Value(nil)
	if diags.HasErrors() || !value.IsKnown() || value.IsNull() || value.Type() != cty.String {
		return "", false
	}
	source := value.AsString()
	if !strings.HasPrefix(source, "./") && !strings.HasPrefix(source, "../") {
		return "", false
	}
	return source, true
}

func (i *Index) resolveLocalModule(parent, source string) (string, bool, error) {
	candidate, err := filepath.EvalSymlinks(filepath.Join(parent, filepath.FromSlash(source)))
	if err != nil {
		if os.IsNotExist(err) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("resolve local module %q: %w", source, err)
	}
	if !pathWithin(i.workspaceRoot, candidate) {
		return "", false, nil
	}
	info, err := os.Stat(candidate)
	if err != nil || !info.IsDir() {
		return "", false, nil
	}
	return candidate, true, nil
}

func (i *Index) location(r hcl.Range) Location {
	rel, err := filepath.Rel(i.artifactRoot, r.Filename)
	if err != nil {
		rel = r.Filename
	}
	return Location{
		Path:        filepath.ToSlash(rel),
		StartLine:   r.Start.Line,
		StartColumn: r.Start.Column,
		EndLine:     r.End.Line,
		EndColumn:   r.End.Column,
	}
}

// staticAddress removes Terraform/OpenTofu instance keys while preserving the
// surrounding address syntax. Quoted for_each keys are handled so a ']' inside
// a string cannot terminate a key prematurely.
func staticAddress(address string) string {
	var b strings.Builder
	for pos := 0; pos < len(address); {
		if address[pos] != '[' {
			b.WriteByte(address[pos])
			pos++
			continue
		}
		pos++
		quoted := false
		escaped := false
		for pos < len(address) {
			ch := address[pos]
			pos++
			if escaped {
				escaped = false
				continue
			}
			if quoted && ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				quoted = !quoted
				continue
			}
			if ch == ']' && !quoted {
				break
			}
		}
	}
	return b.String()
}
