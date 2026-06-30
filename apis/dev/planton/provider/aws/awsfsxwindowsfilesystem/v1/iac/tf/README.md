# Terraform Module: AwsFsxWindowsFileSystem

## Overview

This Terraform module provisions an Amazon FSx for Windows File Server file system. It supports the full FSx for Windows feature set including deployment types (Single-AZ and Multi-AZ), Active Directory integration (AWS Managed and self-managed), audit logging, disk IOPS configuration, backup policies, DNS aliases, and encryption with customer-managed KMS keys.

## Quick Start

Run the module via the Planton CLI (tofu) using the default local backend.

```bash
planton tofu init --manifest hack/manifest.yaml
planton tofu plan --manifest hack/manifest.yaml
planton tofu apply --manifest hack/manifest.yaml --auto-approve
planton tofu destroy --manifest hack/manifest.yaml --auto-approve
```

- Credentials are provided via stack input (by the CLI), not in the manifest `spec`.
- Manifest file: `../hack/manifest.yaml`

## File Structure

```
iac/tf/
├── provider.tf    # AWS provider and Terraform version constraints
├── variables.tf   # Input variables (provider_config, metadata, spec)
├── locals.tf      # Tag construction from metadata
├── main.tf        # FSx Windows File System resource with dynamic blocks
└── outputs.tf     # Eight outputs matching AwsFsxWindowsFileSystemStackOutputs
```

## Resources Created

- `aws_fsx_windows_file_system.this` — the FSx for Windows File Server file system with all configured options

## Variables

### `provider_config`

AWS provider configuration for authentication and region.

| Field               | Type   | Required | Description                    |
|---------------------|--------|----------|--------------------------------|
| `region`            | string | yes      | AWS region for the file system |
| `access_key_id`     | string | no       | AWS access key ID              |
| `secret_access_key` | string | no       | AWS secret access key          |
| `session_token`     | string | no       | AWS session token              |

### `metadata`

Resource metadata used for tagging and naming.

| Field  | Type   | Required | Description                          |
|--------|--------|----------|--------------------------------------|
| `org`  | string | yes      | Organization identifier              |
| `env`  | string | yes      | Environment (dev/staging/prod)       |
| `name` | string | yes      | Resource name (used as FSx name)     |
| `id`   | string | yes      | Unique resource identifier           |

### `spec`

The FSx for Windows File Server specification.

| Field                              | Type   | Default       | Description |
|------------------------------------|--------|---------------|-------------|
| `deployment_type`                  | string | `SINGLE_AZ_2` | `SINGLE_AZ_1`, `SINGLE_AZ_2`, or `MULTI_AZ_1` |
| `storage_capacity_gib`            | number | (required)    | Storage capacity in GiB (min 32 for SSD, 2000 for HDD) |
| `storage_type`                     | string | `SSD`         | `SSD` or `HDD` |
| `throughput_capacity`              | number | (required)    | Throughput in MB/s (8, 16, 32, 64, 128, 256, 512, 1024, 2048) |
| `subnet_ids`                       | list   | (required)    | Subnet IDs (1 for Single-AZ, 2 for Multi-AZ) |
| `preferred_subnet_id`             | string | —             | Preferred subnet for Multi-AZ active file server |
| `security_group_ids`              | list   | `[]`          | Security group IDs for network access control |
| `kms_key_id`                       | string | —             | KMS key ARN for encryption at rest (omit for AWS-managed) |
| `active_directory_id`             | string | —             | AWS Managed Microsoft AD directory ID |
| `self_managed_active_directory`   | object | —             | Self-managed AD config (see below) |
| `aliases`                          | list   | `[]`          | DNS aliases (CNAME records) |
| `audit_log_configuration`         | object | —             | File access audit logging (see below) |
| `disk_iops_configuration`         | object | —             | Disk IOPS settings (see below) |
| `automatic_backup_retention_days` | number | `7`           | Days to retain automatic backups (0 disables) |
| `daily_automatic_backup_start_time` | string | —           | Backup start time in `HH:MM` UTC format |
| `copy_tags_to_backups`            | bool   | `false`       | Propagate tags to backup copies |
| `skip_final_backup`               | bool   | `true`        | Skip final backup on resource deletion |
| `weekly_maintenance_start_time`   | string | —             | Maintenance window in `d:HH:MM` format |

#### `self_managed_active_directory` (nested object)

