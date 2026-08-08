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
	"strings"
	"testing"
)

func TestParseReaderAllowsTrailingWhitespace(t *testing.T) {
	plan, err := ParseReader(strings.NewReader(`{"format_version":"1.0"} 
	`), DefaultMaxPlanBytes)
	if err != nil {
		t.Fatalf("ParseReader() error = %v", err)
	}
	if plan.FormatVersion != "1.0" {
		t.Fatalf("FormatVersion = %q, want 1.0", plan.FormatVersion)
	}
}

func TestParseReaderRejectsTrailingData(t *testing.T) {
	const planJSON = `{"format_version":"1.0"}`
	tests := []struct {
		name  string
		suffix string
	}{
		{name: "second object", suffix: `{"format_version":"1.0"}`},
		{name: "second scalar", suffix: `null`},
		{name: "arbitrary bytes", suffix: `garbage`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseReader(strings.NewReader(planJSON+tt.suffix), DefaultMaxPlanBytes)
			if err == nil {
				t.Fatal("ParseReader() error = nil, want trailing-data error")
			}
			if !strings.Contains(err.Error(), "decode plan JSON") {
				t.Fatalf("ParseReader() error = %q, want decode plan JSON context", err)
			}
		})
	}
}
