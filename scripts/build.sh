#!/bin/bash
set -e

# Build configuration
BINARY_NAME="terraform-ops"
BUILD_DIR="build"
MAIN_PATH="./cmd/terraform-ops"

# Create build directory
mkdir -p "${BUILD_DIR}"

# Build for current platform
echo "Building ${BINARY_NAME}"
go build -o "${BUILD_DIR}"/"${BINARY_NAME}" "${MAIN_PATH}"

echo "Build completed: ${BUILD_DIR}/${BINARY_NAME}"

# Make binary executable
chmod +x "${BUILD_DIR}"/"${BINARY_NAME}"

echo "Binary is ready to use!"
