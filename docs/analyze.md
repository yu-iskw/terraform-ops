# `terraform-ops analyze`

`analyze` is the v2 change-intelligence command. It consumes Terraform/OpenTofu plan JSON and produces deterministic, sanitized evidence for humans and CI.

```bash
terraform show -json tfplan > plan.json
terraform-ops analyze plan.json
terraform-ops analyze plan.json --format json
terraform-ops analyze plan.json --format markdown
terraform-ops analyze plan.json --fail-on high
cat plan.json | terraform-ops analyze - --engine terraform
```

## Security model

Raw plan values are source data, not the domain model. The command:

1. enforces a bounded input size (64 MiB by default);
2. parses the Terraform/OpenTofu 1.x JSON plan contract;
3. applies `before_sensitive` / `after_sensitive` masks before normalization;
4. removes variable values from the normalized model;
5. exposes only sanitized evidence to analyzers and report formatters.

`--redaction strict` removes all before/after resource values, not only values marked sensitive by the source plan.

The stable report intentionally does not publish raw before/after values. It publishes change semantics such as action order, replacement paths, action reasons, unknown/sensitive paths, checks, drift, and dependency/blast-radius counts.

## Source-aware SARIF

SARIF is intentionally separate from the stable analysis-report schema because source locations depend on the repository checkout rather than on the plan itself.

For a Terraform workspace at the repository root:

```bash
terraform-ops analyze plan.json \
  --engine terraform \
  --redaction strict \
  --format sarif \
  --workspace-root . \
  --output analysis.sarif
```

For a Terraform workspace below the repository root, use `--source-root` so SARIF artifact URIs remain repository-relative:

```bash
terraform-ops analyze infra/prod/plan.json \
  --engine terraform \
  --redaction strict \
  --format sarif \
  --workspace-root infra/prod \
  --source-root . \
  --output analysis.sarif
```

In that example a resource declared in `infra/prod/main.tf` is emitted with the SARIF URI `infra/prod/main.tf`, not merely `main.tf`.

`--format sarif` requires `--workspace-root`. terraform-ops parses local `.tf` files under that root and maps resource/data-source addresses to exact HCL declaration ranges. Local module sources (`./...` and `../...`) are traversed only when their canonical path stays inside the workspace root.

`--source-root` controls only the artifact-URI base. It defaults to `--workspace-root` and must canonically contain the workspace root. This separation means a caller can constrain source traversal to one Terraform workspace while still producing repository-relative locations for GitHub code scanning. `actions/analyze-plan` does this automatically by using the caller repository as the source root.

The SARIF contract is deliberately conservative:

- only findings with an exact local resource/data-source source range become SARIF results;
- plan-level findings without a physical source location are omitted;
- unresolved remote-module resources are omitted rather than guessed;
- instance keys from `count`, `for_each`, and module instances are normalized back to their static declaration address;
- SARIF uses version 2.1.0;
- critical/high findings map to `error`, medium to `warning`, and low/info to `note`.

The ordinary text, JSON, and Markdown analysis outputs are unchanged and remain independent of checkout paths.

## Initial rules

| Rule                       | Default severity | Meaning                                         |
| -------------------------- | ---------------- | ----------------------------------------------- |
| `TFOPS-PLAN-ERRORED`       | high             | Planning reported an error.                     |
| `TFOPS-PLAN-INCOMPLETE`    | medium           | The plan may require another round to converge. |
| `TFOPS-CHECK-FAILED`       | high             | A Terraform/OpenTofu check failed or errored.   |
| `TFOPS-LIFECYCLE-DELETE`   | medium           | A managed resource is deleted.                  |
| `TFOPS-LIFECYCLE-REPLACE`  | medium           | A managed resource is replaced.                 |
| `TFOPS-DRIFT-DETECTED`     | medium           | External drift is present in the plan.          |
| `TFOPS-SENSITIVE-MUTATION` | info             | Sensitive paths participate in a change.        |
| `TFOPS-UNKNOWN-AFTER`      | info             | Values remain unknown until apply.              |

Every rule is deterministic and evidence-backed. Provider-specific risk classification is intentionally outside the initial core.

## Engine selection

`--engine auto` records the source as `unknown-compatible` because the compatible JSON plan format does not reliably identify Terraform vs OpenTofu. Use `--engine terraform` or `--engine opentofu` when the caller knows which engine generated the plan.

Terraform emits top-level `applyable` and, on current versions, `complete` metadata. Current OpenTofu JSON plans do not expose those Terraform-specific flags. The source adapter preserves explicit values when present; when `applyable` is absent it derives whether the plan contains actionable resource/output changes, and when both producer flags are absent a non-errored plan is treated as complete. The cross-engine equivalence harness excludes those producer-specific fields from equality checks.

## Exit behavior

`--fail-on` renders the report first and then returns a command error if the highest finding meets or exceeds the configured threshold. This makes the same invocation useful for both PR summaries and CI gates.

## Machine contract

The JSON report is versioned independently of the Terraform/OpenTofu plan format. Its schema lives at `schemas/analysis-report-v1.schema.json`.
