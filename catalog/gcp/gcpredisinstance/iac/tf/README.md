# GcpRedisInstance - Terraform Module

This Terraform module provisions a Google Cloud Memorystore for Redis instance (`google_redis_instance`). It is the Terraform-side implementation of the Planton `GcpRedisInstance` resource kind and has feature parity with the Pulumi module.

## Overview

The module enables the Memorystore API (`redis.googleapis.com`) so a fresh project works first try, then creates the instance with the full released-provider surface: tier and memory, zone pinning (primary + HA replica), VPC connectivity (direct peering or private services access), AUTH and TLS, RDB persistence with a schedule anchor, read replicas with in-place address-space growth, maintenance window to the minute, self-service maintenance version, CMEK, and Redis config overrides.

An empty `project_id` falls back to the provider's default project; empty optional strings become null so the provider omits them from the API payload. The module runs on the plain `google` provider — every modeled field is GA at the pinned release (`~> 8.3`).

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
cd catalog/gcp/gcpredisinstance/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `metadata` | Resource metadata (name, labels, etc.) | — |
| `spec` | GcpRedisInstance spec | — |

The `spec` object includes: `instance_name` + `region` + `tier` + `memory_size_gb` (required), `redis_version`, `display_name`, `location_id` + `alternative_location_id` (zone pinning), `authorized_network` (VPC self link), `connect_mode`, `reserved_ip_range` + `secondary_ip_range`, `auth_enabled`, `transit_encryption_mode`, `redis_configs`, `maintenance_window` (day/hour/minute) + `maintenance_version`, `read_replicas_mode` + `replica_count`, `persistence_config` (mode/period/start time), `customer_managed_key` (KMS key id), and `project_id` (empty falls back to the provider default project).

## Outputs

| Name | Description |
|------|-------------|
| `host` / `port` | Primary Redis endpoint |
| `read_endpoint` / `read_endpoint_port` | Read replica endpoint (STANDARD_HA with read replicas) |
| `current_location_id` | Zone of the primary (may change after failover) |
| `auth_string` | Redis AUTH string — sensitive |
| `server_ca_certs` | PEM CA certificates clients trust for TLS (when transit encryption is on) |
| `persistence_iam_identity` | Service identity to grant on GCS buckets for RDB import/export |
| `effective_reserved_ip_range` | The CIDR actually in use — explicit or GCP-selected |
| `instance_name` | Instance name in GCP |
| `region` | Plain region name hosting the instance |
