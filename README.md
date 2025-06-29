# terraform-ops

A command-line interface tool for managing Terraform operations and workflows. This tool provides utilities to inspect and analyze Terraform configurations across multiple workspaces.

## Features

- **Terraform Block Analysis**: Extract and display information from Terraform configuration blocks
- **Plan Graph Generation**: Generate visual dependency graphs from Terraform plan files
- **Multi-workspace Support**: Process multiple Terraform workspaces in a single command
- **Machine-readable Output**: JSON format for easy integration with scripts and tools
- **Non-recursive Scanning**: Focus on workspace root directories for efficient processing
- **Error Resilience**: Continue processing remaining workspaces even if individual ones fail
- **Multiple Graph Formats**: Support for Graphviz, Mermaid, and PlantUML formats
- **GitHub Action Integration**: Ready-to-use GitHub Action for CI/CD workflows

## Installation

### Using Homebrew (Recommended)

```shell
# Add the custom tap
brew tap yu/terraform-ops

# Install terraform-ops
brew install terraform-ops
```

**Note**: You'll need to create a `homebrew-terraform-ops` repository first. See [Homebrew Installation Guide](docs/homebrew.md) for detailed setup instructions.

### From Source

```shell
git clone https://github.com/yu/terraform-ops.git
cd terraform-ops
make install
```

### Using Go Install

```shell
go install github.com/yu/terraform-ops@latest
```

### Building from Source

```shell
# Build for current platform
make build

# Build for multiple platforms
make build-all

# Development with live reload (requires air)
make dev
```

## Usage

### `show-terraform`

Display information from the terraform block in workspaces.

```shell
terraform-ops show-terraform <path1> <path2>
```

#### Examples

**Basic usage with multiple workspaces:**

```shell
terraform-ops show-terraform ./workspace1 ./workspace2 ./workspace3
```

**Example output:**

```json
[
  {
    "path": "/absolute/path/to/workspace1",
    "terraform": {
      "required_version": ">= 1.4.0, < 2.0.0",
      "backend": {
        "type": "gcs",
        "config": {
          "bucket": "terraform-state-prod",
          "prefix": "terraform/state",
          "impersonate_service_account": "test-service-account@terraform-ops-test.iam.gserviceaccount.com"
        }
      },
      "required_providers": {
        "google": "~> 4.0",
        "random": ""
      }
    }
  },
  {
    "path": "/absolute/path/to/workspace2",
    "terraform": {
      "required_version": ">= 1.0.0",
      "backend": {
        "type": "s3",
        "config": {
          "bucket": "my-bucket",
          "key": "path/to/state.tfstate",
          "region": "us-east-1",
          "encrypt": "true"
        }
      },
      "required_providers": {
        "aws": ">=4.83.0,<5.0.0"
      }
    }
  }
]
```

#### Output Fields

The `show-terraform` command inspects Terraform configuration files (\*.tf) in the specified paths and outputs information contained in the terraform block:

- **`path`** – Absolute path of the scanned workspace directory
- **`terraform.required_version`** – The Terraform CLI version constraint string (empty when not declared)
- **`terraform.backend`** – Backend type and key-value settings (omitted when no backend block is present)
  - **`type`** – Backend type (e.g., "s3", "gcs", "azurerm")
  - **`config`** – Key-value configuration settings (only primitive values: string, number, bool). Includes all optional fields like `impersonate_service_account` for GCS backends.
- **`terraform.required_providers`** – The set of required providers and their declared version constraints (empty object when no providers are declared)

#### Behavior

- **Non-recursive**: Only scans `.tf` files directly in the specified directories (no subdirectory traversal)
- **Error handling**: Individual workspace errors are printed to stderr but don't stop processing of remaining paths
- **JSON output**: All information is returned in a machine-readable JSON array – one element per inspected workspace
- **Order preservation**: Results are returned in the same order as the input paths

### `plan-graph`

Generate a visual graph representation of Terraform plan changes for the given workspace. The generated graph shows relationships between resources, grouped by modules, with clear indication of resource lifecycle changes (create, update, delete).

```shell
terraform-ops plan-graph <PLAN_FILE> [OPTIONS]
```

#### Basic Examples

**Basic usage:**

```shell
# Generate Graphviz graph from plan file
terraform-ops plan-graph plan.json

# Generate Mermaid graph
terraform-ops plan-graph --format mermaid plan.json

# Save to file
terraform-ops plan-graph --output graph.dot plan.json
```

