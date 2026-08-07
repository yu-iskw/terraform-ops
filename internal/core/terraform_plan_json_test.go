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

import (
	"encoding/json"
	"testing"
)

func TestTerraformPlanUnmarshalCurrentApplyableField(t *testing.T) {
	var plan TerraformPlan
	if err := json.Unmarshal([]byte(`{"format_version":"1.0","applyable":true}`), &plan); err != nil {
		t.Fatal(err)
	}
	if !plan.Applicable {
		t.Fatal("expected current applyable field to populate legacy Applicable property")
	}
}

func TestTerraformPlanUnmarshalLegacyApplicableField(t *testing.T) {
	var plan TerraformPlan
	if err := json.Unmarshal([]byte(`{"format_version":"1.0","applicable":true}`), &plan); err != nil {
		t.Fatal(err)
	}
	if !plan.Applicable {
		t.Fatal("expected legacy applicable field to remain supported during migration")
	}
}
