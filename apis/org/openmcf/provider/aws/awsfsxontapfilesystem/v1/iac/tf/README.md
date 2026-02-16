# Terraform Module: AwsFsxOntapFileSystem

## Overview

This Terraform module provisions an Amazon FSx for NetApp ONTAP file system. It supports deployment types (single-AZ and multi-AZ), storage types (SSD and HDD), HA pair scale-out, disk IOPS configuration, backup policies, and encryption with customer-managed KMS keys.

## Quick Start

Run the module via the OpenMCF CLI (tofu) using the default local backend.

```bash
openmcf tofu init --manifest ../hack/manifest.yaml
openmcf tofu plan --manifest ../hack/manifest.yaml
openmcf tofu apply --manifest ../hack/manifest.yaml --auto-approve
openmcf tofu destroy --manifest ../hack/manifest.yaml --auto-approve
```

- Credentials are provided via stack input (by the CLI), not in the manifest `spec`.
- Manifest file: `../hack/manifest.yaml`

## File Structure

```
iac/tf/
├── provider.tf    # AWS provider and Terraform version constraints
├── variables.tf   # Input variables (provider_config, metadata, spec)
├── locals.tf      # Tag construction from metadata
├── main.tf        # FSx ONTAP File System resource
└── outputs.tf     # Outputs matching AwsFsxOntapFileSystemStackOutputs
```

## Resources Created

- `aws_fsx_ontap_file_system.this` — the FSx for NetApp ONTAP file system with all configured options

## Key Spec Fields

| Field | Default | Description |
|-------|---------|-------------|
| `deployment_type` | `SINGLE_AZ_2` | SINGLE_AZ_1, SINGLE_AZ_2, MULTI_AZ_1, or MULTI_AZ_2 |
| `storage_capacity_gib` | (required) | 1024–1048576 GiB |
| `storage_type` | `SSD` | SSD or HDD |
| `throughput_capacity_per_ha_pair` | (required) | 128, 256, 384, 512, 768, 1024, 1536, 2048, 3072, 4096, 6144 MB/s |
| `ha_pairs` | `1` | 1–12 for single-AZ; 1 for multi-AZ |
| `subnet_ids` | (required) | 1 subnet (single-AZ) or 2 subnets (multi-AZ) |
| `preferred_subnet_id` | — | Required for multi-AZ |
| `endpoint_ip_address_range` | — | Required for multi-AZ (CIDR for floating IPs) |
| `automatic_backup_retention_days` | `0` | 0–90 days |

## Outputs

| Output | Description |
|--------|-------------|
| `file_system_id` | The ID of the file system |
| `file_system_arn` | The Amazon Resource Name of the file system |
| `dns_name` | DNS name for the file system |
| `management_dns_name` | DNS for ONTAP CLI (SSH) and REST API |
| `management_ip_addresses` | Management endpoint IP addresses |
| `intercluster_dns_name` | DNS for SnapMirror replication |
| `intercluster_ip_addresses` | Intercluster endpoint IP addresses |
| `network_interface_ids` | Network interface IDs |
| `vpc_id` | VPC ID in which the file system was created |
| `owner_id` | AWS account ID of the file system owner |

## Examples

See [../../examples.md](../../examples.md) for sample manifests.