**Advanced Examples:**

```shell
# Generate compact graph with specific grouping
terraform-ops plan-graph --compact --group-by action plan.json

# Exclude specific elements
terraform-ops plan-graph --no-data-sources --no-outputs plan.json

# Verbose output for debugging
terraform-ops plan-graph --verbose plan.json
```

**Workflow Integration Examples:**

```shell
# Generate plan and create graph in one workflow
terraform plan -out=plan.tfplan
terraform show -json plan.tfplan > plan.json
terraform-ops plan-graph plan.json > infrastructure-graph.dot
dot -Tpng infrastructure-graph.dot -o infrastructure-graph.png
```

#### Options

- `--format <FORMAT>`: Output format (default: "graphviz")
  - Supported formats: `graphviz`, `mermaid`, `plantuml`
- `--output <FILE>`: Output file path (default: stdout)
- `--group-by <GROUPING>`: Grouping strategy (default: "module")
  - Supported groupings: `module`, `action`, `resource_type`
- `--no-data-sources`: Exclude data source resources from the graph
- `--no-outputs`: Exclude output values from the graph
- `--no-variables`: Exclude variable values from the graph
- `--no-locals`: Exclude local values from the graph
- `--compact`: Generate a more compact graph layout
- `--verbose`: Enable verbose output for debugging

#### Supported Graph Visualization Tools

