# Terraform Module: AwsFsxOpenzfsFileSystem

## Quick Start

```bash
terraform init
terraform plan
terraform apply
```

## Resources Created

- `aws_fsx_openzfs_file_system.this` — the FSx for OpenZFS file system with inline root volume configuration

## Inputs

See `variables.tf` for the complete list of input variables, organized by:

- Provider configuration (access_key, secret_key, region, session_token)
- File system core (deployment_type, storage_capacity_gib, throughput_capacity)
- Networking (subnet_ids, security_group_ids, preferred_subnet_id, route_table_ids)
- Encryption (kms_key_id)
- Disk IOPS (disk_iops_mode, disk_iops)
- Root volume (compression, NFS exports, quotas, record size)
- Backup (retention, schedule, tag propagation)
- Maintenance (weekly window)

## Outputs

See `outputs.tf` — matches `AwsFsxOpenzfsFileSystemStackOutputs` proto definition.
