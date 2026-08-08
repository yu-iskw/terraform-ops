# `terraform-ops analyze`

`analyze` is the first v2 change-intelligence command. It consumes Terraform/OpenTofu plan JSON and produces deterministic, sanitized evidence for humans and CI.

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

## Exit behavior

`--fail-on` renders the report first and then returns a command error if the highest finding meets or exceeds the configured threshold. This makes the same invocation useful for both PR summaries and CI gates.

## Machine contract

The JSON report is versioned independently of the Terraform/OpenTofu plan format. Its schema lives at `schemas/analysis-report-v1.schema.json`.
