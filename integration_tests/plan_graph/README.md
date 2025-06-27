# Plan-Graph Integration Tests

This directory contains comprehensive integration tests for the `plan-graph` command of the `terraform-ops` tool.

## Overview

The integration tests validate that the `plan-graph` command can:

1. Generate graphs in all supported formats (Graphviz, Mermaid, PlantUML)
2. Handle different grouping strategies (module, action, resource_type)
3. Process various command line options (compact, verbose, show-dependencies, etc.)
4. Handle error cases gracefully
5. Generate output that can be processed by visualization tools

## Test Structure

### Test Files

- `integration_test.go` - Main integration test suite
- `workspaces/web-app/` - Complex Terraform workspace with web application infrastructure
- `workspaces/simple-random/` - Simple Terraform workspace with only random resources
- `Makefile` - Build and test automation

### Dynamic Plan Generation

The tests now use **dynamic plan generation** instead of static plan files:

1. **Terraform Configuration**: The `workspaces/` directory contains realistic Terraform configurations
2. **Plan Generation**: The Makefile automatically generates Terraform plans from the configurations
3. **JSON Conversion**: The plans are converted to JSON format for testing
4. **Test Execution**: Tests use the dynamically generated plan JSON files

This approach ensures:

- Tests always use up-to-date plan structures
- Realistic test scenarios with actual Terraform configurations
- Better maintainability as plan format evolves

### Test Workspaces

#### Web-App Workspace (`workspaces/web-app/`)

The `workspaces/web-app/` directory contains a realistic Terraform configuration that includes:

- **Networking**: VPC, public subnets, firewall rules
- **Compute**: GCP Compute Engine instance
- **Database**: Cloud SQL PostgreSQL instance and database
- **Security**: Database user with password
- **Random Resources**: Deployment ID, session tokens, secrets, and UUIDs

This provides a comprehensive test case with:

- Multiple resource types
- Sensitive values
- Complex relationships
- Real-world infrastructure patterns
- **Dynamic resources that change between plan runs**

#### Simple-Random Workspace (`workspaces/simple-random/`)

The `workspaces/simple-random/` directory contains a simple Terraform configuration that only uses the random provider:

- **Random Resources**: ID, string, password, UUID, integer, and pet name
- **Local Values**: Tags and prefixes derived from random resources
- **Outputs**: All random values and derived local values

This provides a focused test case with:

- Only random provider resources
- Simple dependency relationships
- Sensitive values (passwords)
- **Dynamic resources that change between plan runs**

## Running the Tests

### Prerequisites

1. Build the `terraform-ops` binary:

   ```bash
   make build
   ```

2. Install Terraform (for plan generation):

   ```bash
   # macOS
   brew install terraform

   # Ubuntu/Debian
   sudo apt-get install terraform

   # Or download from https://www.terraform.io/downloads.html
   ```

3. (Optional) Install Graphviz for DOT validation:

   ```bash
   # macOS
   brew install graphviz

   # Ubuntu/Debian
   sudo apt-get install graphviz

   # CentOS/RHEL
   sudo yum install graphviz
   ```

4. (Optional) Install jq for JSON validation:

   ```bash
   # macOS
   brew install jq

   # Ubuntu/Debian
   sudo apt-get install jq
   ```

### Running Tests

```bash
# Navigate to the plan-graph test directory
cd integration_tests/plan_graph

# Generate plans and run all tests
make all

# Or run specific targets:
make web-app-plan-json    # Generate web-app plan JSON only
make simple-random-plan-json    # Generate simple-random plan JSON only
make all-plans           # Generate both plan JSONs
make test                # Run tests with both generated plans
make web-app-test        # Run tests with web-app plan only
make simple-random-test  # Run tests with simple-random plan only
make graphs              # Generate all graph formats for both workspaces
make validate            # Validate both generated plan JSONs
make show-plan           # Show plan summary for both workspaces
make clean               # Clean up generated files
```

### From the Root Directory

```bash
# Run plan-graph integration tests
make test-integration-plan-graph
```

## Makefile Targets

### Core Targets

- `all` - Generate web-app plan and JSON (default target)
- `web-app-plan` - Generate Terraform plan file for web-app workspace
- `web-app-plan-json` - Convert web-app plan to JSON format
- `simple-random-plan` - Generate Terraform plan file for simple-random workspace
- `simple-random-plan-json` - Convert simple-random plan to JSON format
- `all-plans` - Generate both plan JSONs
- `clean` - Remove generated files and Terraform state

### Graph Generation

#### Web-App Workspace

- `web-app-graph-graphviz` - Generate Graphviz DOT format for web-app
- `web-app-graph-mermaid` - Generate Mermaid format for web-app
- `web-app-graph-plantuml` - Generate PlantUML format for web-app
- `web-app-graphs` - Generate all graph formats for web-app

#### Simple-Random Workspace

- `simple-random-graph-graphviz` - Generate Graphviz DOT format for simple-random
- `simple-random-graph-mermaid` - Generate Mermaid format for simple-random
- `simple-random-graph-plantuml` - Generate PlantUML format for simple-random
- `simple-random-graphs` - Generate all graph formats for simple-random

### Testing and Validation

- `test` - Test the plan-graph command with both workspaces
- `web-app-test` - Test the plan-graph command with web-app plan only
- `simple-random-test` - Test the plan-graph command with simple-random plan only
- `validate` - Validate both generated plan JSONs
- `web-app-validate` - Validate web-app plan JSON only
- `simple-random-validate` - Validate simple-random plan JSON only
- `show-plan` - Display plan summary for both workspaces
- `web-app-show-plan` - Display web-app plan summary only
- `simple-random-show-plan` - Display simple-random plan summary only

## Test Cases

### 1. Basic Functionality Tests

