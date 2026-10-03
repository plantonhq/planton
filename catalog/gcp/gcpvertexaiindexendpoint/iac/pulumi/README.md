# GcpVertexAiIndexEndpoint - Pulumi Module

This Pulumi (Go) module provisions a Vertex AI Vector Search index endpoint (`vertex.AiIndexEndpoint`). It is the Pulumi-side implementation of the Planton `GcpVertexAiIndexEndpoint` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Vertex AI API (`aiplatform.googleapis.com`, `disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

The peered `network` value is normalized from a compute self-link URL (the `GcpVpcNetwork` reference's canonical output) to the relative `projects/{project}/global/networks/{name}` form the Vertex AI API requires — identical normalization to the Terraform module.

The bridged provider's client-side `deletion_policy` knob is pinned to `DELETE` so destroy behavior is byte-identical to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpvertexaiindexendpoint/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/locals.go` — labels merge + network self-link normalization
- `module/index_endpoint.go` — API enablement + the index endpoint resource + outputs
- `module/outputs.go` — output constant names
