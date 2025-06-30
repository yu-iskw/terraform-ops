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

package plan

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/yu/terraform-ops/internal/core"
)

// Parser implements the core.PlanParser interface
type Parser struct{}

// NewParser creates a new plan parser
func NewParser() *Parser {
	return &Parser{}
}

// ParsePlanFile reads and parses a Terraform plan JSON file
func (p *Parser) ParsePlanFile(filename string) (*core.TerraformPlan, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, &core.PlanParseError{
			File:    filename,
			Message: "failed to open file",
			Cause:   err,
		}
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = &core.PlanParseError{
				File:    filename,
				Message: "failed to close file",
				Cause:   closeErr,
			}
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, &core.PlanParseError{
			File:    filename,
			Message: "failed to read file",
			Cause:   err,
		}
	}

	var plan core.TerraformPlan
	if err := json.Unmarshal(data, &plan); err != nil {
		return nil, &core.PlanParseError{
			File:    filename,
			Message: "failed to parse JSON",
			Cause:   err,
		}
	}

	// Validate the plan structure
	if err := p.validatePlan(&plan); err != nil {
		return nil, &core.PlanParseError{
			File:    filename,
			Message: "invalid plan structure",
			Cause:   err,
		}
	}

	return &plan, nil
}

// validatePlan checks if the plan has a valid structure
func (p *Parser) validatePlan(plan *core.TerraformPlan) error {
	if plan.FormatVersion == "" {
		return &core.ValidationError{
			Field:   "format_version",
			Message: "missing format_version",
		}
	}

	// Check if this is a valid format version we support
	// According to the official spec, minor version increments (1.1, 1.2, etc.)
	// are backward-compatible changes, so we support any 1.x version
	if !strings.HasPrefix(plan.FormatVersion, "1.") {
		return &core.ValidationError{
			Field:   "format_version",
			Message: fmt.Sprintf("unsupported format version: %s (only 1.x versions are supported)", plan.FormatVersion),
		}
	}

	return nil
}
