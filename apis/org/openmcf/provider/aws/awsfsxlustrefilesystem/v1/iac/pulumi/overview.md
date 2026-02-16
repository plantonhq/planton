# AwsFsxLustreFileSystem Pulumi Module Architecture

## Module Structure

```
module/
├── main.go          # Entry point: provider setup, orchestration, output exports
├── locals.go        # Locals struct: AWS tags, spec references
├── outputs.go       # Output key constants matching AwsFsxLustreFileSystemStackOutputs
└── file_system.go   # FSx Lustre file system resource creation
```

## Data Flow

1. **main.go** receives `AwsFsxLustreFileSystemStackInput` containing the target resource and provider config
2. **locals.go** constructs AWS tags from metadata (organization, environment, resource kind, resource ID)
3. **file_system.go** creates the `fsx.LustreFileSystem` with:
   - Core: deployment type, storage capacity, storage type, Lustre version
   - Networking: subnet ID, security group IDs (ForceNew)
   - Performance: per-unit storage throughput, data compression
   - Encryption: optional customer-managed KMS key
   - S3 integration: import/export paths (SCRATCH only, ForceNew)
   - Logging: CloudWatch log group destination and log level
   - Backups: retention days, daily window, copy tags, skip final backup
   - Maintenance: weekly maintenance start time
   - Metadata: IOPS mode and provisioned IOPS (PERSISTENT_2 only)
4. **main.go** exports outputs: file system ID, ARN, DNS name, mount name, network interface IDs, VPC ID, file system type version, and owner ID

## Key Patterns

- **StringValueOrRef**: Uses `.GetValue()` to extract literal string values. The platform resolves `valueFrom` references before passing to the IaC module.
- **Conditional fields**: Optional spec fields (KMS key, log config, metadata config, backup settings) are only set in the Pulumi args when the spec provides non-zero/non-empty values. This lets AWS apply its defaults.
- **Single subnet**: Unlike multi-AZ resources (EFS, RDS), Lustre is single-AZ. The spec takes a single `subnet_id` instead of `subnet_ids`.
- **ForceNew awareness**: Deployment type, storage type, subnet, security groups, KMS key, and S3 paths are all ForceNew. Changing any of them triggers file system replacement in Pulumi's diff.
- **Zero means default**: Numeric fields left at 0 (e.g., `per_unit_storage_throughput`, `automatic_backup_retention_days`) are not set in the Pulumi args, deferring to AWS defaults.

## Outputs

| Output Key | Source | Description |
|-----------|--------|-------------|
| `file_system_id` | `createdFs.ID()` | File system ID for CSI drivers, ECS, Batch |
| `file_system_arn` | `createdFs.Arn` | ARN for IAM policies |
| `dns_name` | `createdFs.DnsName` | DNS name for mount commands |
| `mount_name` | `createdFs.MountName` | Lustre mount name for mount path |
| `network_interface_ids` | `createdFs.NetworkInterfaceIds` | ENI IDs for debugging |
| `vpc_id` | `createdFs.VpcId` | VPC ID from subnet |
| `file_system_type_version` | `createdFs.FileSystemTypeVersion` | Deployed Lustre version |
| `owner_id` | `createdFs.OwnerId` | AWS account ID |
