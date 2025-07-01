# Summarize Plan Integration Tests

This directory contains integration tests for the `terraform-ops summarize-plan` command.

## Overview

These tests verify that the `summarize-plan` command works correctly with real Terraform plan files and produces the expected output in different formats.

## Test Coverage

The integration tests cover:

- **All Output Formats**: text, json, markdown, table, and plan formats
- **Plan Format Details**: Testing the new Terraform plan-like output format
- **Error Handling**: Invalid plan files, invalid formats
- **JSON Structure**: Validating JSON output structure and content
- **File Output**: Writing output to files
- **Help Documentation**: Verifying help text includes all formats

## Test Data

The tests use the sample plan file from `../show_terraform/workspaces/sample-plan.json` which contains:

- 3 resources to be created (2 root, 1 in module)
- AWS resources (instances and security groups)
- Sensitive data handling
- Module resources

## Running the Tests

```bash
# Build and run all tests
make all

# Build only
make build

# Run tests only (assumes binary is built)
make test
```

## Test Structure

- `TestSummarizePlanCommand`: Tests all output formats with expected content
- `TestSummarizePlanCommandWithInvalidPlan`: Tests error handling for invalid files
- `TestSummarizePlanCommandWithInvalidFormat`: Tests error handling for invalid formats
- `TestSummarizePlanCommandJSONOutput`: Validates JSON output structure
- `TestSummarizePlanCommandWithOutputFile`: Tests file output functionality
- `TestSummarizePlanCommandHelp`: Verifies help documentation

## Expected Output Examples

### Text Format

```text
Terraform Plan Summary
======================

📊 Statistics
-------------
Total Changes: 3

🔄 Resource Changes
-------------------

➕ Create (3)
--------------
  aws_instance.web
  aws_security_group.web
  module.database.aws_instance.db
```

### Plan Format

```text
Terraform will perform the following actions:

  + aws_instance.web
  + aws_security_group.web
  + module.database.aws_instance.db

Plan: 3 to add.
```

### Plan Format with Details

```text
Terraform will perform the following actions:

  + aws_instance.web
      {
        + ami = "ami-12345678"
        + instance_type = "t2.micro"
      }
  + aws_security_group.web
      {
        + description = "Security group for web server"
        + name = "web-sg"
      }
  + module.database.aws_instance.db
      {
        + ami = "ami-87654321"
        + instance_type = "t2.small"
      }

Plan: 3 to add.
```
