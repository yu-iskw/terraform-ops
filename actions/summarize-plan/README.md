# Terraform Plan Summarizer Action

A GitHub Action that generates human-readable summaries of Terraform plan changes. This action uses the `terraform-ops summarize-plan` command to create comprehensive summaries showing resource changes, organized by action type (create, update, delete, replace), with statistics and breakdowns by provider, module, and resource type.

## Features

- **Multiple Output Formats**: Support for text, JSON, Markdown, table, and plan-like formats
- **Flexible Grouping**: Group resources by action, module, provider, or resource type
- **Comprehensive Statistics**: Detailed breakdowns and change counts
- **Sensitive Data Handling**: Option to hide sensitive value indicators
- **Detailed Information**: Show detailed change information when needed
- **Color Output**: Configurable color output for better readability
- **Easy Integration**: Works seamlessly with existing Terraform workflows

## Usage

### Basic Usage

```yaml
- name: Generate Terraform Plan Summary
  uses: ./actions/summarize-plan
  with:
    plan-file: "plan.json"
```

### Advanced Usage

```yaml
- name: Generate Terraform Plan Summary
  uses: ./actions/summarize-plan
  with:
    plan-file: "plan.json"
    format: "markdown"
    output-file: "terraform-summary.md"
    group-by: "provider"
    no-sensitive: "false"
    compact: "true"
    verbose: "true"
    show-details: "true"
    color: "always"
```

### Complete Workflow Example

```yaml
name: Terraform Plan with Summary

on:
  pull_request:
    paths:
      - "**.tf"
      - "**.tfvars"

jobs:
  plan-and-summary:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: "1.5.0"

      - name: Terraform Init
        run: terraform init

      - name: Terraform Plan
        run: |
          terraform plan -out=plan.tfplan
          terraform show -json plan.tfplan > plan.json

      - name: Generate Plan Summary
        uses: ./actions/summarize-plan
        with:
          plan-file: "plan.json"
          format: "markdown"
          output-file: "terraform-summary.md"
          group-by: "action"
          show-details: "true"

      - name: Upload Summary as Artifact
        uses: actions/upload-artifact@v4
        with:
          name: terraform-summary
          path: terraform-summary.md

      - name: Comment PR with Summary
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const summary = fs.readFileSync('terraform-summary.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## Terraform Plan Summary\n\n${summary}`
            });
```

## Inputs

| Input          | Description                                                 | Required | Default  |
| -------------- | ----------------------------------------------------------- | -------- | -------- |
| `plan-file`    | Path to the Terraform plan JSON file                        | Yes      | -        |
| `format`       | Output format (text, json, markdown, table, plan)           | No       | `text`   |
| `output-file`  | Output file path (default: stdout)                          | No       | -        |
| `group-by`     | Grouping strategy (action, module, provider, resource_type) | No       | `action` |
| `no-sensitive` | Hide sensitive value indicators                             | No       | `false`  |
| `compact`      | Compact output format                                       | No       | `false`  |
| `verbose`      | Enable verbose output for debugging                         | No       | `false`  |
| `show-details` | Show detailed change information                            | No       | `false`  |
| `color`        | Color output mode (auto, always, never)                     | No       | `auto`   |

## Outputs

| Output             | Description                                                       |
| ------------------ | ----------------------------------------------------------------- |
| `summary-content`  | The generated summary content (when output-file is not specified) |
| `output-file-path` | Path to the generated output file (when output-file is specified) |

## Supported Formats

### Text (Default)

Human-readable console output with color coding and clear formatting.

### JSON

Machine-readable structured data for programmatic processing.

### Markdown

GitHub-compatible markdown format for documentation and PR comments.

### Table

Tabular format for easy parsing and analysis.

### Plan

Terraform plan-like output format for familiarity.

## Grouping Strategies

### Action (Default)

Group resources by their lifecycle action:

- Create
- Update
- Delete
- Replace
- No-op

### Module

Group resources by their module address, showing changes within each module.

### Provider

Group resources by their provider, useful for understanding provider-specific changes.

### Resource Type

Group resources by their type (e.g., aws_instance, google_compute_instance).

## Statistics and Breakdowns

The summary includes comprehensive statistics:

- **Total Changes**: Overall count of resource changes
- **Action Breakdown**: Count of changes by action type
- **Provider Breakdown**: Count of changes by provider
- **Resource Breakdown**: Count of changes by resource type
- **Module Breakdown**: Count of changes by module

## Default Behavior

By default, the action:

- Uses text format for human-readable output
- Groups resources by action type
- Shows sensitive value indicators
- Uses auto color mode (color when terminal supports it)
- Provides compact output
- Includes basic change information

## Prerequisites

- The action requires a Terraform plan JSON file
- Generate the JSON file using: `terraform show -json <plan-file> > plan.json`
- The plan file should be generated from a valid Terraform plan

## Error Handling

The action will fail if:

- The plan file is not found
- The plan file is not valid JSON
- The specified format is not supported
- The specified grouping strategy is invalid

## License

This action is licensed under the Apache License 2.0.
