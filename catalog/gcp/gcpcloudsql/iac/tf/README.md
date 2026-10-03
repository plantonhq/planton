# GcpCloudSql - Terraform Module

This Terraform module provisions a Cloud SQL instance (`google_sql_database_instance`). It is the Terraform-side implementation of the Planton `GcpCloudSql` resource kind and has feature parity with the Pulumi module.

## Overview

The module creates a managed MySQL, PostgreSQL, or SQL Server instance — a primary, a read replica (with `master_instance_name`), or a read pool (`instance_type: READ_POOL_INSTANCE` with `node_count` / `read_pool_auto_scale`) — with edition/HA, disk tuning (including hyperdisk provisioned IOPS/throughput), explicit connectivity (public IPv4 / private VPC IP / Private Service Connect with DNS automation), backups with point-in-time recovery and a final backup on delete, restore provenance (clone / backup-run restore / Backup-and-DR point-in-time restore), cross-region DR pairing, maintenance scheduling and version pinning, Query Insights (standard and enhanced), password policies, managed connection pooling, Active Directory and Entra ID authentication for SQL Server, CMEK, both delete guards, and the `deletion_policy` teardown lever. It also enables the Cloud SQL Admin API (`sqladmin.googleapis.com`) so a fresh project works on the first deploy.

## Usage with Planton CLI

```shell
planton tofu init --manifest ../../e2e/manifest.yaml
planton tofu plan --manifest ../../e2e/manifest.yaml
planton tofu apply --manifest ../../e2e/manifest.yaml --auto-approve
planton tofu destroy --manifest ../../e2e/manifest.yaml --auto-approve
```

Credentials are provided via IaC input (by the CLI), not in the manifest `spec`. Manifest file: `../../e2e/manifest.yaml`.

## Direct Terraform Usage

```bash
cd catalog/gcp/gcpcloudsql/iac/tf
terraform init
terraform plan -var-file=terraform.tfvars.json
terraform apply -var-file=terraform.tfvars.json
```

## Variables

| Name | Description | Default |
|------|-------------|---------|
| `spec` | GcpCloudSql spec (instance_name, region, engine/version/tier, disk, network, backup, …) | — |
| `metadata` | Resource metadata (name, org, env, labels) | — |

The `spec` object mirrors the proto spec; `StringValueOrRef` fields (`project_id`, `network.private_network`, `encryption_key_name`, `master_instance_name`) arrive as plain strings after the CLI's ref resolution.

## Outputs

| Name | Description |
|------|-------------|
| `instance_name` | The composition key databases, users, and replicas reference |
| `connection_name` | `project:region:instance` for the Auth Proxy and connectors |
| `private_ip` | Private IP address (empty unless private_network is configured) |
| `public_ip` | Public IPv4 address (empty unless ipv4_enabled) |
| `self_link` | Self-link URL of the Cloud SQL instance |
| `service_account_email` | The instance's Google-managed service account |
| `dns_name` | DNS name (PSC-enabled instances) |
| `psc_service_attachment_link` | PSC service attachment link (PSC-enabled instances) |
