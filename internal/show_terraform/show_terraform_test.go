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

package show_terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetTerraformInfo(t *testing.T) {
	// Create temporary workspace directory
	dir, err := os.MkdirTemp("", "tfws_")
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, os.RemoveAll(dir))
	}()

	tfContent := `
terraform {
  required_version = ">= 1.5.0, < 2.0.0"
  backend "local" {
    path = "state/terraform.tfstate"
  }
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 4.0"
    }
    random = {
      source  = "hashicorp/random"
    }
  }
}
`
	err = os.WriteFile(filepath.Join(dir, "main.tf"), []byte(tfContent), 0644)
	assert.NoError(t, err)

	infos, err := GetTerraformInfo([]string{dir})
	assert.NoError(t, err)
	assert.Len(t, infos, 1)

	info := infos[0]
	assert.Equal(t, dir, info.Path)
	assert.Equal(t, ">= 1.5.0, < 2.0.0", info.Terraform.RequiredVersion)
	if assert.NotNil(t, info.Terraform.Backend) {
		assert.Equal(t, "local", info.Terraform.Backend.Type)
		assert.Equal(t, map[string]string{"path": "state/terraform.tfstate"}, info.Terraform.Backend.Config)
	}
	assert.Equal(t, map[string]string{
		"aws":    "~> 4.0",
		"random": "",
	}, info.Terraform.RequiredProviders)
}

func TestBackendParsing(t *testing.T) {
	testCases := []struct {
		name            string
		tfContent       string
		expectedBackend *Backend
	}{
		{
			name: "GCS Backend",
			tfContent: `
terraform {
  backend "gcs" {
    bucket  = "my-gcs-bucket"
    prefix  = "terraform/state"
  }
}
`,
			expectedBackend: &Backend{
				Type: "gcs",
				Config: map[string]string{
					"bucket": "my-gcs-bucket",
					"prefix": "terraform/state",
				},
			},
		},
		{
			name: "S3 Backend with bool",
			tfContent: `
terraform {
  backend "s3" {
    bucket   = "my-s3-bucket"
    key      = "path/to/state.tfstate"
    region   = "us-west-2"
    encrypt  = true
  }
}
`,
			expectedBackend: &Backend{
				Type: "s3",
				Config: map[string]string{
					"bucket":  "my-s3-bucket",
					"key":     "path/to/state.tfstate",
					"region":  "us-west-2",
					"encrypt": "true",
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dir, err := os.MkdirTemp("", "tfws_")
			assert.NoError(t, err)
			defer func() {
				assert.NoError(t, os.RemoveAll(dir))
			}()

			err = os.WriteFile(filepath.Join(dir, "main.tf"), []byte(tc.tfContent), 0644)
			assert.NoError(t, err)

			infos, err := GetTerraformInfo([]string{dir})
			assert.NoError(t, err)
			assert.Len(t, infos, 1)

			info := infos[0]
			if tc.expectedBackend == nil {
				assert.Nil(t, info.Terraform.Backend)
			} else {
				if assert.NotNil(t, info.Terraform.Backend) {
					assert.Equal(t, tc.expectedBackend.Type, info.Terraform.Backend.Type)
					assert.Equal(t, tc.expectedBackend.Config, info.Terraform.Backend.Config)
				}
			}
		})
	}
}
