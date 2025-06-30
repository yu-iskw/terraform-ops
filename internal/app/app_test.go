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

package app

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

// TestRootCmd ensures the root command is set up correctly
func TestRootCmd(t *testing.T) {
	assert.Equal(t, "terraform-ops", rootCmd.Use)
	assert.Equal(t, "Terraform operations CLI tool", rootCmd.Short)
	assert.NotEmpty(t, rootCmd.Long)

	// Check that subcommands are added
	assert.True(t, rootCmd.HasSubCommands())
	assert.NotNil(t, findCommand(rootCmd, "show-terraform"))
	assert.NotNil(t, findCommand(rootCmd, "plan-graph"))
}

// TestShowTerraformCmd tests the 'show-terraform' command execution
func TestShowTerraformCmd(t *testing.T) {
	// Create a temporary workspace for testing
	dir := t.TempDir()
	tfContent := `
terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}`
	err := os.WriteFile(dir+"/main.tf", []byte(tfContent), 0644)
	assert.NoError(t, err)

	// Redirect stdout to a buffer to capture output
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() { os.Stdout = oldOut }()

	// Execute the command
	rootCmd.SetArgs([]string{"show-terraform", dir})
	err = rootCmd.Execute()
	assert.NoError(t, err)

	// Stop writing to the pipe and read the output
	assert.NoError(t, w.Close())
	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	assert.NoError(t, err)

	// Check the output
	output := buf.String()
	assert.Contains(t, output, `"path":`)
	assert.Contains(t, output, `"required_version": ">= 1.0.0"`)
	assert.Contains(t, output, `"aws": "~> 5.0"`)
}

// TestExecute ensures the main execute function runs without errors
func TestExecute(t *testing.T) {
	// Redirect stdout to prevent printing to console during test
	oldOut := os.Stdout
	os.Stdout, _ = os.Create(os.DevNull)
	defer func() { os.Stdout = oldOut }()

	// We just want to ensure it runs without panicking.
	// We're not testing the full cobra execution logic here.
	assert.NotPanics(t, func() {
		// Need to reset the command path for testing context
		rootCmd.SetArgs([]string{})
		Execute()
	})
}

// findCommand is a helper to find a subcommand by its name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return cmd
		}
	}
	return nil
}

// TestRootCmdRun tests the output of the root command's Run function
func TestRootCmdRun(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	rootCmd.Run(rootCmd, []string{})

	assert.NoError(t, w.Close())
	os.Stdout = old

	var buf bytes.Buffer
	_, err := io.Copy(&buf, r)
	assert.NoError(t, err)
	output := buf.String()

	assert.True(t, strings.Contains(output, "Terraform Ops CLI"))
}