- `TestPlanGraphCommandWebApp` - Tests all supported output formats for web-app workspace
- `TestPlanGraphCommandSimpleRandom` - Tests all supported output formats for simple-random workspace
- `TestPlanGraphCommandWithOutputFileWebApp` - Tests file output functionality for web-app
- `TestPlanGraphCommandWithOutputFileSimpleRandom` - Tests file output functionality for simple-random
- `TestPlanGraphCommandHelp` - Tests help command output

### 2. Configuration Tests

- `TestPlanGraphCommandWithGroupingWebApp` - Tests different grouping strategies for web-app
- `TestPlanGraphCommandWithGroupingSimpleRandom` - Tests different grouping strategies for simple-random
- `TestPlanGraphCommandWithOptionsWebApp` - Tests various command line options for web-app
- `TestPlanGraphCommandWithOptionsSimpleRandom` - Tests various command line options for simple-random

### 3. Error Handling Tests

- `TestPlanGraphCommandErrorHandling` - Tests error cases and validation

### 4. Visualization Tool Tests

- `TestPlanGraphVisualizationToolsWebApp` - Tests that generated graphs can be processed by visualization tools for web-app
- `TestPlanGraphVisualizationToolsSimpleRandom` - Tests that generated graphs can be processed by visualization tools for simple-random
  - **Graphviz**: Validates DOT syntax with `dot` command
  - **Mermaid**: Validates Mermaid syntax structure
  - **PlantUML**: Validates PlantUML syntax structure

### 5. Complex Plan Tests

- `TestPlanGraphWithComplexPlan` - Tests with dynamically generated complex plans

## Expected Output Validation

### Web-App Workspace

#### Graphviz Output

- Contains `digraph terraform_plan`
- Includes `rankdir=TB` for top-to-bottom layout
- Contains `subgraph cluster_` for module grouping
- Includes all expected resources from the web-app plan
- Shows action types like `[CREATE]`
- Contains module groupings: `root`, `module.app`, `module.network`, `module.app.module.database`

#### Mermaid Output

- Starts with `graph TB`
- Contains `subgraph` and `end` statements
- Includes all expected resources
- Shows action types

#### PlantUML Output

- Starts with `@startuml` and ends with `@enduml`
- Contains `package` declarations
- Includes all expected resources
- Shows action types

### Simple-Random Workspace

#### Graphviz Output

- Contains `digraph terraform_plan`
- Includes `rankdir=TB` for top-to-bottom layout
- Contains `subgraph cluster_` for module grouping
- Includes all expected random resources
- Shows action types like `[CREATE]`
- Contains only `root` module grouping

#### Mermaid Output

- Starts with `graph TB`
- Contains `subgraph` and `end` statements
- Includes all expected random resources
- Shows action types

#### PlantUML Output

- Starts with `@startuml` and ends with `@enduml`
- Contains `package` declarations
- Includes all expected random resources
- Shows action types

## Visualization Tool Validation

### Graphviz

The test checks if the `dot` command can parse the generated DOT file:

```bash
dot -Tsvg -o /dev/null generated-graph.dot
```

### Mermaid

Validates basic Mermaid syntax:

- Starts with graph declaration
- Contains subgraph declarations
- Has proper end statements

### PlantUML

Validates basic PlantUML syntax:

- Starts with `@startuml`
- Ends with `@enduml`
- Contains package declarations

## Dynamic Plan Structure

### Web-App Plan

The dynamically generated web-app plan includes:

- **11+ resources** across multiple GCP services and the random provider
- **Multiple resource types**: VPC, subnets, firewall, compute instance, SQL database, random resources
- **Sensitive values**: Database password, random passwords, and tokens
- **Dependencies**: Between VPC, subnets, firewall, and compute resources, plus random resource dependencies
- **Changing resources**: Random resources that regenerate between plan runs due to timestamp keepers

### Simple-Random Plan

The dynamically generated simple-random plan includes:

- **6 random resources**: ID, string, password, UUID, integer, and pet name
- **Sensitive values**: Random password
- **Dependencies**: Between random resources using keepers
- **Changing resources**: All random resources regenerate between plan runs due to timestamp keepers

This provides realistic test scenarios that validate the graph generation capabilities with actual infrastructure and resources that change over time.

## Contributing

When adding new tests:

1. Follow the existing test structure and naming conventions
2. Use descriptive test names that explain the scenario
3. Include validation for both success and error cases
4. Add documentation for new test cases
5. Ensure tests are deterministic and don't depend on external state
6. Update the Terraform configuration if new resource types are needed
7. Consider adding separate test functions for each workspace when appropriate

## Troubleshooting

### Common Issues

1. **Binary not found**: Ensure `terraform-ops` is built and available in `../build/`
2. **Terraform not available**: Install Terraform for plan generation
3. **Graphviz not available**: Tests will skip Graphviz validation if `dot` command is not found
4. **Permission errors**: Ensure test directories are writable for temporary files
5. **Plan generation fails**: Check that the Terraform configuration is valid

### Debug Mode

Run tests with verbose output to see detailed execution:

```bash
go test -v -run TestPlanGraphCommandWebApp
go test -v -run TestPlanGraphCommandSimpleRandom
```

### Regenerate Plans

If the plan structure changes, regenerate them:

```bash
make clean
make all-plans
```

### Manual Plan Generation

For debugging, you can manually generate the plans:

```bash
# Web-app workspace
cd workspaces/web-app
terraform init
terraform plan --out=../web-app-plan.tfplan
terraform show -json ../web-app-plan.tfplan > ../web-app-plan.json

# Simple-random workspace
cd workspaces/simple-random
terraform init
terraform plan --out=../simple-random-plan.tfplan
terraform show -json ../simple-random-plan.tfplan > ../simple-random-plan.json
```
