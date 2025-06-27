# `show-terraform` subcommand Design Document

## 1. Overview

`show-terraform` is a new sub-command of the `terraform-ops` CLI. It inspects the `terraform` block that may appear in one or more `.tf` files located **directly** in the supplied workspace directories (no recursive traversal). It reports:

- `required_version` – the Terraform CLI version constraint string.
- `backend` – backend type and key-value settings (primitive values only).
- `required_providers` – the set of required providers and their declared version constraints.

All information is returned in a machine-readable JSON array – one element per inspected workspace.

## 2. Usage

```shell
terraform-ops show-terraform <path1> <path2> ...
```

### Arguments

- `<path...>`: One or more workspace directories to inspect. Only `.tf` files located **directly** in each directory are parsed.

### Output Format

```json
[
  {
    "path": "/absolute/path/to/workspace",
    "terraform": {
      "required_version": ">= 1.4.0, < 2.0.0",
      "backend": {
        "type": "s3",
        "config": {
          "bucket": "my-bucket",
          "key": "path/to/state.tfstate",
          "region": "us-east-1"
        }
      },
      "required_providers": {
        "aws": "~> 4.0",
        "random": ""
      }
    }
  }
]
```

- `path`: Absolute path that was scanned.
- `terraform`: Object containing all Terraform configuration details.
  - `required_version`: Empty when not declared.
  - `backend`: Omitted when no backend block is present.
  - `required_providers`: Empty object when no providers are declared.

## 3. Implementation Highlights

- Parsing is implemented in `internal/show_terraform/show_terraform.go` using `github.com/hashicorp/hcl/v2/hclparse`.
- Only primitive attribute values inside the backend block are collected.
- Errors in individual files are printed to `stderr`; the command still attempts to process remaining inputs.
- The command is registered in `internal/app/app.go`.

For details on the Terraform `terraform` block refer to HashiCorp documentation: <https://developer.hashicorp.com/terraform/language/terraform>.
