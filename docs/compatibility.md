# Terraform and OpenTofu compatibility harness

The compatibility harness generates real binary plans with Terraform and OpenTofu, converts them with each engine's own `show -json`, normalizes them through terraform-ops, and compares the semantics that both JSON contracts expose.

It is intentionally stronger than fixture-only parsing tests: every tested artifact must be produced by the real engine version under test.

## Supported test matrix

Terraform:

- 1.7.0
- 1.8.0
- 1.9.0
- 1.10.0
- 1.11.0
- 1.12.0
- 1.15.8

OpenTofu:

- 1.10.10
- 1.11.6
- 1.12.5

The exact matrix is encoded in `compat/run.sh`. Updating support policy therefore requires an explicit code review rather than silently following an unpinned latest tag.

## Run locally

Docker with Compose v2 and the repository's Go toolchain are required.

```bash
make test-compatibility
```

You can override either version list for focused testing:

```bash
TERRAFORM_VERSIONS="1.15.8" \
OPENTOFU_VERSIONS="1.12.5" \
make test-compatibility
```

Generated plan JSON is written under `compat/artifacts/` and ignored by Git because raw plan JSON can contain sensitive values.

## Architecture

```text
compat/workspaces/equivalence
          │
          ├───────────────┐
          ▼               ▼
 Terraform container   OpenTofu container
 plan + show -json     plan + show -json
          │               │
          └───────┬───────┘
                  ▼
          raw plan JSON artifacts
                  │
                  ▼
        terraform-ops source adapter
                  │
          strict normalization
                  │
                  ▼
       overlapping semantic projection
                  │
                  ▼
          deterministic equality
```

The fixture uses the built-in `terraform_data` resource and a local child module, so plan generation does not depend on an external provider registry. It includes a sensitive canary, module inputs/outputs, multiple resource changes, and an output dependency.

## What equivalence means

The test does **not** require byte-for-byte equality between Terraform and OpenTofu plan JSON. It compares normalized change semantics where the contracts overlap:

- resource addresses, modes, types, names, instance identity, and action order;
- action reasons, replacement paths, unknown paths, and import metadata when present;
- output names, actions, and unknown paths;
- checks when present;
- drift and relevant attributes when present;
- normalized dependency graph and evidence;
- plan error state;
- plan-format **major** version.

The comparison canonicalizes collection ordering before equality because JSON producer ordering is not semantic. Resource values and output values are excluded because strict mode deliberately removes them.

Producer identity/version and plan-format minor versions are not equality dimensions. Terraform-specific `applyable` / `complete` metadata is also excluded from cross-engine equality because current OpenTofu JSON plans do not expose those fields.

Sensitivity-mask paths and redaction counters are **not** equality dimensions. Real engine testing showed that Terraform and OpenTofu can propagate sensitivity masks differently through equivalent derived values while still producing the same resource actions and dependency semantics. Instead of treating one producer's mask propagation as canonical, the harness tests the stronger security invariant independently for every artifact: raw input contains the canary, strict normalization removes all values, and the serialized `ChangeSet` cannot contain the canary.

The source adapter still provides compatible `ChangeSet.Plan` booleans for OpenTofu: when `applyable` is absent it derives whether actionable resource/output changes exist, and when both Terraform-specific metadata fields are absent a non-errored plan is treated as complete.

## Secret-canary invariant

Before normalization, every generated plan must contain the synthetic secret canary. If it does not, the compatibility test fails because a redaction assertion would otherwise be meaningless.

After strict normalization, the same canary must be absent from the serialized `ChangeSet` for every engine/version combination. The test also requires strict redaction mode, removal of source variable values, and removal of resource/output values independently of producer sensitivity-mask behavior.

## Container model

Terraform uses versioned HashiCorp images.

For OpenTofu 1.10 and newer, the harness follows OpenTofu's current container guidance: it copies the `tofu` binary from the official `*-minimal` image into a small runtime image instead of relying on the legacy direct-image execution contract.

Both compatibility images run as an unprivileged UID. The workspace is mounted read-only into the engine containers. Each runner copies it into an ephemeral directory before `init`, `validate`, `plan`, and `show -json`. Only generated JSON artifacts are written back to the repository's ignored artifact directory.
