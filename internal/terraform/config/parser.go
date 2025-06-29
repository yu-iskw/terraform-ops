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

package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"

	"github.com/yu/terraform-ops/internal/core"
)

// Parser implements the core.ConfigParser interface
type Parser struct{}

// NewParser creates a new configuration parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseConfigFiles scans the provided paths (non-recursive) for *.tf files and
// extracts information from the terraform block
func (p *Parser) ParseConfigFiles(paths []string) ([]core.TerraformConfig, error) {
	var allConfigs []core.TerraformConfig

	for _, path := range paths {
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, &core.ConfigParseError{
				Path:    path,
				Message: "failed to get absolute path",
				Cause:   err,
			}
		}

		info, err := os.Stat(path)
		if os.IsNotExist(err) || !info.IsDir() {
			return nil, &core.ConfigParseError{
				Path:    path,
				Message: "path does not exist or is not a directory",
				Cause:   err,
			}
		}

		tfFiles, err := p.findTerraformFiles(path)
		if err != nil {
			return nil, &core.ConfigParseError{
				Path:    path,
				Message: "failed to find terraform files",
				Cause:   err,
			}
		}

		config := core.TerraformConfig{
			Path:              absPath,
			RequiredProviders: map[string]string{},
		}

		for _, filePath := range tfFiles {
			if err := p.collectFromFile(filePath, &config); err != nil {
				// Log error but continue processing other files
				fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
				continue
			}
		}

		allConfigs = append(allConfigs, config)
	}

	return allConfigs, nil
}

// findTerraformFiles discovers all .tf files directly under the specified directory
func (p *Parser) findTerraformFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %w", dirPath, err)
	}

	var tfFiles []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".tf") {
			tfFiles = append(tfFiles, filepath.Join(dirPath, e.Name()))
		}
	}
	return tfFiles, nil
}

// collectFromFile parses an individual .tf file and populates the provided TerraformConfig
func (p *Parser) collectFromFile(filePath string, dest *core.TerraformConfig) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return &core.ConfigParseError{
			Path:    filePath,
			Message: "failed to read file",
			Cause:   err,
		}
	}

	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(src, filepath.Base(filePath))
	if diags.HasErrors() {
		return &core.ConfigParseError{
			Path:    filePath,
			Message: "failed to parse HCL",
			Cause:   diags,
		}
	}

	if hclFile == nil || hclFile.Body == nil {
		return nil // nothing to do
	}

	rootSchema := &hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "terraform"},
		},
	}

	content, _, diags := hclFile.Body.PartialContent(rootSchema)
	if diags.HasErrors() {
		return &core.ConfigParseError{
			Path:    filePath,
			Message: "failed to process content",
			Cause:   diags,
		}
	}

	for _, block := range content.Blocks {
		if block.Type != "terraform" {
			continue
		}

		// Handle attributes: required_version
		terraformAttrsSchema := &hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "required_version"},
			},
			Blocks: []hcl.BlockHeaderSchema{
				{Type: "required_providers"},
				{Type: "backend", LabelNames: []string{"type"}},
			},
		}

		bodyContent, _, diags := block.Body.PartialContent(terraformAttrsSchema)
		if diags.HasErrors() {
			fmt.Fprintf(os.Stderr, "Warning: Error parsing terraform block in '%s': %v\n", filePath, diags.Error())
		}

		// required_version attribute
		if attr, ok := bodyContent.Attributes["required_version"]; ok {
			val, _ := attr.Expr.Value(nil)
			if val.Type().IsPrimitiveType() {
				v := val.AsString()
				// prefer first non-empty encountered required_version
				if dest.RequiredVersion == "" && v != "" {
					dest.RequiredVersion = v
				}
			}
		}

		// required_providers block
		for _, b := range bodyContent.Blocks {
			switch b.Type {
			case "required_providers":
				p.parseRequiredProvidersBlock(b.Body, dest.RequiredProviders, filePath)
			case "backend":
				p.parseBackendBlock(b, dest, filePath)
			}
		}
	}

	return nil
}

// parseRequiredProvidersBlock fills providers map with any versions found
func (p *Parser) parseRequiredProvidersBlock(body hcl.Body, providers map[string]string, filePath string) {
	attrs, diags := body.JustAttributes()
	if diags.HasErrors() {
		fmt.Fprintf(os.Stderr, "Warning: Error getting attributes from required_providers block in '%s': %v\n", filePath, diags.Error())
		return
	}

	for name, attr := range attrs {
		val, diags := attr.Expr.Value(nil)
		if diags.HasErrors() {
			fmt.Fprintf(os.Stderr, "Warning: Error evaluating attribute '%s' in required_providers block in '%s': %v\n", name, filePath, diags.Error())
			continue
		}

		version := ""
		if val.Type().IsObjectType() {
			if v, ok := val.AsValueMap()["version"]; ok && v.Type().IsPrimitiveType() {
				version = v.AsString()
			}
		} else if val.Type().IsPrimitiveType() {
			// Handle simple string version constraints
			version = val.AsString()
		}

		// overwrite or set
		providers[name] = version
	}
}

// parseBackendBlock extracts backend type and simple config attributes
func (p *Parser) parseBackendBlock(block *hcl.Block, dest *core.TerraformConfig, filePath string) {
	if len(block.Labels) == 0 {
		return // malformed backend block
	}
	backendType := block.Labels[0]

	// If we already have a backend recorded, skip subsequent ones to keep first occurrence
	if dest.Backend != nil {
		return
	}

	backendInfo := &core.Backend{
		Type:   backendType,
		Config: map[string]string{},
	}

	attrs, diags := block.Body.JustAttributes()
	if diags.HasErrors() {
		fmt.Fprintf(os.Stderr, "Warning: Error reading backend attributes in '%s': %v\n", filePath, diags.Error())
	}

	for key, attr := range attrs {
		val, _ := attr.Expr.Value(nil)
		if !val.IsKnown() {
			continue
		}

		switch val.Type() {
		case cty.String:
			backendInfo.Config[key] = val.AsString()
		case cty.Bool:
			backendInfo.Config[key] = strconv.FormatBool(val.True())
		case cty.Number:
			backendInfo.Config[key] = val.AsBigFloat().Text('f', -1)
		}
	}

	dest.Backend = backendInfo
}