| Field                                      | Type   | Default          | Description |
|--------------------------------------------|--------|------------------|-------------|
| `domain_name`                              | string | (required)       | Fully qualified domain name |
| `dns_ips`                                  | list   | (required)       | DNS server IP addresses |
| `username`                                 | string | —                | Service account username (plaintext path) |
| `password`                                 | string | —                | Service account password (plaintext path) |
| `domain_join_service_account_secret_arn`   | string | —                | Secrets Manager ARN for credentials |
| `file_system_administrators_group`         | string | `Domain Admins`  | Windows group with admin rights |
| `organizational_unit_distinguished_name`   | string | —                | OU DN for computer object placement |

#### `audit_log_configuration` (nested object)

| Field                              | Type   | Default    | Description |
|------------------------------------|--------|------------|-------------|
| `file_access_audit_log_level`      | string | `DISABLED` | `DISABLED`, `SUCCESS_ONLY`, `FAILURE_ONLY`, `SUCCESS_AND_FAILURE` |
| `file_share_access_audit_log_level` | string | `DISABLED` | Same values as above |
| `audit_log_destination`            | string | —          | CloudWatch Logs group ARN or Firehose stream ARN |

#### `disk_iops_configuration` (nested object)

| Field  | Type   | Default     | Description |
|--------|--------|-------------|-------------|
| `mode` | string | `AUTOMATIC` | `AUTOMATIC` or `USER_PROVISIONED` |
| `iops` | number | —           | IOPS value (required when mode is `USER_PROVISIONED`) |

## Outputs

| Output                             | Description                                       |
|------------------------------------|---------------------------------------------------|
| `file_system_id`                   | The ID of the file system                         |
| `file_system_arn`                  | The Amazon Resource Name of the file system       |
| `dns_name`                         | DNS name for mounting via SMB                     |
| `preferred_file_server_ip`         | IP address of the preferred file server           |
| `remote_administration_endpoint`   | Endpoint for remote PowerShell administration     |
| `network_interface_ids`            | Network interface IDs created for the file system |
| `vpc_id`                           | VPC ID in which the file system was created       |
| `owner_id`                         | AWS account ID of the file system owner           |

## Dynamic Blocks Explained

The `main.tf` uses three Terraform `dynamic` blocks to conditionally include nested configuration objects. This pattern avoids errors from setting empty/null nested blocks.

### `self_managed_active_directory`

```hcl
dynamic "self_managed_active_directory" {
  for_each = var.spec.self_managed_active_directory != null ? [var.spec.self_managed_active_directory] : []
  content { ... }
}
```

**When included**: Only when the user provides a `self_managed_active_directory` object in the spec. Mutually exclusive with `active_directory_id`.

**Key detail**: Supports both credential paths—plaintext `username`/`password` or `domain_join_service_account_secret` (Secrets Manager ARN). The Secrets Manager path is preferred for production.

### `audit_log_configuration`

```hcl
dynamic "audit_log_configuration" {
  for_each = var.spec.audit_log_configuration != null ? [var.spec.audit_log_configuration] : []
  content { ... }
}
```

**When included**: Only when audit logging is configured. If the object is null, no audit log block is emitted and FSx uses its default (no auditing).

**Important**: The `audit_log_destination` must be a valid CloudWatch Logs group ARN or Kinesis Data Firehose delivery stream ARN. If omitted, audit events are emitted but not persisted.

### `disk_iops_configuration`

```hcl
dynamic "disk_iops_configuration" {
  for_each = var.spec.disk_iops_configuration != null ? [var.spec.disk_iops_configuration] : []
  content { ... }
}
```

**When included**: Only when custom IOPS configuration is desired. If null, AWS uses `AUTOMATIC` mode (3 IOPS per GiB for SSD).

**Important**: When `mode` is `USER_PROVISIONED`, the `iops` field is required. When `mode` is `AUTOMATIC`, `iops` should be omitted.

## Lifecycle Configuration

```hcl
lifecycle {
  ignore_changes = [tags["CreatedAt"]]
}
```

The `CreatedAt` tag is excluded from change detection to prevent unnecessary updates when AWS auto-populates this tag on resource creation.

## Tags

All file systems are tagged with five standard Planton metadata tags (defined in `locals.tf`):

| Tag            | Value                         |
|----------------|-------------------------------|
| `Resource`     | `true`                        |
| `Organization` | From `metadata.org`           |
| `Environment`  | From `metadata.env`           |
| `ResourceKind` | `AwsFsxWindowsFileSystem`     |
| `ResourceId`   | From `metadata.id`            |
