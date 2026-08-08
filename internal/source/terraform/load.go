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

import "github.com/yu/terraform-ops/internal/ir"

// LoadFile parses a Terraform/OpenTofu-compatible plan JSON file and immediately
// normalizes it across the sanitization boundary. Command/application code should
// consume the returned ChangeSet rather than the source Plan DTO.
func LoadFile(path string, maxBytes int64, engine ir.Engine, mode ir.RedactionMode) (*ir.ChangeSet, error) {
	plan, err := ParseFile(path, maxBytes)
	if err != nil {
		return nil, err
	}
	return Normalize(plan, engine, mode)
}
