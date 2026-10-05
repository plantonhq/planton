# GcpRedisInstance - Pulumi Module

This Pulumi (Go) module provisions a Google Cloud Memorystore for Redis instance (`redis.Instance`). It is the Pulumi-side implementation of the Planton `GcpRedisInstance` resource kind and has feature parity with the Terraform module.

## Overview

The module enables the Memorystore API (`redis.googleapis.com`) so a fresh project works first try, then creates the instance with the full released-provider surface: tier and memory, zone pinning (primary + HA replica), VPC connectivity (direct peering or private services access), AUTH and TLS, RDB persistence with a schedule anchor, read replicas with in-place address-space growth, maintenance window to the minute, self-service maintenance version, CMEK, and Redis config overrides.

An empty `project_id` falls back to the provider's default project — the ambient-project contract every GCP kind honors. The AUTH string is exported as a Pulumi secret, so it is encrypted in state and masked in console output.

## Usage with Planton CLI

```shell
planton pulumi up --manifest ../../e2e/manifest.yaml --module-dir .
planton pulumi destroy --manifest ../../e2e/manifest.yaml --module-dir .
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Build

```bash
cd catalog/gcp/gcpredisinstance/iac/pulumi
make build
```

## Outputs

| Name | Description |
|------|-------------|
| `host` / `port` | Primary Redis endpoint |
| `read_endpoint` / `read_endpoint_port` | Read replica endpoint (STANDARD_HA with read replicas) |
| `current_location_id` | Zone of the primary (may change after failover) |
| `auth_string` | Redis AUTH string — exported as a secret |
| `server_ca_certs` | PEM CA certificates clients trust for TLS (when transit encryption is on) |
| `persistence_iam_identity` | Service identity to grant on GCS buckets for RDB import/export |
| `effective_reserved_ip_range` | The CIDR actually in use — explicit or GCP-selected |
| `instance_name` | Instance name in GCP |
| `region` | Plain region name hosting the instance |
