# GcpArtifactRegistryRepo - Pulumi Module

This Pulumi (Go) module provisions an Artifact Registry repository (`artifactregistry.Repository`) plus additive per-repository IAM grants. It is the Pulumi-side implementation of the Planton `GcpArtifactRegistryRepo` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Artifact Registry API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module. The repository ID falls back to `metadata.name`; all three serving modes (standard, remote, virtual) are driven from the spec, with the mode↔config coherence enforced pre-deploy by the spec's CEL rules.

The `deletion_policy` spec field (DELETE / PREVENT / ABANDON, provider default DELETE) is wired identically to the Terraform module — sent only when set. The `registry_uri` output is constructed from resolved attributes — the released provider exports no registry URI attribute — using the exact expression the Terraform module uses.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpartifactregistryrepo/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — Pulumi entrypoint (loads the IaC input, calls the module)
- `module/main.go` — provider setup + orchestration
- `module/locals.go` — repository-ID fallback + label merge
- `module/repository.go` — API enablement + the repository + additive IAM members + outputs
- `module/outputs.go` — output key constants

## Outputs

| Output | Description |
|--------|-------------|
| `name` | Short repository name |
| `repository_path` | Fully qualified repository resource path |
| `registry_uri` | Push/pull endpoint |
| `location` | Repository location |
