# Copyright 2025 yu-iskw
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

.PHONY: build clean test coverage fmt vet lint install run help test-integration homebrew-test homebrew-install homebrew-uninstall

# Build variables
BINARY_NAME=terraform-ops
BUILD_DIR=build
MAIN_PATH=./cmd/terraform-ops

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build the binary
build:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

# Run unit tests
test: build
	$(GOTEST) -v ./internal/...

# Run integration tests
test-integration: build test-show-terraform test-integration-plan-graph test-summarize-plan

test-show-terraform:
	$(MAKE) -C integration_tests/show_terraform all

test-integration-plan-graph:
	$(MAKE) -C integration_tests/plan_graph all

test-summarize-plan:
	$(MAKE) -C integration_tests/summarize_plan all

# Run tests with coverage
coverage:
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Format code
format:
	trunk fmt

format-all:
	trunk fmt -a

# Vet code
vet:
	$(GOVET) ./...

# Lint code (requires golangci-lint)
lint: format
	trunk check

lint-all: format-all
	trunk check -a

# Install dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Install the binary to $GOPATH/bin
install:
	$(GOCMD) install $(MAIN_PATH)

# Run the application
run:
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH) && ./$(BUILD_DIR)/$(BINARY_NAME)

# Development run with live reload (requires air)
dev:
	air

# Homebrew-related targets
homebrew-test:
	brew install --formula Formula/terraform-ops.rb
	brew test terraform-ops

homebrew-install:
	brew install --formula Formula/terraform-ops.rb

homebrew-uninstall:
	brew uninstall terraform-ops || true
