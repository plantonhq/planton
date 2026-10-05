# GcpVertexAiIndex - Pulumi Module

This Pulumi (Go) module provisions a Vertex AI Vector Search index (`vertex.AiIndex`). It is the Pulumi-side implementation of the Planton `GcpVertexAiIndex` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Vertex AI API (`aiplatform.googleapis.com`, `disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

The spec's flattened `contents_delta_uri` / `is_complete_overwrite` / `config` fields are reassembled into the provider's nested `metadata` block. `config` always exists (required by the spec — the API rejects an index without it); the algorithm arm is sent only when the spec picks one, letting GCP default to tree-AH otherwise.

The bridged provider's client-side `deletion_policy` knob is pinned to `DELETE` so destroy behavior is byte-identical to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpvertexaiindex/iac/pulumi
make build
pulumi up --stack dev
```

Note: index creation is a long-running operation — minutes for an empty streaming index, up to hours for a large batch build.

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/locals.go` — labels merge
- `module/vector_index.go` — API enablement + metadata/config assembly + the index resource + outputs
- `module/outputs.go` — output constant names
