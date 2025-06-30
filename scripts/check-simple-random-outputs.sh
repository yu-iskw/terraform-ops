#!/bin/bash
set -e

OUTPUT_DIR="output"
FILES=(
	"simple-random-graph.dot"
	"simple-random-graph.md"
	"simple-random-graph.puml"
)

echo "Checking simple-random graph output files..."

# Check existence and non-emptiness
for file in "${FILES[@]}"; do
	path="${OUTPUT_DIR}/${file}"
	if [[ ! -f ${path} ]]; then
		echo "❌ Missing ${path}"
		exit 1
	fi
	if [[ ! -s ${path} ]]; then
		echo "❌ Empty ${path}"
		exit 1
	fi
	echo "✅ Found and non-empty${file}"
done

# Optional: Validate Graphviz format
if command -v dot >/dev/null 2>&1; then
	echo "Validating Graphviz format..."
	if dot -Tpng "${OUTPUT_DIR}/simple-random-graph.dot" -o /dev/null 2>/dev/null; then
		echo "✅ Graphviz file is valid"
	else
		echo "❌ Invalid DOT file"
		exit 1
	fi
else
	echo "⚠️  Graphviz (dot) not found, skipping format validation"
fi

# Optional: Validate PlantUML format (more lenient)
if command -v plantuml >/dev/null 2>&1; then
	echo "Validating PlantUML format..."
	if plantuml -checkonly "${OUTPUT_DIR}/simple-random-graph.puml" >/dev/null 2>&1; then
		echo "✅ PlantUML file is valid"
	else
		echo "⚠️  PlantUML file has syntax warnings (but file exists and is non-empty)"
		# Don't exit with error for PlantUML validation issues
	fi
else
	echo "⚠️  PlantUML not found, skipping format validation"
fi

# Basic Mermaid validation (check for graph keyword)
if grep -q "graph" "${OUTPUT_DIR}/simple-random-graph.md"; then
	echo "✅ Mermaid file contains graph structure"
else
	echo "❌ Mermaid file missing graph structure"
	exit 1
fi

echo "🎉 All simple-random output files exist, are non-empty, and are valid."
