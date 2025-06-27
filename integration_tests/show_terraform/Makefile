# Integration test Makefile for show_terraform

.PHONY: all build test

# Build the binary in the expected location
build:
	cd ../.. && go build -o integration_tests/build/terraform-ops ./cmd/terraform-ops

# Run the tests in this directory
# (Assumes the binary is already built)
test:
	go test -v ./...

# Default target: build and test
all: build test
