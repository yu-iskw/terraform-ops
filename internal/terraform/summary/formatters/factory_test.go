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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/yu/terraform-ops/internal/core"
)

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
}

func TestFactory_CreateFormatter(t *testing.T) {
	factory := NewFactory()

	tests := []struct {
		format       core.SummaryFormat
		useColor     bool
		expectedType interface{}
		expectError  bool
	}{
		{core.FormatText, false, (*TextFormatter)(nil), false},
		{core.FormatText, true, (*TextFormatter)(nil), false},
		{core.FormatJSON, false, (*JSONFormatter)(nil), false},
		{core.FormatJSON, true, (*JSONFormatter)(nil), false},
		{core.FormatMarkdown, false, (*MarkdownFormatter)(nil), false},
		{core.FormatMarkdown, true, (*MarkdownFormatter)(nil), false},
		{core.FormatTable, false, (*TableFormatter)(nil), false},
		{core.FormatTable, true, (*TableFormatter)(nil), false},
		{core.FormatPlan, false, (*PlanFormatter)(nil), false},
		{core.FormatPlan, true, (*PlanFormatter)(nil), false},
		{core.SummaryFormat("invalid"), false, nil, true},
	}

	for _, test := range tests {
		t.Run(string(test.format), func(t *testing.T) {
			formatter, err := factory.CreateFormatter(test.format, test.useColor)

			if test.expectError {
				assert.Error(t, err)
				assert.Nil(t, formatter)
				assert.Contains(t, err.Error(), "unsupported format")
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, formatter)

				// Check the type of formatter returned
				switch test.format {
				case core.FormatText:
					textFormatter, ok := formatter.(*TextFormatter)
					assert.True(t, ok)
					assert.Equal(t, test.useColor, textFormatter.useColor)
				case core.FormatJSON:
					_, ok := formatter.(*JSONFormatter)
					assert.True(t, ok)
				case core.FormatMarkdown:
					_, ok := formatter.(*MarkdownFormatter)
					assert.True(t, ok)
				case core.FormatTable:
					_, ok := formatter.(*TableFormatter)
					assert.True(t, ok)
				case core.FormatPlan:
					planFormatter, ok := formatter.(*PlanFormatter)
					assert.True(t, ok)
					assert.Equal(t, test.useColor, planFormatter.useColor)
				}
			}
		})
	}
}

func TestFactory_CreateFormatter_UnsupportedFormat(t *testing.T) {
	factory := NewFactory()

	formatter, err := factory.CreateFormatter("unsupported", false)

	assert.Error(t, err)
	assert.Nil(t, formatter)
	assert.Contains(t, err.Error(), "unsupported format: unsupported")
}
