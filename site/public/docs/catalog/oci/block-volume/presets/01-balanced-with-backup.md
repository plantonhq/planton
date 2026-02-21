---
title: "Balanced with Backup"
description: "This preset creates a 100 GB OCI Block Volume at the Balanced performance tier (10 VPUs/GB) with an assigned backup policy and a detached-volume autotune policy. This is the standard configuration..."
type: "preset"
rank: "01"
presetSlug: "01-balanced-with-backup"
componentSlug: "block-volume"
componentTitle: "Block Volume"
provider: "oci"
icon: "package"
order: 1
---

# Balanced with Backup

This preset creates a 100 GB OCI Block Volume at the Balanced performance tier (10 VPUs/GB) with an assigned backup policy and a detached-volume autotune policy. This is the standard configuration for application data volumes -- it provides consistent I/O performance for general workloads, automatic scheduled backups for data protection, and cost savings when the volume is temporarily detached.

## When to Use

- Application data volumes for compute instances (logs, application state, user uploads)
- Volumes that need scheduled backups without manual intervention (use Oracle-defined Gold, Silver, or Bronze backup policies)
- General-purpose workloads where 60 IOPS/GB and 480 KB/s/GB throughput per GB is sufficient
- Volumes that may be detached during maintenance windows and should not incur full performance charges while idle

## Key Configuration Choices

- **100 GB** (`sizeInGbs: 100`) -- a practical starting size for application data. Block volumes can be resized online (increased only, not shrunk) without downtime. The minimum is 50 GB; maximum is 32 TB. Start small and grow based on actual usage.
- **Balanced performance tier** (`vpusPerGb: 10`) -- provides 60 IOPS/GB (6,000 total IOPS for a 100 GB volume) and 480 KB/s/GB throughput. This is the OCI default and suits the vast majority of workloads. For I/O-intensive databases, use the Higher Performance preset instead.
- **Backup policy assignment** (`backupPolicyId`) -- OCI provides three Oracle-defined backup policies: Gold (daily + weekly + monthly + yearly), Silver (daily + weekly + monthly), and Bronze (daily + weekly). Custom backup policies can also be created. The policy runs scheduled backups automatically without application downtime.
- **Detached-volume autotune** (`autotunePolicies` with `detached_volume`) -- automatically reduces VPUs to 0 (Lower Cost tier) when the volume is detached from all instances, and restores the configured VPUs when re-attached. This saves costs during maintenance windows, instance recreation, or when volumes are pre-provisioned but not yet in use.

## Placeholders to Replace

| Placeholder | Description | Where to Find |
|---|---|---|
| `<compartment-ocid>` | OCID of the compartment where the volume will be created | OCI Console > Identity > Compartments, or `OciCompartment` status outputs |
| `<availability-domain>` | Availability domain for the volume (e.g., `Uocm:US-ASHBURN-AD-1`) | `oci iam availability-domain list` CLI command, or OCI Console > region selector |
| `<backup-policy-ocid>` | OCID of the backup policy to assign (Gold, Silver, Bronze, or custom) | `oci bv volume-backup-policy list` CLI command, or OCI Console > Block Storage > Backup Policies |

## Related Presets

- **02-high-performance-encrypted** -- Use instead for database data files or latency-sensitive workloads requiring higher IOPS, KMS encryption, and cross-region DR
