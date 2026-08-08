#!/bin/sh

set -eu

ENGINE_BIN=${ENGINE_BIN:?ENGINE_BIN is required}
OUTPUT_FILE=${OUTPUT_FILE:?OUTPUT_FILE is required}

case "$OUTPUT_FILE" in
	""|*/*|..|.*)
		echo "invalid OUTPUT_FILE: $OUTPUT_FILE" >&2
		exit 2
		;;
esac

workdir=/tmp/terraform-ops-compat
rm -rf "$workdir"
mkdir -p "$workdir"
cp -R /workspace/. "$workdir/"

"$ENGINE_BIN" -chdir="$workdir" init -backend=false -input=false -no-color
"$ENGINE_BIN" -chdir="$workdir" validate -no-color
"$ENGINE_BIN" -chdir="$workdir" plan -input=false -lock=false -refresh=false -no-color -out=/tmp/plan.bin
"$ENGINE_BIN" -chdir="$workdir" show -json /tmp/plan.bin >"/artifacts/$OUTPUT_FILE"

# Prove that the artifact is a non-empty JSON plan before the container exits.
test -s "/artifacts/$OUTPUT_FILE"
grep -q '"format_version"' "/artifacts/$OUTPUT_FILE"