- **[Graphviz](https://graphviz.org/)**: Industry-standard graph visualization tool
  - Supports complex layouts and styling
  - Excellent for large infrastructure graphs
  - Can generate PNG, SVG, PDF outputs
- **[Mermaid](https://mermaid.js.org/)**: Modern diagramming tool
  - Web-based rendering
  - Good for documentation and web interfaces
  - Supports interactive features
- **[PlantUML](https://plantuml.com/)**: UML-focused diagramming
  - Clean, professional appearance
  - Good for documentation
  - Supports various output formats

#### Node Types and Visual Representation

- **Resources**: House shapes (with action-based colors)
  - **CREATE** (`actions: ["create"]`): Resources to be created (Green)
  - **UPDATE** (`actions: ["update"]`): Resources to be modified (Blue)
  - **DELETE** (`actions: ["delete"]`): Resources to be destroyed (Red)
  - **REPLACE** (`actions: ["delete", "create"]`): Resources to be recreated (Orange)
  - **NO-OP** (`actions: ["no-op"]`): No changes planned (Grey)
- **Data Sources**: Cyan diamonds
- **Outputs**: Blue inverted houses (exports)
- **Variables**: Yellow cylinders (inputs)
- **Locals**: Pink octagons (computed values)

### GitHub Action Integration

The project includes a GitHub Action for generating plan graphs in CI/CD workflows:

```yaml
- name: Generate Terraform Plan Graph
  uses: yu/terraform-ops/actions/plan-graph@v1
  with:
    plan-file: "plan.json"
    format: "mermaid"
    group-by: "module"
    show-outputs: "true"
    compact: "false"
```

See [actions/plan-graph/README.md](actions/plan-graph/README.md) for detailed usage instructions.

### CI/CD Workflows

The project includes comprehensive GitHub Actions workflows for automated testing, building, and releasing:

#### Available Workflows

- **Unit Tests** (`.github/workflows/unit_tests.yml`): Runs unit tests on pull requests and pushes
- **Integration Tests** (`.github/workflows/integration_tests.yml`): Tests against multiple Terraform versions (1.7.0-1.12.0)
- **Code Quality** (`.github/workflows/trunk_check.yml`): Runs linting, formatting, and security checks
- **Release** (`.github/workflows/release.yml`): Automatically builds and releases binaries when tags are pushed
- **Dependency Updates** (`.github/dependabot.yml`): Monthly updates for GitHub Actions and Go modules

#### Automated Testing Matrix

The integration tests run against multiple Terraform versions to ensure compatibility:

- Terraform 1.7.0 through 1.12.0
- Tests both `show-terraform` and `plan-graph` commands
- Validates against various workspace configurations

#### Release Process

1. **Tag a Release**: `git tag v1.0.0 && git push origin v1.0.0`
2. **Automated Build**: GitHub Actions builds for all platforms:
   - Linux AMD64
   - macOS AMD64/ARM64
   - Windows AMD64
3. **Release Creation**: Automatically creates GitHub release with binaries
4. **Homebrew Update**: Use the provided script to update the Homebrew formula

#### Code Quality Tools

The project uses [Trunk](https://trunk.io/) for comprehensive code quality checks:

Run locally with:

```shell
make format    # Format code
make lint      # Lint and check code quality
make vet       # Run Go vet
```

## Documentation

The project includes comprehensive documentation to help users and contributors:

### Command Documentation

- **[Plan Graph Command](docs/plan_graph.md)**: Complete specification and usage guide for the `plan-graph` command
- **[Show Terraform Command](docs/show_terraform.md)**: Detailed documentation for the `show-terraform` command
- **[Project Structure](docs/project_structure.md)**: Overview of the codebase organization and architecture

### Installation Guides

- **[Homebrew Installation](docs/homebrew.md)**: User guide for installing via Homebrew
- **[Homebrew Setup Guide](docs/homebrew-setup-guide.md)**: Complete developer guide for setting up Homebrew distribution
- **[Homebrew Summary](docs/homebrew-summary.md)**: Summary of Homebrew support and maintenance

### GitHub Actions

- **[Plan Graph Action](actions/plan-graph/README.md)**: Detailed usage guide for the GitHub Action

### Development Resources

- **[Internal Documentation](internal/README.md)**: Architecture overview and design patterns
- **[Integration Tests](integration_tests/)**: Test workspaces and examples for both commands
- **[Examples](examples/)**: Basic usage examples and code samples

## Development

### Prerequisites

- Go 1.24.4 or later
- Make (for build automation)

### Setup

```shell
# Clone the repository
git clone https://github.com/yu/terraform-ops.git
cd terraform-ops

# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Run integration tests
make test-integration
```

### Available Make Targets

- `build` - Build the binary for current platform
- `build-all` - Build for multiple platforms (Linux, macOS, Windows)
- `test` - Run unit tests
- `test-integration` - Run integration tests
- `coverage` - Run tests with coverage report
- `format` - Format code using trunk
- `lint` - Lint code using trunk
- `install` - Install binary to $GOPATH/bin
- `clean` - Clean build artifacts
- `dev` - Development with live reload (requires air)
- `homebrew-test` - Test Homebrew formula installation
- `homebrew-install` - Install via Homebrew formula
- `homebrew-uninstall` - Uninstall via Homebrew

### Project Structure

```shell
terraform-ops/
├── cmd/terraform-ops/     # Main application entry point
├── internal/              # Private application code
│   ├── app/              # CLI command definitions
│   ├── commands/         # CLI command implementations
│   │   ├── show_terraform.go
│   │   └── plan_graph.go
│   ├── config/           # Application configuration
│   ├── core/             # Core types and error handling
│   └── terraform/        # Terraform-specific functionality
│       ├── config/       # HCL configuration parsing
│       ├── plan/         # Terraform plan parsing
│       └── graph/        # Graph generation and dependency analysis
│           └── generators/ # Graph format generators
│               ├── factory.go
│               ├── graphviz.go
│               ├── mermaid.go
│               └── plantuml.go
├── pkg/                  # Public library code
│   └── terraform/        # Terraform client utilities
├── integration_tests/    # Integration tests
│   ├── show_terraform/   # Show terraform command tests
│   └── plan_graph/       # Plan graph command tests
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
├── Formula/              # Homebrew formula
├── examples/             # Usage examples
├── actions/              # GitHub Actions
│   └── plan-graph/       # Plan graph GitHub Action
└── assets/               # Static assets
```

## Testing

The project includes comprehensive test coverage:

```shell
# Run all tests
make test

# Run integration tests
make test-integration

# Generate coverage report
make coverage
```

Test workspaces are located in `integration_tests/` and cover various scenarios:

- Simple provider configurations
- Backend configurations (S3, GCS)
- Invalid Terraform files
- Empty workspaces
- Missing provider declarations
- Plan graph generation scenarios
  - Web application with modules and dependencies
  - Simple random resource configurations
  - Multiple graph formats (Graphviz, Mermaid, PlantUML)
  - Different grouping strategies (module, action, resource_type)
  - Various command line options and filters

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go conventions and best practices
- Use `make format` to format code
- Use `make lint` to check code quality
- Write tests for new functionality

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## References

- [Terraform Documentation](https://developer.hashicorp.com/terraform/docs)
- [HCL Language Documentation](https://developer.hashicorp.com/terraform/language)
- [Cobra CLI Framework](https://github.com/spf13/cobra)
