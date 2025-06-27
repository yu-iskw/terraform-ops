.PHONY: build clean test coverage fmt vet lint install run help

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

# Run tests
test: build
	$(GOTEST) -v ./...

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
