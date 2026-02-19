# AlicloudNasFileSystem Terraform Module

This Terraform module provisions an Alibaba Cloud NAS file system with an optional custom access group and a VPC mount target.

## Resources Created

- `alicloud_nas_file_system` -- the NAS file system
- `alicloud_nas_access_group` -- (conditional) custom VPC access group
- `alicloud_nas_access_rule` -- (conditional) one per access rule
- `alicloud_nas_mount_target` -- VPC mount point

## Files

| File | Description |
|------|-------------|
| `main.tf` | File system and mount target resources |
| `access_group.tf` | Conditional access group and access rules |
| `variables.tf` | Input variable definitions with validations |
| `outputs.tf` | Output values |
| `locals.tf` | Tag computation and derived values |
| `provider.tf` | Alicloud provider configuration |

## Usage

```bash
cd apis/org/openmcf/provider/alicloud/alicloudnasfilesystem/v1/iac/tf
terraform init
terraform validate
```
