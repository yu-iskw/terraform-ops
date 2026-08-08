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

// LocatedFinding pairs a deterministic analysis finding with an exact local
// declaration range. Keeping this mapping outside AnalysisReport avoids making
// repository layout part of the stable report schema.
type LocatedFinding struct {
	Finding  report.Finding
	Location Location
}

// LocateFindings returns only resource-scoped findings for which the index can
// prove an exact local source range. Unresolved and plan-level findings are
// intentionally excluded from source-aware outputs such as SARIF.
func LocateFindings(findings []report.Finding, idx *Index) []LocatedFinding {
	if idx == nil {
		return nil
	}
	out := make([]LocatedFinding, 0, len(findings))
	for _, finding := range findings {
		if finding.Resource == nil {
			continue
		}
		location, ok := idx.Resolve(finding.Resource.Address)
		if !ok {
			continue
		}
		out = append(out, LocatedFinding{Finding: finding, Location: location})
	}
	return out
}
