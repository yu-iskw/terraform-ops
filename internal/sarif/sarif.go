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

package sarif

import (
	"bytes"
	"encoding/json"
	"sort"

	"github.com/yu/terraform-ops/internal/report"
	"github.com/yu/terraform-ops/internal/sourceindex"
)

const schemaURI = "https://json.schemastore.org/sarif-2.1.0.json"

// Render converts exact, source-located terraform-ops findings to SARIF 2.1.0.
// Findings without an exact local location are deliberately omitted because
// GitHub code scanning requires a physical location for a useful result.
func Render(analysis report.AnalysisReport, located []sourceindex.LocatedFinding) ([]byte, error) {
	sort.SliceStable(located, func(i, j int) bool {
		a, b := located[i], located[j]
		if a.Finding.RuleID != b.Finding.RuleID {
			return a.Finding.RuleID < b.Finding.RuleID
		}
		if a.Location.Path != b.Location.Path {
			return a.Location.Path < b.Location.Path
		}
		if a.Location.StartLine != b.Location.StartLine {
			return a.Location.StartLine < b.Location.StartLine
		}
		return resourceAddress(a.Finding) < resourceAddress(b.Finding)
	})

	rulesByID := map[string]report.Finding{}
	for _, item := range located {
		if _, exists := rulesByID[item.Finding.RuleID]; !exists {
			rulesByID[item.Finding.RuleID] = item.Finding
		}
	}
	ruleIDs := make([]string, 0, len(rulesByID))
	for id := range rulesByID {
		ruleIDs = append(ruleIDs, id)
	}
	sort.Strings(ruleIDs)

	rules := make([]reportingDescriptor, 0, len(ruleIDs))
	for _, id := range ruleIDs {
		finding := rulesByID[id]
		rules = append(rules, reportingDescriptor{
			ID:               id,
			Name:             id,
			ShortDescription: message{Text: finding.Title},
			Properties: descriptorProperties{
				Category: string(finding.Category),
			},
		})
	}

	results := make([]result, 0, len(located))
	for _, item := range located {
		finding := item.Finding
		location := item.Location
		properties := resultProperties{
			Category:   string(finding.Category),
			Confidence: string(finding.Confidence),
		}
		if finding.Resource != nil {
			properties.Resource = finding.Resource.Address
			properties.ResourceType = finding.Resource.Type
		}
		results = append(results, result{
			RuleID: finding.RuleID,
			Level:  sarifLevel(finding.Severity),
			Message: message{Text: finding.Message},
			Locations: []locationWrapper{{PhysicalLocation: physicalLocation{
				ArtifactLocation: artifactLocation{URI: location.Path},
				Region: region{
					StartLine:   location.StartLine,
					StartColumn: location.StartColumn,
					EndLine:     location.EndLine,
					EndColumn:   location.EndColumn,
				},
			}}},
			Properties: properties,
		})
	}

	doc := sarifLog{
		Version: "2.1.0",
		Schema:  schemaURI,
		Runs: []run{{
			Tool: tool{Driver: driver{
				Name:           analysis.Tool.Name,
				Version:        analysis.Tool.Version,
				InformationURI: "https://github.com/yu-iskw/terraform-ops",
				Rules:          rules,
			}},
			Results: results,
		}},
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func sarifLevel(severity report.Severity) string {
	switch severity {
	case report.SeverityCritical, report.SeverityHigh:
		return "error"
	case report.SeverityMedium:
		return "warning"
	default:
		return "note"
	}
}

func resourceAddress(finding report.Finding) string {
	if finding.Resource == nil {
		return ""
	}
	return finding.Resource.Address
}

type sarifLog struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []run  `json:"runs"`
}

type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}

type tool struct {
	Driver driver `json:"driver"`
}

type driver struct {
	Name           string                `json:"name"`
	Version        string                `json:"version,omitempty"`
	InformationURI string                `json:"informationUri,omitempty"`
	Rules          []reportingDescriptor `json:"rules,omitempty"`
}

type reportingDescriptor struct {
	ID               string               `json:"id"`
	Name             string               `json:"name,omitempty"`
	ShortDescription message              `json:"shortDescription"`
	Properties       descriptorProperties `json:"properties,omitempty"`
}

type descriptorProperties struct {
	Category string `json:"category,omitempty"`
}

type result struct {
	RuleID     string            `json:"ruleId"`
	Level      string            `json:"level"`
	Message    message           `json:"message"`
	Locations  []locationWrapper `json:"locations"`
	Properties resultProperties  `json:"properties,omitempty"`
}

type message struct {
	Text string `json:"text"`
}

type locationWrapper struct {
	PhysicalLocation physicalLocation `json:"physicalLocation"`
}

type physicalLocation struct {
	ArtifactLocation artifactLocation `json:"artifactLocation"`
	Region           region           `json:"region"`
}

type artifactLocation struct {
	URI string `json:"uri"`
}

type region struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
	EndLine     int `json:"endLine,omitempty"`
	EndColumn   int `json:"endColumn,omitempty"`
}

type resultProperties struct {
	Category     string `json:"category,omitempty"`
	Confidence   string `json:"confidence,omitempty"`
	Resource     string `json:"resource,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
}
