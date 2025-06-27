# terraform-ops

A command-line interface tool for managing Terraform operations and workflows. This tool provides utilities to inspect and analyze Terraform configurations across multiple workspaces.

## Features

- **Terraform Block Analysis**: Extract and display information from Terraform configuration blocks
- **Multi-workspace Support**: Process multiple Terraform workspaces in a single command
- **Machine-readable Output**: JSON format for easy integration with scripts and tools
- **Non-recursive Scanning**: Focus on workspace root directories for efficient processing
- **Error Resilience**: Continue processing remaining workspaces even if individual ones fail

## Installation

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
        "type": "s3",
        "config": {
          "bucket": "my-bucket",
          "key": "path/to/state.tfstate",
          "region": "us-east-1",
          "encrypt": "true"
        }
      },
      "required_providers": {
        "aws": "~> 4.0",
        "random": ""
      }
    }
  },
  {
    "path": "/absolute/path/to/workspace2",
    "terraform": {
      "required_version": ">= 1.0.0",
      "required_providers": {
        "google": ">=4.83.0,<5.0.0",
        "aws": "3.0.0"
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
  - **`config`** – Key-value configuration settings (only primitive values: string, number, bool)
- **`terraform.required_providers`** – The set of required providers and their declared version constraints (empty object when no providers are declared)

#### Behavior

- **Non-recursive**: Only scans `.tf` files directly in the specified directories (no subdirectory traversal)
- **Error handling**: Individual workspace errors are printed to stderr but don't stop processing of remaining paths
- **JSON output**: All information is returned in a machine-readable JSON array – one element per inspected workspace
- **Order preservation**: Results are returned in the same order as the input paths

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

### Project Structure

```
terraform-ops/
├── cmd/terraform-ops/     # Main application entry point
├── internal/              # Private application code
│   ├── app/              # CLI command definitions
│   ├── pkg/              # Internal packages
│   └── show_terraform/   # Terraform block parsing logic
├── pkg/                  # Public library code
│   └── terraform/        # Terraform client utilities
├── test/                 # Integration tests
│   └── workspaces/       # Test Terraform configurations
├── docs/                 # Documentation
├── scripts/              # Build and utility scripts
└── examples/             # Usage examples
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

Test workspaces are located in `test/workspaces/` and cover various scenarios:

- Simple provider configurations
- Backend configurations (S3, GCS)
- Invalid Terraform files
- Empty workspaces
- Missing provider declarations

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
