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

package generators

import (
	"github.com/yu/terraform-ops/internal/core"
)

// Factory creates graph generators for different formats
type Factory struct{}

// NewFactory creates a new graph generator factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateGenerator creates a graph generator for the specified format
func (f *Factory) CreateGenerator(format core.GraphFormat) (core.GraphGenerator, error) {
	switch format {
	case core.FormatGraphviz:
		return NewGraphvizGenerator(), nil
	case core.FormatMermaid:
		return NewMermaidGenerator(), nil
	case core.FormatPlantUML:
		return NewPlantUMLGenerator(), nil
	default:
		return nil, &core.UnsupportedFormatError{Format: string(format)}
	}
}
