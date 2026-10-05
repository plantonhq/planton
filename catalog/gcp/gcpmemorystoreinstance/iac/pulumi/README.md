# GcpMemorystoreInstance - Pulumi Module

This Pulumi (Go) module provisions a Memorystore (Valkey) instance (`memorystore.Instance`) — the new-generation, PSC-first in-memory data store. It is the Pulumi-side implementation of the Planton `GcpMemorystoreInstance` resource kind and has feature parity with the Terraform module.

## Overview

Connectivity is PSC-only and driven by service connectivity automation: a `GcpServiceConnectionPolicy` for the `gcp-memorystore` class must exist on each endpoint's network in the instance's region before the instance is created. The module enables the Memorystore and Network Connectivity APIs so a fresh project works first try.

Deletion protection defaults TRUE in the spec and is always sent explicitly, so destroy behavior is identical on both engines. A PSC endpoint entry that omits its consumer project rides the provider's effective project (resolved via `GetClientConfig`, mirroring the Terraform module's `data.google_project`).

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Pulumi Usage

```bash
cd catalog/gcp/gcpmemorystoreinstance/iac/pulumi
make build
pulumi up --stack dev
```

## Module Layout

- `main.go` — entrypoint; loads the IaC input and calls the module
- `module/main.go` — provider setup and resource orchestration
- `module/locals.go` — metadata-derived values and the label merge
- `module/memorystore_instance.go` — API enablement + the instance resource + discovery-endpoint extraction
- `module/outputs.go` — output keys (must match `outputs.proto`)

## Outputs

| Name | Description |
|------|-------------|
| `discovery_address` | IP address of the PSC discovery endpoint |
| `discovery_port` | Port of the discovery endpoint (typically 6379) |
| `instance_uid` | Server-generated unique identifier |
| `node_size_gb` | Memory per node in GB (a consequence of `node_type`) |
| `name` | Full resource path (the DR composition key) |
| `backup_collection` | Where automated backups land |

## Lifecycle Notes

The immutables (ForceNew): instance id, location, mode, auth/TLS modes, CMEK key, zone distribution, the PSC endpoints, and the seed sources. Shards, replicas, engine configs, persistence, maintenance, backups, labels, and the DR role update in place.
