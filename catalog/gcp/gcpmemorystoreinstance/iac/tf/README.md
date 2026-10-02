# GcpMemorystoreInstance - Terraform Module

This Terraform module provisions a Memorystore (Valkey) instance (`google_memorystore_instance`) — the new-generation, PSC-first in-memory data store. It is the Terraform-side implementation of the Planton `GcpMemorystoreInstance` resource kind and has feature parity with the Pulumi module.

## Overview

Connectivity is PSC-only and driven by service connectivity automation: a `GcpServiceConnectionPolicy` for the `gcp-memorystore` class must exist on each endpoint's network in the instance's region before the instance is created. The module enables the Memorystore and Network Connectivity APIs so a fresh project works first try, and runs on the plain `google` provider — every modeled field is GA at the pinned release (`~> 8.3`).

Deletion protection defaults TRUE in the spec and is always sent explicitly, so destroy behavior is identical on both engines. A PSC endpoint entry that omits its consumer project rides the provider's effective project (resolved via `data.google_project`, mirroring the Pulumi module's `GetClientConfig`).

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via stack input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpmemorystoreinstance/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpMemorystoreInstance spec | — |

The `spec` object includes: identity (`instance_name`, `location`, `project_id`), topology (`shard_count`, `mode`, `node_type`, `replica_count`, `zone_distribution_config`), engine (`engine_version`, `engine_configs`), security (`authorization_mode`, `transit_encryption_mode`, `kms_key`), durability (`persistence_config`, `automated_backup_config`), networking (`psc_auto_connections`), DR (`cross_instance_replication_config`), seeding (`gcs_source` XOR `managed_backup_source`), `labels`, and `deletion_protection_enabled` (default true).

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
