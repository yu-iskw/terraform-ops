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

package core

import "encoding/json"

// UnmarshalJSON keeps the legacy TerraformPlan domain compatible with current
// Terraform/OpenTofu plan JSON. The machine-readable plan contract calls this
// field "applyable"; older terraform-ops tests and fixtures used "applicable".
// Accept both during the legacy-command migration and prefer the source-contract
// spelling when both are present.
func (p *TerraformPlan) UnmarshalJSON(data []byte) error {
	type alias TerraformPlan
	decoded := struct {
		*alias
		Applyable *bool `json:"applyable"`
	}{
		alias: (*alias)(p),
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.Applyable != nil {
		p.Applicable = *decoded.Applyable
	}
	return nil
}
