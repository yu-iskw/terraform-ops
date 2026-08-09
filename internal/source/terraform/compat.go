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

import (
	"strconv"
	"strings"
)

func planComplete(plan *Plan) bool {
	if plan.Complete != nil {
		return *plan.Complete
	}

	// Current OpenTofu JSON plans omit Terraform's applyable/complete metadata.
	// The absence of both fields therefore means there is no producer-level
	// deferred-plan signal to preserve. Treat a successfully produced plan as
	// complete while still propagating the explicit errored flag.
	if !plan.applyablePresent {
		return !plan.Errored
	}

	major, minor, ok := parseMajorMinorVersion(plan.TerraformVersion)
	if !ok {
		// Missing completeness metadata is ambiguous for unknown/current
		// Terraform-compatible producers. Stay conservative unless the producer
		// version proves that the field predates Terraform 1.8.
		return false
	}
	return major < 1 || (major == 1 && minor < 8)
}

func parseMajorMinorVersion(version string) (int, int, bool) {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	if version == "" {
		return 0, 0, false
	}
	if core, _, ok := strings.Cut(version, "-"); ok {
		version = core
	}
	parts := strings.Split(version, ".")
	if len(parts) < 2 {
		return 0, 0, false
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}
	return major, minor, true
}
