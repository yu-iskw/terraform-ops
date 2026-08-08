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

package sourceindex

import "github.com/yu/terraform-ops/internal/report"

// AttachLocations enriches only resource-scoped findings for which the index can
// prove an exact local source range. Unresolved and plan-level findings remain
// location-free and are therefore not emitted as SARIF code-scanning results.
func AttachLocations(findings []report.Finding, idx *Index) []report.Finding {
	out := append([]report.Finding(nil), findings...)
	for i := range out {
		if out[i].Resource == nil || idx == nil {
			continue
		}
		location, ok := idx.Resolve(out[i].Resource.Address)
		if !ok {
			continue
		}
		out[i].Location = &report.SourceLocation{
			Path:        location.Path,
			StartLine:   location.StartLine,
			StartColumn: location.StartColumn,
			EndLine:     location.EndLine,
			EndColumn:   location.EndColumn,
		}
	}
	return out
}
