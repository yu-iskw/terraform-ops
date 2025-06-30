# Project Directory Structure

This project follows a standard Go project layout with clear separation of concerns and modular organization.

## Core Application Directories

### `cmd/`

Contains the main application entry points and executable binaries. This is where the CLI application's main function and command-line interface are defined.

### `internal/`

Houses private application and library code that is not intended to be imported by external applications. This includes:

- **`internal/app/`**: Core application logic and CLI command definitions using the Cobra framework
- **`internal/pkg/`**: Internal packages for application configuration and shared utilities
- **`internal/list_providers/`**: Functionality for listing and analyzing Terraform providers
- **`internal/plan_graph/`**: Implementation for generating and visualizing Terraform plan graphs
- **`internal/show_terraform/`**: Features for displaying Terraform configuration and state information

### `pkg/`

Contains public library code that can be imported and used by other applications. This follows Go's convention for reusable packages.

- **`pkg/terraform/`**: Public Terraform client and utilities for interacting with Terraform operations

## Configuration and Data Directories

### `configs/`

Stores application configuration files, templates, and settings that control the behavior of the application.

### `deployments/`

Contains deployment-related files, infrastructure configurations, and deployment scripts for different environments.

## Documentation and Examples

### `docs/`

Documentation files including user guides, API documentation, and feature explanations.

### `examples/`

Example code and usage demonstrations showing how to use the application's features and APIs.

## Testing and Quality Assurance

### `integration_tests/`

Comprehensive integration tests that verify the complete functionality of the application, including:

- **`integration_tests/plan_graph/`**: Tests for plan graph generation functionality
- **`integration_tests/show_terraform/`**: Tests for Terraform display features
- **`integration_tests/build/`**: Build and deployment testing

## Development and Build Tools

### `scripts/`

Utility scripts for building, installing, testing, and other development operations.

### `tools/`

Development tools, utilities, and helper scripts for maintaining code quality and development workflow.

### `assets/`

Static assets, resources, and files used by the application (templates, images, etc.).

## Build and Output Directories

### `build/`

Generated build artifacts, compiled binaries, and temporary build files (typically gitignored).

## Development Environment

### `.github/`

GitHub-specific configurations including workflows, issue templates, and repository settings.

### `.vscode/`

VS Code editor configuration and workspace settings.

### `.cursor/`

Cursor IDE configuration and settings.

### `.trunk/`

Trunk-based development configuration and tooling setup.

## Key Files

- **`go.mod`**: Go module definition and dependency management
- **`go.sum`**: Dependency checksums for reproducible builds
- **`Makefile`**: Build automation and development tasks
- **`README.md`**: Project overview and getting started guide
- **`LICENSE`**: Project license information
- **`.gitignore`**: Git ignore patterns for build artifacts and temporary files
- **`.go-version`**: Go version specification for the project
