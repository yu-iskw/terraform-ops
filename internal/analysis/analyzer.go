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

package analysis

import (
	"context"
	"fmt"

	"github.com/yu/terraform-ops/internal/ir"
	"github.com/yu/terraform-ops/internal/report"
)

type Analyzer interface {
	ID() string
	Analyze(context.Context, *ir.ChangeSet) ([]report.Finding, error)
}

type Registry struct {
	analyzers []Analyzer
}

func NewRegistry(analyzers ...Analyzer) *Registry {
	return &Registry{analyzers: append([]Analyzer(nil), analyzers...)}
}

func DefaultRegistry() *Registry {
	return NewRegistry(
		planAnalyzer{},
		checkAnalyzer{},
		lifecycleAnalyzer{},
		driftAnalyzer{},
		sensitivityAnalyzer{},
		unknownAnalyzer{},
	)
}

func (r *Registry) Analyze(ctx context.Context, changeSet *ir.ChangeSet) ([]report.Finding, error) {
	if changeSet == nil {
		return nil, fmt.Errorf("change set is nil")
	}
	var findings []report.Finding
	for _, analyzer := range r.analyzers {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		items, err := analyzer.Analyze(ctx, changeSet)
		if err != nil {
			return nil, fmt.Errorf("analyzer %s: %w", analyzer.ID(), err)
		}
		findings = append(findings, items...)
	}
	wrapper := report.AnalysisReport{Findings: findings}
	report.Sort(&wrapper)
	return wrapper.Findings, nil
}
