#!/bin/sh -l

set -e

# GitHub Action inputs are available as environment variables with INPUT_ prefix
PLAN_FILE="${INPUT_PLAN_FILE}"
FORMAT="${INPUT_FORMAT:-graphviz}"
OUTPUT_FILE="${INPUT_OUTPUT_FILE}"
GROUP_BY="${INPUT_GROUP_BY:-module}"
NO_DATA_SOURCES="${INPUT_NO_DATA_SOURCES:-false}"
NO_OUTPUTS="${INPUT_NO_OUTPUTS:-false}"
NO_VARIABLES="${INPUT_NO_VARIABLES:-false}"
NO_LOCALS="${INPUT_NO_LOCALS:-false}"
NO_MODULES="${INPUT_NO_MODULES:-false}"
COMPACT="${INPUT_COMPACT:-false}"
VERBOSE="${INPUT_VERBOSE:-false}"

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
ARGS="plan-graph"
ARGS="${ARGS} --format ${FORMAT}"
ARGS="${ARGS} --group-by ${GROUP_BY}"

# Add boolean flags (only add --no-* flags when they are true)
if [ "${NO_DATA_SOURCES}" = "true" ]; then
	ARGS="${ARGS} --no-data-sources"
fi

if [ "${NO_OUTPUTS}" = "true" ]; then
	ARGS="${ARGS} --no-outputs"
fi

if [ "${NO_VARIABLES}" = "true" ]; then
	ARGS="${ARGS} --no-variables"
fi

if [ "${NO_LOCALS}" = "true" ]; then
	ARGS="${ARGS} --no-locals"
fi

if [ "${NO_MODULES}" = "true" ]; then
	ARGS="${ARGS} --no-modules"
fi

if [ "${COMPACT}" = "true" ]; then
	ARGS="${ARGS} --compact"
fi

if [ "${VERBOSE}" = "true" ]; then
	ARGS="${ARGS} --verbose"
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
	GRAPH_CONTENT=$(eval "/app/terraform-ops ${ARGS}")
	# Escape newlines for GitHub output
	GRAPH_CONTENT_ESCAPED=$(echo "${GRAPH_CONTENT}" | sed ':a;N;$!ba;s/\n/\\n/g')
	echo "graph-content=${GRAPH_CONTENT_ESCAPED}" >>"${GITHUB_OUTPUT}"
	# Also print to stdout for visibility
	echo "${GRAPH_CONTENT}"
fi

echo "Plan graph generation completed successfully"
