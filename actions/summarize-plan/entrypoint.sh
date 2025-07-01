#!/bin/sh -l

set -e

# GitHub Action inputs are available as environment variables with INPUT_ prefix
PLAN_FILE="${INPUT_PLAN_FILE}"
FORMAT="${INPUT_FORMAT:-text}"
OUTPUT_FILE="${INPUT_OUTPUT_FILE}"
GROUP_BY="${INPUT_GROUP_BY:-action}"
NO_SENSITIVE="${INPUT_NO_SENSITIVE:-false}"
COMPACT="${INPUT_COMPACT:-false}"
VERBOSE="${INPUT_VERBOSE:-false}"
SHOW_DETAILS="${INPUT_SHOW_DETAILS:-false}"
COLOR="${INPUT_COLOR:-auto}"

# Validate required inputs
if [ -z "${PLAN_FILE}" ]; then
	echo "Error: plan-file input is required"
	exit 1
fi

# Check if plan file exists
if [ ! -f "${PLAN_FILE}" ]; then
	echo "Error: Plan file '${PLAN_FILE}' not found"
	exit 1
fi

# Build command arguments
ARGS="summarize-plan"
ARGS="${ARGS} --format ${FORMAT}"
ARGS="${ARGS} --group-by ${GROUP_BY}"
ARGS="${ARGS} --color ${COLOR}"

# Add boolean flags (only add --no-* flags when they are true)
if [ "${NO_SENSITIVE}" = "true" ]; then
	ARGS="${ARGS} --no-sensitive"
fi

if [ "${COMPACT}" = "true" ]; then
	ARGS="${ARGS} --compact"
fi

if [ "${VERBOSE}" = "true" ]; then
	ARGS="${ARGS} --verbose"
fi

if [ "${SHOW_DETAILS}" = "true" ]; then
	ARGS="${ARGS} --show-details"
fi

# Add output file if specified
if [ -n "${OUTPUT_FILE}" ]; then
	ARGS="${ARGS} --output ${OUTPUT_FILE}"
fi

# Add plan file as the last argument
ARGS="${ARGS} ${PLAN_FILE}"

echo "Executing: /app/terraform-ops ${ARGS}"

# Execute the command
if [ -n "${OUTPUT_FILE}" ]; then
	# If output file is specified, execute and set output
	eval "/app/terraform-ops ${ARGS}"
	echo "output-file-path=${OUTPUT_FILE}" >>"${GITHUB_OUTPUT}"
else
	# If no output file, capture stdout and set as output
	SUMMARY_CONTENT=$(eval "/app/terraform-ops ${ARGS}")
	# Escape newlines for GitHub output
	SUMMARY_CONTENT_ESCAPED=$(echo "${SUMMARY_CONTENT}" | sed ':a;N;$!ba;s/\n/\\n/g')
	echo "summary-content=${SUMMARY_CONTENT_ESCAPED}" >>"${GITHUB_OUTPUT}"
	# Also print to stdout for visibility
	echo "${SUMMARY_CONTENT}"
fi

echo "Plan summary generation completed successfully"
