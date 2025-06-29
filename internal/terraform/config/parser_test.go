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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	parser := NewParser()
	assert.NotNil(t, parser)
}

func TestParseConfigFiles_SimpleProviders(t *testing.T) {
	// Create a temporary directory with a simple Terraform configuration
	tmpDir := t.TempDir()

	mainTf := `terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    google = {
      source  = "hashicorp/google"
      version = ">= 4.83.0,< 5.0.0"
    }
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, tmpDir, config.Path)
	assert.Equal(t, ">= 1.0.0", config.RequiredVersion)
	assert.Equal(t, map[string]string{
		"aws":    "~> 5.0",
		"google": ">= 4.83.0,< 5.0.0",
	}, config.RequiredProviders)
	assert.Nil(t, config.Backend)
}

func TestParseConfigFiles_WithBackend(t *testing.T) {
	// Create a temporary directory with a Terraform configuration including backend
	tmpDir := t.TempDir()

	mainTf := `terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = "~> 5.0"
  }

  backend "s3" {
    bucket = "terraform-state-prod"
    key    = "terraform/state.tfstate"
    region = "us-west-2"
    encrypt = true
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, ">= 1.0.0", config.RequiredVersion)
	assert.Equal(t, map[string]string{"aws": "~> 5.0"}, config.RequiredProviders)
	assert.NotNil(t, config.Backend)
	assert.Equal(t, "s3", config.Backend.Type)
	assert.Equal(t, map[string]string{
		"bucket":  "terraform-state-prod",
		"key":     "terraform/state.tfstate",
		"region":  "us-west-2",
		"encrypt": "true",
	}, config.Backend.Config)
}

func TestParseConfigFiles_MultipleFiles(t *testing.T) {
	// Create a temporary directory with multiple Terraform files
	tmpDir := t.TempDir()

	mainTf := `terraform {
  required_version = ">= 1.0.0"
  required_providers {
    aws = "~> 5.0"
  }
}`

	providersTf := `terraform {
  required_providers {
    google = ">= 4.83.0,< 5.0.0"
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTf), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "providers.tf"), []byte(providersTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, ">= 1.0.0", config.RequiredVersion)
	// The second file should override/add to the providers
	assert.Equal(t, map[string]string{
		"aws":    "~> 5.0",
		"google": ">= 4.83.0,< 5.0.0",
	}, config.RequiredProviders)
}

func TestParseConfigFiles_NoTerraformBlock(t *testing.T) {
	// Create a temporary directory with a file that doesn't contain a terraform block
	tmpDir := t.TempDir()

	resourceTf := `resource "aws_instance" "web" {
  ami           = "ami-12345678"
  instance_type = "t3.micro"
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(resourceTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, "", config.RequiredVersion)
	assert.Empty(t, config.RequiredProviders)
	assert.Nil(t, config.Backend)
}

func TestParseConfigFiles_InvalidHCL(t *testing.T) {
	// Create a temporary directory with invalid HCL
	tmpDir := t.TempDir()

	invalidTf := `terraform {
  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = "~> 5.0"
    }
  }

  backend "s3" {
    bucket = "terraform-state-prod"
    key    = "terraform/state.tfstate"
    region = "us-west-2"
    encrypt = true
  }

  invalid_syntax {
    this is not valid HCL
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(invalidTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	// Should still work as we're using PartialContent which ignores invalid blocks
	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Empty(t, config.RequiredProviders)
	assert.Nil(t, config.Backend)
}

func TestParseConfigFiles_DirectoryNotFound(t *testing.T) {
	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{"nonexistent_directory"})

	assert.Error(t, err)
	assert.Nil(t, configs)
	assert.Contains(t, err.Error(), "path does not exist or is not a directory")
}

func TestParseConfigFiles_FileInsteadOfDirectory(t *testing.T) {
	// Create a temporary file (not directory)
	tmpFile, err := os.CreateTemp("", "test_*.tf")
	assert.NoError(t, err)
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil && !os.IsNotExist(err) {
			t.Logf("Failed to remove temp file: %v", err)
		}
	}()

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpFile.Name()})

	assert.Error(t, err)
	assert.Nil(t, configs)
	assert.Contains(t, err.Error(), "path does not exist or is not a directory")
}

func TestParseConfigFiles_EmptyDirectory(t *testing.T) {
	// Create an empty temporary directory
	tmpDir := t.TempDir()

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, "", config.RequiredVersion)
	assert.Empty(t, config.RequiredProviders)
	assert.Nil(t, config.Backend)
}

func TestParseConfigFiles_MultipleDirectories(t *testing.T) {
	// Create two temporary directories with different configurations
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	config1 := `terraform {
  required_providers {
    aws = "~> 5.0"
  }
}`

	config2 := `terraform {
  required_providers {
    google = ">= 4.83.0,< 5.0.0"
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir1, "main.tf"), []byte(config1), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir2, "main.tf"), []byte(config2), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir1, tmpDir2})

	assert.NoError(t, err)
	assert.Len(t, configs, 2)

	// First config
	assert.Equal(t, map[string]string{"aws": "~> 5.0"}, configs[0].RequiredProviders)

	// Second config
	assert.Equal(t, map[string]string{"google": ">= 4.83.0,< 5.0.0"}, configs[1].RequiredProviders)
}

func TestParseConfigFiles_ComplexBackend(t *testing.T) {
	// Create a temporary directory with a complex backend configuration
	tmpDir := t.TempDir()

	mainTf := `terraform {
  required_version = ">= 1.0.0"
  required_providers {
    google = ">= 4.83.0,< 5.0.0"
  }

  backend "gcs" {
    bucket                      = "terraform-state-prod"
    prefix                      = "terraform/state"
    impersonate_service_account = "test-service-account@terraform-ops-test.iam.gserviceaccount.com"
  }
}`

	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(mainTf), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	configs, err := parser.ParseConfigFiles([]string{tmpDir})

	assert.NoError(t, err)
	assert.Len(t, configs, 1)

	config := configs[0]
	assert.Equal(t, ">= 1.0.0", config.RequiredVersion)
	assert.Equal(t, map[string]string{"google": ">= 4.83.0,< 5.0.0"}, config.RequiredProviders)
	assert.NotNil(t, config.Backend)
	assert.Equal(t, "gcs", config.Backend.Type)
	assert.Equal(t, map[string]string{
		"bucket":                      "terraform-state-prod",
		"prefix":                      "terraform/state",
		"impersonate_service_account": "test-service-account@terraform-ops-test.iam.gserviceaccount.com",
	}, config.Backend.Config)
}

func TestFindTerraformFiles(t *testing.T) {
	// Create a temporary directory with various files
	tmpDir := t.TempDir()

	// Create .tf files
	err := os.WriteFile(filepath.Join(tmpDir, "main.tf"), []byte(""), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "variables.tf"), []byte(""), 0644)
	assert.NoError(t, err)

	// Create non-.tf files
	err = os.WriteFile(filepath.Join(tmpDir, "main.txt"), []byte(""), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "variables.hcl"), []byte(""), 0644)
	assert.NoError(t, err)

	// Create a subdirectory (should be ignored)
	subDir := filepath.Join(tmpDir, "subdir")
	err = os.Mkdir(subDir, 0755)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir, "main.tf"), []byte(""), 0644)
	assert.NoError(t, err)

	parser := NewParser()
	files, err := parser.findTerraformFiles(tmpDir)

	assert.NoError(t, err)
	assert.Len(t, files, 2)

	// Check that only .tf files are found
	for _, file := range files {
		assert.True(t, filepath.Ext(file) == ".tf")
	}
}
