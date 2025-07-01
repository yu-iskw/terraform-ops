# Integration test Makefile for summarize_plan

.PHONY: all build test generate-plans clean-plans

# Build the binary in the expected location
build:
	cd ../.. && go build -o integration_tests/build/terraform-ops ./cmd/terraform-ops

# Generate plan.json files using shell commands (not Go)
generate-plans:
	cd workspaces/simple && terraform init -input=false
	cd workspaces/simple && terraform plan -out=plan.tfplan -input=false
	cd workspaces/simple && terraform show -json plan.tfplan > plan.json

# Clean generated plan files
clean-plans:
	rm -f workspaces/simple/plan.tfplan
	rm -f workspaces/simple/plan.json

# Run the tests in this directory
# (Assumes the binary is already built and plan.json files exist)
test:
	go test -v ./...

# Default target: build, generate plans, and test
all: build generate-plans test
