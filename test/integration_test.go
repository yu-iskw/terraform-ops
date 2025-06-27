package test

import (
	"os/exec"
	"testing"
)

func TestCLIVersion(t *testing.T) {
	// Build the CLI first
	buildCmd := exec.Command("make", "build")
	buildCmd.Dir = "/Users/yu/local/src/github/terraform-ops" // Explicitly set working directory
	buildOutput, buildErr := buildCmd.CombinedOutput()
	if buildErr != nil {
		t.Fatalf("Failed to build CLI: %v\nOutput: %s", buildErr, string(buildOutput))
	}
	t.Logf("Build output: %s", string(buildOutput))

	// Test running the CLI
	cmd := exec.Command("../build/terraform-ops")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI execution failed: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected output from CLI, got empty response")
	}

	t.Logf("CLI output: %s", string(output))
}

func TestCLIHelp(t *testing.T) {
	cmd := exec.Command("../build/terraform-ops", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI help command failed: %v\nOutput: %s", err, string(output))
	}

	if len(output) == 0 {
		t.Error("Expected help output from CLI, got empty response")
	}

	t.Logf("CLI help output: %s", string(output))
}
