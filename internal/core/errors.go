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

package core

import (
	"fmt"
)

// PlanParseError represents an error that occurs during plan parsing
type PlanParseError struct {
	File    string
	Message string
	Cause   error
}

func (e *PlanParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to parse plan file %s: %s: %v", e.File, e.Message, e.Cause)
	}
	return fmt.Sprintf("failed to parse plan file %s: %s", e.File, e.Message)
}

func (e *PlanParseError) Unwrap() error {
	return e.Cause
}

// ConfigParseError represents an error that occurs during configuration parsing
type ConfigParseError struct {
	Path    string
	Message string
	Cause   error
}

func (e *ConfigParseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to parse config at %s: %s: %v", e.Path, e.Message, e.Cause)
	}
	return fmt.Sprintf("failed to parse config at %s: %s", e.Path, e.Message)
}

func (e *ConfigParseError) Unwrap() error {
	return e.Cause
}

// GraphBuildError represents an error that occurs during graph building
type GraphBuildError struct {
	Message string
	Cause   error
}

func (e *GraphBuildError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to build graph: %s: %v", e.Message, e.Cause)
	}
	return fmt.Sprintf("failed to build graph: %s", e.Message)
}

func (e *GraphBuildError) Unwrap() error {
	return e.Cause
}

// GraphGenerationError represents an error that occurs during graph generation
type GraphGenerationError struct {
	Format  string
	Message string
	Cause   error
}

func (e *GraphGenerationError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("failed to generate %s graph: %s: %v", e.Format, e.Message, e.Cause)
	}
	return fmt.Sprintf("failed to generate %s graph: %s", e.Format, e.Message)
}

func (e *GraphGenerationError) Unwrap() error {
	return e.Cause
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error in field %s: %s", e.Field, e.Message)
}

// UnsupportedFormatError represents an error for unsupported formats
type UnsupportedFormatError struct {
	Format string
}

func (e *UnsupportedFormatError) Error() string {
	return fmt.Sprintf("unsupported format: %s", e.Format)
}
