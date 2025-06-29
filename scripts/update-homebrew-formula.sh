#!/bin/bash

# Script to update Homebrew formula with new SHA256 hashes
# Usage: ./scripts/update-homebrew-formula.sh <version> <linux-sha256> <darwin-amd64-sha256> <darwin-arm64-sha256>

set -e

if [[ $# -ne 4 ]]; then
	echo "Usage: $0 <version> <linux-sha256> <darwin-amd64-sha256> <darwin-arm64-sha256>"
	echo "Example: $0 0.1.0 abc123... def456... ghi789..."
	exit 1
fi

VERSION=$1
LINUX_SHA256=$2
DARWIN_AMD64_SHA256=$3
DARWIN_ARM64_SHA256=$4

FORMULA_FILE="Formula/terraform-ops.rb"

if [[ ! -f ${FORMULA_FILE} ]]; then
	echo "Error: Formula file ${FORMULA_FILE} not found"
	exit 1
fi

echo "Updating Homebrew formula for version ${VERSION}..."

# Update version
sed -i.bak "s/version \"[^\"]*\"/version \"${VERSION}\"/" "${FORMULA_FILE}"

# Update SHA256 hashes
sed -i.bak "s/sha256 \"PLACEHOLDER_SHA256_LINUX\"/sha256 \"${LINUX_SHA256}\"/" "${FORMULA_FILE}"
sed -i.bak "s/sha256 \"PLACEHOLDER_SHA256_AMD64\"/sha256 \"${DARWIN_AMD64_SHA256}\"/" "${FORMULA_FILE}"
sed -i.bak "s/sha256 \"PLACEHOLDER_SHA256_ARM64\"/sha256 \"${DARWIN_ARM64_SHA256}\"/" "${FORMULA_FILE}"

# Remove backup files
rm -f "${FORMULA_FILE}.bak"

echo "Formula updated successfully!"
echo "Please review the changes and commit them:"
echo "  git add ${FORMULA_FILE}"
echo "  git commit -m \"Update Homebrew formula for v${VERSION}\""
