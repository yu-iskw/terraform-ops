# Terraform Plan Graph Action

A GitHub Action that generates visual graph representations of Terraform plan changes. This action uses the `terraform-ops plan-graph` command to create dependency graphs showing relationships between resources, grouped by modules, with clear indication of resource lifecycle changes (create, update, delete).

## Features

- **Multiple Output Formats**: Support for Graphviz DOT, Mermaid, and PlantUML formats
- **Flexible Grouping**: Group resources by module, action, or resource type
- **Comprehensive Visualization**: Include dependencies, outputs, variables, locals, and data sources by default
- **Customizable Layout**: Compact mode and exclusion options
- **Easy Integration**: Works seamlessly with existing Terraform workflows

## Usage

### Basic Usage

```yaml
- name: Generate Terraform Plan Graph
  uses: ./actions/plan-graph
  with:
    plan-file: "plan.json"
```

### Advanced Usage

```yaml
- name: Generate Terraform Plan Graph
  uses: ./actions/plan-graph
  with:
    plan-file: "plan.json"
    format: "mermaid"
    output-file: "terraform-graph.md"
    group-by: "module"
    no-outputs: "false"
    no-variables: "false"
    no-data-sources: "false"
    compact: "true"
    verbose: "true"
```

### Complete Workflow Example

```yaml
name: Terraform Plan with Graph

on:
  pull_request:
    paths:
      - "**.tf"
      - "**.tfvars"

jobs:
  plan-and-graph:
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

      - name: Generate Plan Graph
        uses: ./actions/plan-graph
        with:
          plan-file: "plan.json"
          format: "mermaid"
          output-file: "terraform-graph.md"

      - name: Upload Graph as Artifact
        uses: actions/upload-artifact@v4
        with:
          name: terraform-graph
          path: terraform-graph.md

      - name: Comment PR with Graph
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const graph = fs.readFileSync('terraform-graph.md', 'utf8');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## Terraform Plan Graph\n\n\`\`\`mermaid\n${graph}\n\`\`\``
            });
```

## Inputs

| Input             | Description                                       | Required | Default    |
| ----------------- | ------------------------------------------------- | -------- | ---------- |
| `plan-file`       | Path to the Terraform plan JSON file              | Yes      | -          |
| `format`          | Output format (graphviz, mermaid, plantuml)       | No       | `graphviz` |
| `output-file`     | Output file path (default: stdout)                | No       | -          |
| `group-by`        | Grouping strategy (module, action, resource_type) | No       | `module`   |
| `no-data-sources` | Exclude data source resources from the graph      | No       | `false`    |
| `no-outputs`      | Exclude output values from the graph              | No       | `false`    |
| `no-variables`    | Exclude variable values from the graph            | No       | `false`    |
| `no-locals`       | Exclude local values from the graph               | No       | `false`    |
| `no-modules`      | Exclude resources from modules from the graph     | No       | `false`    |
| `compact`         | Generate a more compact graph layout              | No       | `false`    |
| `verbose`         | Enable verbose output for debugging               | No       | `false`    |

## Outputs

| Output             | Description                                                       |
| ------------------ | ----------------------------------------------------------------- |
| `graph-content`    | The generated graph content (when output-file is not specified)   |
| `output-file-path` | Path to the generated output file (when output-file is specified) |

## Supported Formats

### Graphviz (DOT)

Default format that generates DOT files for use with Graphviz tools.

### Mermaid

Popular format for embedding diagrams in GitHub README files and documentation.

### PlantUML

Suitable for integration with documentation systems that support PlantUML.

## Node Types and Visual Representation

- **Resources**: House shapes with action-based colors
  - Create: Green
  - Update: Yellow
  - Delete: Red
  - Replace: Orange
  - No-op: Gray
- **Data Sources**: Cyan diamonds
- **Outputs**: Blue inverted houses (exports)
- **Variables**: Yellow cylinders (inputs)
- **Locals**: Pink octagons (computed values)

## Default Behavior

By default, the action includes all node types in the graph:

- Resources (including those in modules)
- Data sources
- Outputs
- Variables
- Local values
- Dependencies between resources

Use the `no-*` flags to exclude specific node types from the visualization.

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
