# Integration test Makefile for analyze

.PHONY: all build test generate-plans clean-plans

# Build the binary in the expected location.
build:
	cd ../.. && go build -o integration_tests/build/terraform-ops ./cmd/terraform-ops

# Generate plan JSON with the Terraform version selected by the CI matrix.
generate-plans:
	cd workspaces/simple && terraform init -input=false
	cd workspaces/simple && terraform plan -out=plan.tfplan -input=false
	cd workspaces/simple && terraform show -json plan.tfplan > plan.json

clean-plans:
	rm -f workspaces/simple/plan.tfplan
	rm -f workspaces/simple/plan.json
	rm -rf workspaces/simple/.terraform
	rm -f workspaces/simple/.terraform.lock.hcl

# The generated plan is deliberately retained while the Go test executes.
test:
	go test -v ./...

all: build generate-plans test
