# GcpVertexAiDeployedIndex - Pulumi Module

This Pulumi (Go) module deploys a Vertex AI Vector Search index onto an index endpoint (`vertex.AiIndexEndpointDeployedIndex`). It is the Pulumi-side implementation of the Planton `GcpVertexAiDeployedIndex` resource kind and has feature parity with the Terraform module.

## Overview

The module creates the deployment only — no API enablement (the referenced index endpoint cannot exist without the Vertex AI API already on) and no labels (the GCP resource class carries none; platform attribution is impossible here and none is faked).

Only the replica bounds inside the sizing arm update in place after deployment; every other change undeploys and redeploys. Deploying is a long-running operation (tens of minutes).

The bridged provider's client-side `deletion_policy` knob is pinned to `DELETE` so destroy behavior is byte-identical to the Terraform module. The private-endpoint outputs (`match_grpc_address`, `service_attachment`) are exported as empty strings on public endpoints so the output shape is stable — identical to the Terraform module's `try()` fallbacks.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpvertexaideployedindex/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/locals.go` — target unwrapping (no label merge: the resource class has no labels)
- `module/deployed_index.go` — sizing/auth/range assembly + the deployment resource + outputs
- `module/outputs.go` — output constant names
