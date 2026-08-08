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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const DefaultMaxPlanBytes int64 = 64 << 20

var ErrPlanTooLarge = errors.New("plan JSON exceeds maximum input size")

type Plan struct {
	FormatVersion      string                     `json:"format_version"`
	TerraformVersion   string                     `json:"terraform_version"`
	Applyable          bool                       `json:"applyable"`
	Complete           *bool                      `json:"complete"`
	Errored            bool                       `json:"errored"`
	ResourceChanges    []ResourceChange           `json:"resource_changes"`
	ResourceDrift      []ResourceChange           `json:"resource_drift"`
	RelevantAttributes []RelevantAttribute        `json:"relevant_attributes"`
	OutputChanges      map[string]OutputChange    `json:"output_changes"`
	Checks             []Check                    `json:"checks"`
	Configuration      Configuration              `json:"configuration"`
	Variables          map[string]json.RawMessage `json:"variables"`
}

type ResourceChange struct {
	Address         string          `json:"address"`
	PreviousAddress string          `json:"previous_address"`
	ModuleAddress   string          `json:"module_address"`
	Mode            string          `json:"mode"`
	Type            string          `json:"type"`
	Name            string          `json:"name"`
	Index           json.RawMessage `json:"index"`
	Deposed         string          `json:"deposed"`
	Change          Change          `json:"change"`
	ActionReason    string          `json:"action_reason"`
}

type Change struct {
	Actions         []string          `json:"actions"`
	Before          json.RawMessage   `json:"before"`
	After           json.RawMessage   `json:"after"`
	AfterUnknown    json.RawMessage   `json:"after_unknown"`
	BeforeSensitive json.RawMessage   `json:"before_sensitive"`
	AfterSensitive  json.RawMessage   `json:"after_sensitive"`
	ReplacePaths    []json.RawMessage `json:"replace_paths"`
	Importing       *Importing        `json:"importing"`
}

type Importing struct {
	ID      string `json:"id"`
	Unknown bool   `json:"unknown"`
}

type OutputChange struct {
	Change Change `json:"change"`
}

type RelevantAttribute struct {
	Resource  string          `json:"resource"`
	Attribute json.RawMessage `json:"attribute"`
}

type Check struct {
	Address   CheckAddress    `json:"address"`
	Status    string          `json:"status"`
	Instances []CheckInstance `json:"instances"`
}

type CheckAddress struct {
	Kind      string `json:"kind"`
	ToDisplay string `json:"to_display"`
	Module    string `json:"module"`
}

type CheckInstance struct {
	Address  CheckAddress   `json:"address"`
	Status   string         `json:"status"`
	Problems []CheckProblem `json:"problems"`
}

type CheckProblem struct {
	Message string `json:"message"`
}

type Configuration struct {
	RootModule Module `json:"root_module"`
}

type Module struct {
	Resources   []ConfigResource        `json:"resources"`
	Outputs     map[string]ConfigOutput `json:"outputs"`
	ModuleCalls map[string]ModuleCall   `json:"module_calls"`
}

type ConfigResource struct {
	Address     string                     `json:"address"`
	Mode        string                     `json:"mode"`
	Type        string                     `json:"type"`
	Name        string                     `json:"name"`
	Expressions map[string]json.RawMessage `json:"expressions"`
	DependsOn   []string                   `json:"depends_on"`
}

type ConfigOutput struct {
	Expression json.RawMessage `json:"expression"`
}

type ModuleCall struct {
	Expressions map[string]json.RawMessage `json:"expressions"`
	Module      *Module                    `json:"module"`
}

func ParseFile(path string, maxBytes int64) (*Plan, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open plan JSON: %w", err)
	}
	plan, parseErr := ParseReader(f, maxBytes)
	closeErr := f.Close()
	if parseErr != nil {
		return nil, parseErr
	}
	if closeErr != nil {
		return nil, fmt.Errorf("close plan JSON: %w", closeErr)
	}
	return plan, nil
}

func ParseReader(r io.Reader, maxBytes int64) (*Plan, error) {
	if maxBytes <= 0 {
		maxBytes = DefaultMaxPlanBytes
	}
	limited := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read plan JSON: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("%w: limit is %d bytes", ErrPlanTooLarge, maxBytes)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var plan Plan
	if err := dec.Decode(&plan); err != nil {
		return nil, fmt.Errorf("decode plan JSON: %w", err)
	}

	// Require exactly one JSON value. Decoder.Decode accepts a valid first value
	// even when another JSON value or malformed non-whitespace data follows it,
	// while plan files are expected to contain one complete JSON document.
	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, errors.New("decode plan JSON: multiple JSON values are not allowed")
		}
		return nil, fmt.Errorf("decode plan JSON: trailing data: %w", err)
	}

	if err := validateFormatVersion(plan.FormatVersion); err != nil {
		return nil, err
	}
	return &plan, nil
}

func validateFormatVersion(version string) error {
	if version == "" {
		return errors.New("plan JSON is missing format_version")
	}
	major, _, ok := strings.Cut(version, ".")
	if !ok {
		major = version
	}
	if major != "1" {
		return fmt.Errorf("unsupported plan format version %q: only major version 1 is supported", version)
	}
	return nil
}

func decodeJSON(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value any
	if err := dec.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}
