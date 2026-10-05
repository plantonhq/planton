# Terraform Module: AwsFsxOpenzfsFileSystem

## Quick Start

```bash
terraform init
terraform plan
terraform apply
```

Region and credentials are injected by the runtime as environment variables
(resolved from the IaC input's provider configuration); the module itself
takes only `metadata` and `spec`.

## Resources Created

- `aws_fsx_openzfs_file_system.this` — the FSx for OpenZFS file system with inline root volume configuration

## Inputs

`variables.tf` is generator-owned (regenerated from the proto contract) and
carries two variables:

- `metadata` — resource identity (name, id, org, env, labels)
- `spec` — the typed `AwsFsxOpenzfsFileSystemSpec`: deployment/storage shape
  (incl. INTELLIGENT_TIERING with its read cache), networking, encryption,
  restore, disk IOPS, root volume (compression, NFS exports, quotas), backup
  and deletion behavior, and the maintenance window

## Outputs

See `outputs.tf` — matches `AwsFsxOpenzfsFileSystemOutputs` proto definition.
