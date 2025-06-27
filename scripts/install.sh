#!/bin/bash
set -e

# Installation script for terraform-ops CLI tool

BINARY_NAME="terraform-ops"
INSTALL_DIR="/usr/local/bin"
BUILD_DIR="build"

# Check if binary exists
if [[ ! -f "${BUILD_DIR}/${BINARY_NAME}" ]]; then
	echo "Binary not found. Please run 'make build' first."
	exit 1
fi

# Install binary
echo "Installing ${BINARY_NAME} to ${INSTALL_DIR}..."
sudo cp "${BUILD_DIR}/${BINARY_NAME}" "${INSTALL_DIR}/"
sudo chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

echo "${BINARY_NAME} installed successfully!"
echo "You can now run '${BINARY_NAME}' from anywhere."
