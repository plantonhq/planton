# GcpFilestoreInstance - Terraform Module

This Terraform/OpenTofu module provisions a Filestore instance (`google_filestore_instance`) with its single file share and VPC attachment. It is the Terraform-side implementation of the Planton `GcpFilestoreInstance` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Filestore API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Pulumi module. An empty `instance_name` falls back to `metadata.name`; an empty `project_id` falls back to the provider's default project; empty `modes` becomes `["MODE_IPV4"]` — all via explicit conditionals so both engines realize the identical instance.

**Immutability is the sharp edge**: name, location, tier, protocol, network attachment, KMS key, LDAP directory services, and replication are all ForceNew — changing any of them replaces the instance and its data. File share capacity grows in place but never shrinks. `deletion_protection_enabled` is the destroy guard (it must be flipped false before a protected instance can be destroyed); `deletion_policy` is the second, client-side lever evaluated after it (empty/`DELETE` deletes, `PREVENT` refuses, `ABANDON` keeps the instance running outside management). The restore sources (`source_backup` / `source_backupdr_backup`, mutually exclusive) and `initial_replication` apply at create time only; `desired_replica_state` pauses/resumes an existing replica relationship in place.

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
- `locals.tf` — instance-name and project fallbacks, empty-string→null normalization, modes default, label merge
- `main.tf` — API enablement + the instance (file share, network incl. PSC config, LDAP directory services, performance config, replication)
- `outputs.tf` — `instance_id`, `instance_name`, `ip_addresses`, `file_share_name`, `create_time`, `reserved_ip_range`, `etag`

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | Fully qualified resource ID (`projects/{p}/locations/{l}/instances/{i}`) |
| `instance_name` | Short name of the instance |
| `ip_addresses` | IP addresses on the VPC network (use the first for NFS mounts) |
| `file_share_name` | File share name for the NFS mount path |
| `create_time` | Instance creation timestamp (RFC3339) |
| `reserved_ip_range` | The `/29` block as resolved by GCP (also populated when auto-picked) |
| `etag` | Server-specified ETag guarding concurrent updates |
