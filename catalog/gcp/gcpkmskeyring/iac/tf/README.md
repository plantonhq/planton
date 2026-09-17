# GcpKmsKeyRing - Terraform Module

This Terraform/OpenTofu module provisions a Cloud KMS key ring (`google_kms_key_ring`) — the permanent container for cryptographic keys. It is the Terraform-side implementation of the Planton `GcpKmsKeyRing` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Cloud KMS API with `disable_on_destroy=false` so a fresh project works first try and teardown never disables the API project-wide. An empty `project_id` falls back to the provider's default project, identically to the Pulumi module. The key ring API has no labels surface, so no platform attribution labels are stamped.

**Destroy is state-only by GCP design**: key rings have no delete API. `destroy` removes the ring from state and leaves the (free, inert) ring in the project permanently — its name can never be reused there.

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
- `locals.tf` — ambient-project resolution
- `main.tf` — API enablement + the key ring resource
- `outputs.tf` — stack outputs (must match `outputs.proto`)

## Inputs

| Variable | Description |
|----------|-------------|
| `spec.project_id` | GCP project (empty = provider default project). ForceNew |
| `spec.key_ring_name` | The permanent ring name. ForceNew, never reusable |
| `spec.location` | Region, multi-region, or `global`. ForceNew |

## Outputs

| Name | Description |
|------|-------------|
| `key_ring_id` | Fully qualified ring path (`projects/{p}/locations/{l}/keyRings/{name}`) — the exact string a `GcpKmsKey`'s `keyRingId` reference consumes |
| `key_ring_name` | The short name of the ring |
| `location` | The ring's location |

## Lifecycle Notes

Every field is ForceNew, which for an undeletable resource means "abandon and create anew" — plan names and locations as permanent decisions. IAM granted on the ring flows down to every key inside it: group keys by environment or data domain, not one ring per key.
