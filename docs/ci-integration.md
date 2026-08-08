# CI integration

`actions/analyze-plan` packages the deterministic `terraform-ops analyze` pipeline for GitHub Actions. The action does not run `terraform plan`, does not mutate infrastructure, and does not upload data to GitHub code scanning by itself.

The caller owns plan generation, permissions, artifact retention, and optional SARIF upload.

## Minimal workflow

```yaml
name: Terraform change analysis

on:
  pull_request:

permissions:
  contents: read

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: 1.15.8
          terraform_wrapper: false

      - name: Generate plan JSON
        run: |
          terraform init -input=false
          terraform plan -input=false -out=plan.tfplan
          terraform show -json plan.tfplan > plan.json

      - name: Analyze plan
        id: analyze
        uses: ./actions/analyze-plan
        with:
          plan-file: plan.json
          workspace-root: .
          engine: terraform
          redaction: strict
          fail-on: high

      - name: Inspect generated paths
        run: |
          echo "SARIF: ${{ steps.analyze.outputs.sarif-file }}"
          echo "JSON: ${{ steps.analyze.outputs.json-file }}"
          echo "Markdown: ${{ steps.analyze.outputs.markdown-file }}"
```

For consumers outside this repository, replace the local action reference with a released/tagged terraform-ops action reference after the feature is released.

## Upload SARIF to GitHub code scanning

The action intentionally does not request `security-events: write`. A caller that wants code-scanning results grants that permission to the upload step/job and uploads the generated file explicitly.

```yaml
permissions:
  contents: read
  security-events: write

steps:
  - uses: actions/checkout@v4

  # Generate plan.json here.

  - name: Analyze plan
    id: analyze
    uses: ./actions/analyze-plan
    with:
      plan-file: plan.json
      workspace-root: .
      engine: terraform
      redaction: strict
      fail-on: none

  - name: Upload terraform-ops SARIF
    uses: github/codeql-action/upload-sarif@v3
    with:
      sarif_file: ${{ steps.analyze.outputs.sarif-file }}
```

Pin third-party actions to reviewed commit SHAs in production workflows according to your organization policy.

## Inputs

| Input               | Default          | Purpose                                                                    |
| ------------------- | ---------------- | -------------------------------------------------------------------------- |
| `plan-file`         | required         | Terraform/OpenTofu JSON plan produced by `show -json`                      |
| `workspace-root`    | `.`              | Root for exact local HCL source mapping                                    |
| `engine`            | `auto`           | `auto`, `terraform`, or `opentofu`                                         |
| `redaction`         | `strict`         | `standard` or `strict`; CI defaults to strict                              |
| `fail-on`           | `none`           | Severity threshold evaluated after artifacts are written                   |
| `output-directory`  | `.terraform-ops` | Repository-relative output directory                                       |
| `write-job-summary` | `true`           | Append the Markdown analysis to `GITHUB_STEP_SUMMARY`                      |

## Outputs

The action writes three sanitized artifacts and exposes repository-relative paths through GitHub outputs:

- `sarif-file` — SARIF 2.1.0, containing only exactly source-located findings;
- `json-file` — stable terraform-ops analysis-report JSON;
- `markdown-file` — human-readable PR/job summary.

## Security boundaries

The action defaults to strict redaction. It also enforces path containment before analysis:

- `plan-file` must resolve inside `GITHUB_WORKSPACE`;
- `workspace-root` must resolve inside `GITHUB_WORKSPACE`;
- symlink resolution is performed before containment checks;
- `output-directory` must be repository-relative and cannot contain `..` traversal;
- the action never runs Terraform/OpenTofu or shell commands derived from plan content;
- source indexing follows local modules only when their canonical path remains inside `workspace-root`.

Raw plan files can contain sensitive values. The caller should avoid uploading raw plan JSON unless that is explicitly intended and governed. terraform-ops outputs remain subject to the redaction guarantees of the normalized `ChangeSet` boundary.

## Source-location limitations

Source-aware SARIF is intentionally precision-first. terraform-ops does not invent a location when it cannot prove one.

Currently indexed source declarations are standard `.tf` HCL `resource` and `data` blocks in the workspace and local modules. Remote registry/Git/HTTP modules are not traversed. Findings without an exact local range remain available in normal analysis output but are omitted from SARIF.

## Failure behavior

`fail-on` is evaluated only after Markdown, JSON, and SARIF files are generated. This allows a failing quality gate to preserve evidence for the job summary and optional upload.

Exit behavior therefore separates two concerns:

1. produce deterministic evidence;
2. enforce the configured severity policy.
