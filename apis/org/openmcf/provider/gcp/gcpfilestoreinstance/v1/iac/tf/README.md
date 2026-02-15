# GcpFilestoreInstance Terraform Module

Terraform implementation of the GcpFilestoreInstance deployment component.

## Provider Requirements

| Provider | Version | Reason |
|----------|---------|--------|
| hashicorp/google | `~> 6.0` | Required for `performance_config`, `deletion_protection_enabled`, `protocol` |

## Resources Created

- `google_filestore_instance.this` — the Filestore instance with file share, network, and all optional configuration

## Inputs

All inputs are provided via the `spec` variable, which mirrors the protobuf `GcpFilestoreInstanceSpec` structure. See `variables.tf` for the full type definition.

## Outputs

| Output | Description |
|--------|-------------|
| `instance_id` | Fully qualified resource ID |
| `instance_name` | Short name of the instance |
| `ip_addresses` | IP addresses on the VPC network |
| `file_share_name` | File share name for NFS mount path |
| `create_time` | Instance creation timestamp |

## Feature Parity

This module has feature parity with the Pulumi implementation:
- Singular file_shares and networks blocks (not dynamic)
- Dynamic nfs_export_options within file_shares
- Dynamic performance_config with mutually exclusive fixed_iops/iops_per_tb
- Hardcoded `modes = ["MODE_IPV4"]`
- Framework labels applied via locals
