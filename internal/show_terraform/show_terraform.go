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

package show_terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/zclconf/go-cty/cty"
)

// Backend represents the backend configuration discovered in a terraform block.
// Type is the backend type (e.g., "s3", "gcs", "azurerm")
// Config holds any key/value settings found within the backend block. Only
// primitive (string, number, bool) values are captured – complex types are ignored.
type Backend struct {
	Type   string            `json:"type"`
	Config map[string]string `json:"config,omitempty"`
}

// TerraformConfig represents the terraform block configuration details.
type TerraformConfig struct {
	RequiredVersion   string            `json:"required_version,omitempty"`
	Backend           *Backend          `json:"backend,omitempty"`
	RequiredProviders map[string]string `json:"required_providers"`
}

// TerraformInfo represents the aggregated information for a workspace directory.
type TerraformInfo struct {
	Path      string          `json:"path"`
	Terraform TerraformConfig `json:"terraform"`
}

// GetTerraformInfo scans the provided paths (non-recursive) for *.tf files and
// extracts information from the terraform block: required_version, backend, and
// required_providers. It returns a slice ordered in the same order as the input
// paths. Any individual workspace errors are printed to stderr, but do not stop
// processing of remaining paths.
func GetTerraformInfo(paths []string) ([]TerraformInfo, error) {
	var allInfo []TerraformInfo

	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting absolute path for '%s': %v\n", p, err)
			continue
		}

		info, err := os.Stat(p)
		if os.IsNotExist(err) || !info.IsDir() {
			fmt.Fprintf(os.Stderr, "Error: Path '%s' does not exist or is not a directory\n", p)
			continue
		}

		tfFiles, err := findTerraformFiles(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			continue
		}

		workspace := TerraformInfo{
			Path:      absPath,
			Terraform: TerraformConfig{RequiredProviders: map[string]string{}},
		}

		for _, filePath := range tfFiles {
			src, err := os.ReadFile(filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading file '%s': %v\n", filePath, err)
				continue
			}

			// Extract info from this file
			err = collectFromFile(filePath, src, &workspace)
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				continue
			}
		}

		allInfo = append(allInfo, workspace)
	}

	return allInfo, nil
}

// findTerraformFiles discovers all .tf files directly under the specified directory.
func findTerraformFiles(dirPath string) ([]string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("error reading directory '%s': %v", dirPath, err)
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

// collectFromFile parses an individual .tf file and populates the provided TerraformInfo pointer.
func collectFromFile(filePath string, src []byte, dest *TerraformInfo) error {
	parser := hclparse.NewParser()
	hclFile, diags := parser.ParseHCL(src, filepath.Base(filePath))
	if diags.HasErrors() {
		return fmt.Errorf("error parsing file '%s': %v", filePath, diags.Error())
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
		return fmt.Errorf("error processing content of '%s': %v", filePath, diags.Error())
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
				if dest.Terraform.RequiredVersion == "" && v != "" {
					dest.Terraform.RequiredVersion = v
				}
			}
		}

		// required_providers block
		for _, b := range bodyContent.Blocks {
			switch b.Type {
			case "required_providers":
				parseRequiredProvidersBlock(b.Body, dest.Terraform.RequiredProviders, filePath)
			case "backend":
				parseBackendBlock(b, &dest.Terraform, filePath)
			}
		}
	}

	return nil
}

// parseRequiredProvidersBlock fills providers map with any versions found.
func parseRequiredProvidersBlock(body hcl.Body, providers map[string]string, filePath string) {
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
		}

		// overwrite or set
		providers[name] = version
	}
}

// parseBackendBlock extracts backend type and simple config attributes.
func parseBackendBlock(block *hcl.Block, dest *TerraformConfig, filePath string) {
	if len(block.Labels) == 0 {
		return // malformed backend block
	}
	backendType := block.Labels[0]

	// If we already have a backend recorded, skip subsequent ones to keep first occurrence.
	if dest.Backend != nil {
		return
	}

	backendInfo := &Backend{
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
