# GcpFilestoreInstance - Pulumi Module

This Pulumi (Go) module provisions a Filestore instance (`filestore.Instance`) with its single file share and VPC attachment. It is the Pulumi-side implementation of the Planton `GcpFilestoreInstance` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Filestore API (`disable_on_destroy=false`) so a fresh project works first try and teardown never disables the API project-wide. User labels are merged beneath the platform attribution labels (`planton-ai_*`), identically to the Terraform module. An empty `instance_name` falls back to `metadata.name`; an empty `project_id` falls back to the provider's default project; empty `modes` becomes `["MODE_IPV4"]` — all via explicit conditionals so both engines realize the identical instance. The bridged provider's client-side `deletion_policy` flag is pinned to `DELETE` so destroy semantics match Terraform exactly (`deletion_protection_enabled` remains the real guard).

**Immutability is the sharp edge**: name, location, tier, protocol, network attachment, KMS key, and replication are all replace-on-change — changing any of them replaces the instance and its data. File share capacity grows in place but never shrinks. `deletion_protection_enabled` is the destroy guard: it must be flipped false before a protected instance can be destroyed. `source_backup` and `initial_replication` apply at create time only.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpfilestoreinstance/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — Pulumi entrypoint (loads the IaC input, calls the module)
- `module/main.go` — provider setup + orchestration
- `module/locals.go` — instance-name fallback + label merge
- `module/filestore_instance.go` — API enablement + the instance (file share, network, performance config, replication) + outputs
- `module/outputs.go` — output key constants

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
