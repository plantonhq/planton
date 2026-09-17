# GcpArtifactRegistryRepo - Terraform Module

This Terraform/OpenTofu module provisions an Artifact Registry repository (`google_artifact_registry_repository`) plus additive per-repository IAM grants. It is the Terraform-side implementation of the Planton `GcpArtifactRegistryRepo` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Artifact Registry API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Pulumi module. The repository ID falls back to `metadata.name`; all three serving modes (standard, remote, virtual) are driven from the spec, with the mode↔config coherence enforced pre-deploy by the spec's CEL rules.

**Immutability is the sharp edge**: `format`, `mode`, `location`, `project`, and `kms_key_name` are ForceNew — changing any of them replaces the repository and every artifact in it. Remote-repository upstream credentials rotate in place; the upstream itself does not.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu plan --manifest ../../e2e/manifest.yaml --module-dir .
planton tofu apply --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --module-dir . --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Module Layout

- `provider.tf` — google provider pin (`~> 8.3`; all fields GA on the released line)
- `variables.tf` — the converter-contract `metadata`/`spec` variables
- `locals.tf` — ambient-project fallback, repository-ID fallback, label merge, IAM grant keying
- `main.tf` — API enablement + the repository + additive IAM members
- `outputs.tf` — `name`, `repository_path`, `registry_uri` (constructed — the released provider exports no registry URI attribute), `location`

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Short repository name |
| `repository_path` | Fully qualified repository resource path |
| `registry_uri` | Push/pull endpoint |
| `location` | Repository location |
