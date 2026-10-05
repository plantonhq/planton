# GcpKmsKeyRing - Pulumi Module

This Pulumi (Go) module provisions a Cloud KMS key ring (`kms.KeyRing`) — the permanent container for cryptographic keys. It is the Pulumi-side implementation of the Planton `GcpKmsKeyRing` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Cloud KMS API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `projectId` falls back to the provider's default project, identically to the Terraform module. The key ring API has no labels surface, so no platform attribution labels are stamped.

**Destroy is state-only by GCP design**: key rings have no delete API. `destroy` removes the ring from state and leaves the (free, inert) ring in the project permanently — its name can never be reused there.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpkmskeyring/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — iac-input plumbing (no labels: the API has no labels surface)
- `module/key_ring.go` — API enablement + the key ring resource + exports
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `key_ring_id` | Fully qualified ring path (`projects/{p}/locations/{l}/keyRings/{name}`) — the exact string a `GcpKmsKey`'s `keyRingId` reference consumes |
| `key_ring_name` | The short name of the ring |
| `location` | The ring's location |

## Lifecycle Notes

Every field is immutable, which for an undeletable resource means "abandon and create anew" — plan names and locations as permanent decisions. IAM granted on the ring flows down to every key inside it: group keys by environment or data domain, not one ring per key.
