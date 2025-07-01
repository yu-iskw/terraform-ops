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

package formatters

import (
	"fmt"

	"github.com/yu/terraform-ops/internal/core"
)

// SummaryFormatter defines the interface for formatting plan summaries
type SummaryFormatter interface {
	Format(summary *core.PlanSummary, opts core.SummaryOptions) (string, error)
}

// Factory creates formatters for different output formats
type Factory struct{}

// NewFactory creates a new formatter factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateFormatter creates a formatter for the specified format
func (f *Factory) CreateFormatter(format core.SummaryFormat, useColor bool) (SummaryFormatter, error) {
	switch format {
	case core.FormatText:
		return NewTextFormatter(useColor), nil
	case core.FormatJSON:
		return NewJSONFormatter(), nil
	case core.FormatMarkdown:
		return NewMarkdownFormatter(), nil
	case core.FormatTable:
		return NewTableFormatter(), nil
	case core.FormatPlan:
		return NewPlanFormatter(useColor), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}
