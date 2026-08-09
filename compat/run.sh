#!/bin/sh
# shellcheck disable=SC2250

set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
compose_file="$repo_root/compat/compose.yaml"
artifact_dir="$repo_root/compat/artifacts"

terraform_versions=${TERRAFORM_VERSIONS:-"1.7.0 1.8.0 1.9.0 1.10.0 1.11.0 1.12.0 1.15.8"}
opentofu_versions=${OPENTOFU_VERSIONS:-"1.10.10 1.11.6 1.12.5"}

mkdir -p "$artifact_dir"
chmod 0777 "$artifact_dir"
rm -f "$artifact_dir"/*.plan.json

for version in $terraform_versions; do
	output="terraform-$version.plan.json"
	echo "==> Terraform $version"
	TERRAFORM_VERSION="$version" OUTPUT_FILE="$output" docker compose -f "$compose_file" build terraform
	TERRAFORM_VERSION="$version" OUTPUT_FILE="$output" docker compose -f "$compose_file" run --rm --no-deps terraform
	test -s "$artifact_dir/$output"
done

for version in $opentofu_versions; do
	output="opentofu-$version.plan.json"
	echo "==> OpenTofu $version"
	OPENTOFU_VERSION="$version" OUTPUT_FILE="$output" docker compose -f "$compose_file" build opentofu
	OPENTOFU_VERSION="$version" OUTPUT_FILE="$output" docker compose -f "$compose_file" run --rm --no-deps opentofu
	test -s "$artifact_dir/$output"
done

COMPAT_ARTIFACT_DIR="$artifact_dir" go test ./integration_tests/compatibility -run '^TestSemanticEquivalence$' -count=1 -v
