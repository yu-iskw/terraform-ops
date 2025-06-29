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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yu/terraform-ops/internal/core"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestCreateGenerator_Graphviz(t *testing.T) {
	factory := NewFactory()
	generator, err := factory.CreateGenerator(core.FormatGraphviz)

	assert.NoError(t, err)
	assert.NotNil(t, generator)

	// Check that it's the right type
	_, ok := generator.(*GraphvizGenerator)
	assert.True(t, ok)
}

func TestCreateGenerator_Mermaid(t *testing.T) {
	factory := NewFactory()
	generator, err := factory.CreateGenerator(core.FormatMermaid)

	assert.NoError(t, err)
	assert.NotNil(t, generator)

	// Check that it's the right type
	_, ok := generator.(*MermaidGenerator)
	assert.True(t, ok)
}

func TestCreateGenerator_PlantUML(t *testing.T) {
	factory := NewFactory()
	generator, err := factory.CreateGenerator(core.FormatPlantUML)

	assert.NoError(t, err)
	assert.NotNil(t, generator)

	// Check that it's the right type
	_, ok := generator.(*PlantUMLGenerator)
	assert.True(t, ok)
}

func TestCreateGenerator_UnsupportedFormat(t *testing.T) {
	factory := NewFactory()
	generator, err := factory.CreateGenerator("unsupported_format")

	assert.Error(t, err)
	assert.Nil(t, generator)

	// Check that it's the right error type
	_, ok := err.(*core.UnsupportedFormatError)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestCreateGenerator_EmptyFormat(t *testing.T) {
	factory := NewFactory()
	generator, err := factory.CreateGenerator("")

	assert.Error(t, err)
	assert.Nil(t, generator)

	// Check that it's the right error type
	_, ok := err.(*core.UnsupportedFormatError)
	assert.True(t, ok)
	assert.Contains(t, err.Error(), "unsupported format")
}

func TestCreateGenerator_AllSupportedFormats(t *testing.T) {
	factory := NewFactory()

	supportedFormats := []core.GraphFormat{
		core.FormatGraphviz,
		core.FormatMermaid,
		core.FormatPlantUML,
	}

	for _, format := range supportedFormats {
		t.Run(string(format), func(t *testing.T) {
			generator, err := factory.CreateGenerator(format)

			assert.NoError(t, err)
			assert.NotNil(t, generator)

			// Verify it implements the GraphGenerator interface
			assert.Implements(t, (*core.GraphGenerator)(nil), generator)
		})
	}
}
