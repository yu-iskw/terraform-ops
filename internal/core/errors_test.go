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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlanParseError_WithCause(t *testing.T) {
	cause := errors.New("file not found")
	err := &PlanParseError{
		File:    "plan.json",
		Message: "failed to open file",
		Cause:   cause,
	}

	expected := "failed to parse plan file plan.json: failed to open file: file not found"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, cause, err.Unwrap())
}

func TestPlanParseError_WithoutCause(t *testing.T) {
	err := &PlanParseError{
		File:    "plan.json",
		Message: "invalid format",
		Cause:   nil,
	}

	expected := "failed to parse plan file plan.json: invalid format"
	assert.Equal(t, expected, err.Error())
	assert.Nil(t, err.Unwrap())
}

func TestConfigParseError_WithCause(t *testing.T) {
	cause := errors.New("permission denied")
	err := &ConfigParseError{
		Path:    "/path/to/config",
		Message: "failed to read directory",
		Cause:   cause,
	}

	expected := "failed to parse config at /path/to/config: failed to read directory: permission denied"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, cause, err.Unwrap())
}

func TestConfigParseError_WithoutCause(t *testing.T) {
	err := &ConfigParseError{
		Path:    "/path/to/config",
		Message: "not a directory",
		Cause:   nil,
	}

	expected := "failed to parse config at /path/to/config: not a directory"
	assert.Equal(t, expected, err.Error())
	assert.Nil(t, err.Unwrap())
}

func TestGraphBuildError_WithCause(t *testing.T) {
	cause := errors.New("invalid node data")
	err := &GraphBuildError{
		Message: "failed to create node",
		Cause:   cause,
	}

	expected := "failed to build graph: failed to create node: invalid node data"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, cause, err.Unwrap())
}

func TestGraphBuildError_WithoutCause(t *testing.T) {
	err := &GraphBuildError{
		Message: "no nodes found",
		Cause:   nil,
	}

	expected := "failed to build graph: no nodes found"
	assert.Equal(t, expected, err.Error())
	assert.Nil(t, err.Unwrap())
}

func TestGraphGenerationError_WithCause(t *testing.T) {
	cause := errors.New("unsupported format")
	err := &GraphGenerationError{
		Format:  "unknown",
		Message: "failed to generate graph",
		Cause:   cause,
	}

	expected := "failed to generate unknown graph: failed to generate graph: unsupported format"
	assert.Equal(t, expected, err.Error())
	assert.Equal(t, cause, err.Unwrap())
}

func TestGraphGenerationError_WithoutCause(t *testing.T) {
	err := &GraphGenerationError{
		Format:  "graphviz",
		Message: "invalid syntax",
		Cause:   nil,
	}

	expected := "failed to generate graphviz graph: invalid syntax"
	assert.Equal(t, expected, err.Error())
	assert.Nil(t, err.Unwrap())
}

func TestValidationError(t *testing.T) {
	err := &ValidationError{
		Field:   "format_version",
		Message: "missing required field",
	}

	expected := "validation error in field format_version: missing required field"
	assert.Equal(t, expected, err.Error())
}

func TestUnsupportedFormatError(t *testing.T) {
	err := &UnsupportedFormatError{
		Format: "unknown_format",
	}

	expected := "unsupported format: unknown_format"
	assert.Equal(t, expected, err.Error())
}

func TestErrorWrapping(t *testing.T) {
	// Test that errors can be wrapped and unwrapped properly
	originalErr := errors.New("original error")

	planErr := &PlanParseError{
		File:    "test.json",
		Message: "test error",
		Cause:   originalErr,
	}

	// Test unwrapping
	unwrapped := errors.Unwrap(planErr)
	assert.Equal(t, originalErr, unwrapped)

	// Test error chain
	assert.True(t, errors.Is(planErr, originalErr))
}
