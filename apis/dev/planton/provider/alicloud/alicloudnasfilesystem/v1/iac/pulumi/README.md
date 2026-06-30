# AliCloudNasFileSystem Pulumi Module

This Pulumi module provisions an Alibaba Cloud NAS file system with an optional custom access group and a VPC mount target.

## Resources Created

- `alicloud:nas/fileSystem:FileSystem` -- the NAS file system
- `alicloud:nas/accessGroup:AccessGroup` -- (conditional) custom VPC access group, created only when access rules are specified
- `alicloud:nas/accessRule:AccessRule` -- (conditional) one per access rule entry
- `alicloud:nas/mountTarget:MountTarget` -- VPC mount point producing the mount domain name

## Architecture

The module creates the file system first, then conditionally creates an access group with rules if `spec.AccessRules` is non-empty. The mount target references either the custom access group or omits the field to use the default VPC group. For extreme NAS, VPC and VSwitch are set on both the file system and mount target.

## Local Development

```bash
cd apis/dev/planton/provider/alicloud/alicloudnasfilesystem/v1/iac/pulumi
go build ./...
go vet ./...
```

## Stack Outputs

| Name | Description |
| --- | --- |
| `file_system_id` | NAS file system ID |
| `mount_target_domain` | Mount target domain name for NFS/SMB mounting |
