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

TERRAFORM_OPS_PATH=../../build/terraform-ops

# Default target - generate plan and JSON for web-app
all: web-app-plan web-app-plan-json

# Test both workspaces
test: web-app-test simple-random-test

# Web-app workspace targets
web-app-plan:
	cd workspaces/web-app && terraform init
	cd workspaces/web-app && terraform plan --out=../web-app-plan.tfplan

web-app-plan-json: web-app-plan
	cd workspaces/web-app && terraform show -json ../web-app-plan.tfplan > ../../web-app-plan.json

# Simple-random workspace targets
simple-random-plan:
	cd workspaces/simple-random && terraform init
	cd workspaces/simple-random && terraform plan --out=../simple-random-plan.tfplan

simple-random-plan-json: simple-random-plan
	cd workspaces/simple-random && terraform show -json ../simple-random-plan.tfplan > ../../simple-random-plan.json

# Generate all plans
all-plans: web-app-plan-json simple-random-plan-json

# Clean generated files
clean:
	rm -f web-app-plan.tfplan
	rm -f web-app-plan.json
	rm -f simple-random-plan.tfplan
	rm -f simple-random-plan.json
	rm -f workspaces/web-app-plan.json
	rm -f workspaces/simple-random-plan.json
	rm -f web-app-graph.dot
	rm -f web-app-graph.md
	rm -f web-app-graph.puml
	rm -f simple-random-graph.dot
	rm -f simple-random-graph.md
	rm -f simple-random-graph.puml
	rm -rf output/

# Clean all including terraform state
clean-all: clean
	cd workspaces/web-app && terraform init || true
	cd workspaces/web-app && terraform destroy -auto-approve || true
	cd workspaces/web-app && rm -rf .terraform .terraform.lock.hcl
	cd workspaces/simple-random && terraform init || true
	cd workspaces/simple-random && terraform destroy -auto-approve || true
	cd workspaces/simple-random && rm -rf .terraform .terraform.lock.hcl

# Generate graphs and save to files for web-app
web-app-graph-graphviz: web-app-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=graphviz > output/web-app-graph.dot

web-app-graph-mermaid: web-app-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=mermaid > output/web-app-graph.md

web-app-graph-plantuml: web-app-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=plantuml > output/web-app-graph.puml

# Generate graphs and save to files for simple-random
simple-random-graph-graphviz: simple-random-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=graphviz > output/simple-random-graph.dot

simple-random-graph-mermaid: simple-random-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=mermaid > output/simple-random-graph.md

simple-random-graph-plantuml: simple-random-plan-json
	@mkdir -p output
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=plantuml > output/simple-random-graph.puml

# Generate all graph formats
web-app-graphs: clean web-app-graph-graphviz web-app-graph-mermaid web-app-graph-plantuml
simple-random-graphs: clean simple-random-graph-graphviz simple-random-graph-mermaid simple-random-graph-plantuml

# Test the plan-graph command with generated plans
web-app-test: web-app-plan-json
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=graphviz
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=mermaid
	$(TERRAFORM_OPS_PATH) plan-graph web-app-plan.json --format=plantuml

simple-random-test: simple-random-plan-json
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=graphviz
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=mermaid
	$(TERRAFORM_OPS_PATH) plan-graph simple-random-plan.json --format=plantuml

.PHONY: all web-app-plan web-app-plan-json simple-random-plan simple-random-plan-json all-plans clean clean-all web-app-graph-graphviz web-app-graph-mermaid web-app-graph-plantuml simple-random-graph-graphviz simple-random-graph-mermaid simple-random-graph-plantuml web-app-graphs simple-random-graphs web-app-test simple-random-test test
