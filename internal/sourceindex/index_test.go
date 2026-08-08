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

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildResolvesRootAndLocalModuleResources(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "modules", "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "main.tf"), `resource "terraform_data" "root" {
  input = "x"
}

module "child" {
  source = "./modules/child"
}
`)
	writeTestFile(t, filepath.Join(child, "main.tf"), `data "terraform_remote_state" "example" {
  backend = "local"
}

resource "terraform_data" "inner" {
  input = "y"
}
`)

	idx, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}

	rootLocation, ok := idx.Resolve(`terraform_data.root[0]`)
	if !ok {
		t.Fatal("root resource was not indexed")
	}
	if rootLocation.Path != "main.tf" || rootLocation.StartLine != 1 {
		t.Fatalf("root location = %#v", rootLocation)
	}

	childLocation, ok := idx.Resolve(`module.child["blue"].terraform_data.inner[0]`)
	if !ok {
		t.Fatal("child resource was not indexed")
	}
	if childLocation.Path != "modules/child/main.tf" || childLocation.StartLine != 5 {
		t.Fatalf("child location = %#v", childLocation)
	}

	dataLocation, ok := idx.Resolve(`module.child.data.terraform_remote_state.example`)
	if !ok || dataLocation.StartLine != 1 {
		t.Fatalf("data location = %#v, ok=%v", dataLocation, ok)
	}
}

func TestBuildWithArtifactRootEmitsRepositoryRelativePaths(t *testing.T) {
	repositoryRoot := t.TempDir()
	workspaceRoot := filepath.Join(repositoryRoot, "infra", "prod")
	if err := os.MkdirAll(workspaceRoot, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(workspaceRoot, "main.tf"), `resource "terraform_data" "example" {}`)

	idx, err := BuildWithArtifactRoot(workspaceRoot, repositoryRoot)
	if err != nil {
		t.Fatal(err)
	}
	location, ok := idx.Resolve("terraform_data.example")
	if !ok {
		t.Fatal("resource was not indexed")
	}
	if location.Path != "infra/prod/main.tf" {
		t.Fatalf("location path = %q, want infra/prod/main.tf", location.Path)
	}
}

func TestBuildWithArtifactRootRejectsWorkspaceOutsideArtifactRoot(t *testing.T) {
	artifactRoot := t.TempDir()
	workspaceRoot := t.TempDir()
	if _, err := BuildWithArtifactRoot(workspaceRoot, artifactRoot); err == nil {
		t.Fatal("expected workspace/artifact containment error")
	}
}

func TestBuildDoesNotTraverseModuleOutsideWorkspace(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "root")
	external := filepath.Join(parent, "external")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(external, 0o755); err != nil {
		t.Fatal(err)
	}
	writeTestFile(t, filepath.Join(root, "main.tf"), `module "external" {
  source = "../external"
}
`)
	writeTestFile(t, filepath.Join(external, "main.tf"), `resource "terraform_data" "outside" {}`)

	idx, err := Build(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := idx.Resolve("module.external.terraform_data.outside"); ok {
		t.Fatal("module outside workspace root must not be indexed")
	}
}

func TestStaticAddressHandlesQuotedInstanceKeys(t *testing.T) {
	got := staticAddress(`module.child["a]b"].terraform_data.example["x]y"]`)
	want := "module.child.terraform_data.example"
	if got != want {
		t.Fatalf("staticAddress() = %q, want %q", got, want)
	}
}

func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
