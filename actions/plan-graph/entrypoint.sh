#!/bin/sh -l

set -e

# GitHub Action inputs are available as environment variables with INPUT_ prefix
PLAN_FILE="${INPUT_PLAN-FILE}"
FORMAT="${INPUT_FORMAT:-graphviz}"
OUTPUT_FILE="${INPUT_OUTPUT-FILE}"
GROUP_BY="${INPUT_GROUP-BY:-module}"
SHOW_DEPENDENCIES="${INPUT_SHOW-DEPENDENCIES:-true}"
SHOW_SENSITIVITY="${INPUT_SHOW-SENSITIVITY:-false}"
SHOW_OUTPUTS="${INPUT_SHOW-OUTPUTS:-false}"
SHOW_VARIABLES="${INPUT_SHOW-VARIABLES:-false}"
SHOW_LOCALS="${INPUT_SHOW-LOCALS:-false}"
SHOW_DATA_SOURCES="${INPUT_SHOW-DATA-SOURCES:-false}"
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

# Add boolean flags
if [ "${SHOW_DEPENDENCIES}" = "false" ]; then
	ARGS="${ARGS} --show-dependencies=false"
fi

if [ "${SHOW_SENSITIVITY}" = "true" ]; then
	ARGS="${ARGS} --show-sensitivity"
fi

if [ "${SHOW_OUTPUTS}" = "true" ]; then
	ARGS="${ARGS} --show-outputs"
fi

if [ "${SHOW_VARIABLES}" = "true" ]; then
	ARGS="${ARGS} --show-variables"
fi

if [ "${SHOW_LOCALS}" = "true" ]; then
	ARGS="${ARGS} --show-locals"
fi

if [ "${SHOW_DATA_SOURCES}" = "true" ]; then
	ARGS="${ARGS} --show-data-sources"
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
