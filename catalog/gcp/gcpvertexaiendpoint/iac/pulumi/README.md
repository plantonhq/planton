# GcpVertexAiEndpoint - Pulumi Module

This Pulumi (Go) module provisions a Vertex AI Endpoint (`vertex.AiEndpoint`). It is the Pulumi-side implementation of the Planton `GcpVertexAiEndpoint` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Vertex AI API (`aiplatform.googleapis.com`, `disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

**The numeric endpoint name is always sent explicitly.** Vertex AI requires a numeric-only endpoint ID (max 10 digits, no leading zero) and never generates one; engine auto-naming would produce a non-numeric name the API rejects. When `spec.endpoint_name` is empty, the module derives a stable ID from the resource identity — the identical derivation as the Terraform module, so the same manifest yields the same endpoint ID on either engine.

The bridged provider's client-side `deletion_policy` knob is pinned to `DELETE` so destroy behavior is byte-identical to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpvertexaiendpoint/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/locals.go` — labels merge + endpoint-name resolution
- `module/endpoint_name.go` — the identity-based numeric ID derivation (kept in lockstep with the Terraform module)
- `module/ai_endpoint.go` — API enablement + the endpoint resource + outputs
- `module/outputs.go` — output constant names
