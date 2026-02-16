# Pulumi Module: AwsFsxOpenzfsFileSystem

## Quick Start

```bash
# Build the module
cd module && go build ./...

# Preview changes
make preview

# Deploy
make up

# Destroy
make destroy
```

## Stack Input

The module reads `AwsFsxOpenzfsFileSystemStackInput` from Pulumi config, containing:

- `target` — the `AwsFsxOpenzfsFileSystem` resource manifest (metadata + spec)
- `provider_config` — AWS credentials (access key, secret key, region, session token)

## Outputs

| Key | Source | Description |
|-----|--------|-------------|
| `file_system_id` | `createdFs.ID()` | File system ID |
| `file_system_arn` | `createdFs.Arn` | File system ARN |
| `dns_name` | `createdFs.DnsName` | NFS mount DNS name |
| `endpoint_ip_address` | `createdFs.EndpointIpAddress` | Endpoint IP (floating for Multi-AZ) |
| `root_volume_id` | `createdFs.RootVolumeId` | Root volume ID for child volumes |
| `network_interface_ids` | `createdFs.NetworkInterfaceIds` | ENI IDs |
| `vpc_id` | `createdFs.VpcId` | VPC ID |
| `owner_id` | `createdFs.OwnerId` | AWS account ID |
