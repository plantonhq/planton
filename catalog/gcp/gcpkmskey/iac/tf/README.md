# GcpKmsKey - Terraform Module

This Terraform/OpenTofu module provisions a Cloud KMS crypto key (`google_kms_crypto_key`) inside an existing key ring. It is the Terraform-side implementation of the Planton `GcpKmsKey` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Cloud KMS API (project extracted from the ring path, `disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Pulumi module.

**Destroy destroys versions, not the key**: crypto keys have no delete API. `destroy` schedules every key version for destruction (data encrypted under them becomes unrecoverable once the recovery window elapses), disables automatic rotation, and removes the key from state — the key object remains permanently in the ring and its name can never be reused there.

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
- `locals.tf` — optional-field null-folding + the label merge
- `main.tf` — API enablement + the crypto key resource
- `outputs.tf` — stack outputs (must match `outputs.proto`)

## Inputs

| Variable | Description |
|----------|-------------|
| `spec.key_ring_id` | Fully qualified ring path (resolved from a `GcpKmsKeyRing` reference). ForceNew |
| `spec.key_name` | The permanent key name. ForceNew, never reusable within the ring |
| `spec.purpose` | Key purpose (empty = `ENCRYPT_DECRYPT`). ForceNew |
| `spec.rotation_period` | Auto-rotation cadence (ENCRYPT_DECRYPT only). Mutable |
| `spec.destroy_scheduled_duration` | Recovery window for destroyed versions. ForceNew |
| `spec.version_template` | Algorithm (mutable) + protection level (ForceNew) |
| `spec.skip_initial_version_creation` | Create the key empty. Create-time only |
| `spec.import_only` | BYOK container. ForceNew |
| `spec.crypto_key_backend` | EKM connection for `EXTERNAL_VPC` keys. ForceNew |
| `spec.labels` | User labels, merged beneath `planton-ai_*` attribution labels |

## Outputs

| Name | Description |
|------|-------------|
| `key_id` | Fully qualified key path — the CMEK reference every consumer takes |
| `key_name` | The short name of the key |
| `primary_version_name` | Current primary version resource name (ENCRYPT_DECRYPT keys; empty otherwise) |
| `primary_state` | Lifecycle state of the primary version |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.

## Lifecycle Notes

Only `rotation_period`, `version_template.algorithm`, and `labels` update in place; every other field is ForceNew, which for an undeletable resource means "abandon and create under a new name". Grant each consuming service agent `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key before pointing CMEK fields at it.
