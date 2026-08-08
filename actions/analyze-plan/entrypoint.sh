#!/bin/sh

set -eu

PLAN_FILE=${INPUT_PLAN_FILE:-}
WORKSPACE_ROOT=${INPUT_WORKSPACE_ROOT:-.}
ENGINE=${INPUT_ENGINE:-auto}
REDACTION=${INPUT_REDACTION:-strict}
FAIL_ON=${INPUT_FAIL_ON:-none}
OUTPUT_DIRECTORY=${INPUT_OUTPUT_DIRECTORY:-.terraform-ops}
WRITE_JOB_SUMMARY=${INPUT_WRITE_JOB_SUMMARY:-true}
GITHUB_WORKSPACE=${GITHUB_WORKSPACE:-/github/workspace}

if [ -z "$PLAN_FILE" ]; then
	echo "Error: plan-file input is required" >&2
	exit 2
fi

workspace=$(realpath "$GITHUB_WORKSPACE")
cd "$workspace"

resolve_existing_in_workspace() {
	candidate=$1
	if [ "${candidate#/}" != "$candidate" ]; then
		resolved=$(realpath "$candidate")
	else
		resolved=$(realpath "$workspace/$candidate")
	fi
	case "$resolved" in
		"$workspace"|"$workspace"/*) printf '%s\n' "$resolved" ;;
		*)
			echo "Error: path escapes GITHUB_WORKSPACE: $candidate" >&2
			exit 2
			;;
	esac
}

plan_path=$(resolve_existing_in_workspace "$PLAN_FILE")
workspace_root=$(resolve_existing_in_workspace "$WORKSPACE_ROOT")

if [ ! -f "$plan_path" ]; then
	echo "Error: plan file not found: $PLAN_FILE" >&2
	exit 2
fi
if [ ! -d "$workspace_root" ]; then
	echo "Error: workspace-root is not a directory: $WORKSPACE_ROOT" >&2
	exit 2
fi

case "$OUTPUT_DIRECTORY" in
	""|/*|..|../*|*/../*|*/..)
		echo "Error: output-directory must be a repository-relative path without '..': $OUTPUT_DIRECTORY" >&2
		exit 2
		;;
esac

output_dir="$workspace/$OUTPUT_DIRECTORY"
mkdir -p "$output_dir"
output_dir=$(realpath "$output_dir")
case "$output_dir" in
	"$workspace"|"$workspace"/*) ;;
	*)
		echo "Error: output-directory escapes GITHUB_WORKSPACE" >&2
		exit 2
		;;
esac

markdown_rel="$OUTPUT_DIRECTORY/analysis.md"
json_rel="$OUTPUT_DIRECTORY/analysis.json"
sarif_rel="$OUTPUT_DIRECTORY/analysis.sarif"
markdown_file="$workspace/$markdown_rel"
json_file="$workspace/$json_rel"
sarif_file="$workspace/$sarif_rel"

/app/terraform-ops analyze "$plan_path" \
	--engine "$ENGINE" \
	--redaction "$REDACTION" \
	--fail-on none \
	--format markdown \
	--output "$markdown_file"

/app/terraform-ops analyze "$plan_path" \
	--engine "$ENGINE" \
	--redaction "$REDACTION" \
	--fail-on none \
	--format json \
	--output "$json_file"

set +e
/app/terraform-ops analyze "$plan_path" \
	--engine "$ENGINE" \
	--redaction "$REDACTION" \
	--fail-on "$FAIL_ON" \
	--format sarif \
	--workspace-root "$workspace_root" \
	--output "$sarif_file"
status=$?
set -e

if [ "$WRITE_JOB_SUMMARY" = "true" ] && [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
	cat "$markdown_file" >>"$GITHUB_STEP_SUMMARY"
fi

if [ -n "${GITHUB_OUTPUT:-}" ]; then
	printf 'sarif-file=%s\n' "$sarif_rel" >>"$GITHUB_OUTPUT"
	printf 'json-file=%s\n' "$json_rel" >>"$GITHUB_OUTPUT"
	printf 'markdown-file=%s\n' "$markdown_rel" >>"$GITHUB_OUTPUT"
fi

exit "$status"
