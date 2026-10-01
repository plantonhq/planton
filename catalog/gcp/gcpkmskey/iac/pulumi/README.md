# GcpKmsKey - Pulumi Module

This Pulumi (Go) module provisions a Cloud KMS crypto key (`kms.CryptoKey`) inside an existing key ring. It is the Pulumi-side implementation of the Planton `GcpKmsKey` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Cloud KMS API (project extracted from the ring path, `disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module.

**Destroy destroys versions, not the key**: crypto keys have no delete API. Under the default `deletion_policy` (DELETE), `destroy` schedules every key version for destruction (data encrypted under them becomes unrecoverable once the recovery window elapses), disables automatic rotation, and removes the key from state — the key object remains permanently in the ring and its name can never be reused there. The spec's `deletion_policy` field offers PREVENT (destroy fails) and ABANDON (the key leaves management with every version intact), wired identically to the Terraform module.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpkmskey/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the stack input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — the user-label + attribution-label merge
- `module/kms_key.go` — API enablement + the crypto key resource + exports
- `module/outputs.go` — stack output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `key_id` | Fully qualified key path — the CMEK reference every consumer takes |
| `key_name` | The short name of the key |
| `primary_version_name` | Current primary version resource name (ENCRYPT_DECRYPT keys; empty otherwise) |
| `primary_state` | Lifecycle state of the primary version |
| `initial_version_name` | The version created with the key (`.../cryptoKeyVersions/1`), for every purpose -- what asymmetric-sign consumers name; empty with `skip_initial_version_creation` |

## Required Permissions

See [`../permissions.yaml`](../permissions.yaml) for the least-privilege permission set the deploying principal needs.

## Lifecycle Notes

Only `rotationPeriod`, `versionTemplate.algorithm`, and `labels` update in place; every other field is immutable, which for an undeletable resource means "abandon and create under a new name". Grant each consuming service agent `roles/cloudkms.cryptoKeyEncrypterDecrypter` on the key before pointing CMEK fields at it.
