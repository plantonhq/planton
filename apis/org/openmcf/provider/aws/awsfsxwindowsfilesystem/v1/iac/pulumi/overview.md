# Pulumi Module Architecture: AWS FSx for Windows File Server

## Overview

This Pulumi module provisions Amazon FSx for Windows File Server file systems through a declarative, protobuf-defined specification. The architecture follows OpenMCF's standard pattern: input transformation → resource provisioning → output extraction.

The module exposes FSx for Windows' full feature set—deployment types, Active Directory integration (both AWS Managed and self-managed), audit logging, disk IOPS configuration, backup policies, and DNS aliases—while maintaining simplicity through careful abstraction of AWS API complexity.

## Module Structure

```
iac/pulumi/
├── main.go              # Pulumi entrypoint
├── Pulumi.yaml          # Pulumi project metadata
├── Makefile             # Build/test helpers
├── debug.sh             # Delve debugging wrapper
└── module/
    ├── main.go          # Orchestration logic (provider setup, output export)
    ├── locals.go        # Input transformation and tag construction
    ├── file_system.go   # FSx Windows File System resource implementation
    └── outputs.go       # Output key constants
```

### File Responsibilities

#### `main.go` (entrypoint)
- Unmarshals `AwsFsxWindowsFileSystemStackInput` from Pulumi stack configuration
- Delegates to `module.Resources()` for actual provisioning
- Minimal logic—purely a thin entrypoint wrapper

#### `module/main.go` (orchestrator)
**Key Function:** `Resources(ctx, stackInput)`

Responsibilities:
1. **Locals Initialization**: Transform stack input into typed locals struct
2. **Provider Configuration**: Handle two provider scenarios:
   - **Default**: Create provider with ambient AWS credentials (IAM role, environment variables)
   - **Explicit**: Create provider with credentials from `stackInput.ProviderConfig` (access key, secret key, session token)
3. **Resource Creation**: Invoke `fileSystem()` with initialized locals and provider
4. **Output Export**: Export eight FSx-specific outputs using Pulumi's export mechanism

**Design Decision**: Provider handling is bifurcated to support both:
- **CI/CD environments**: IRSA (IAM Roles for Service Accounts) or instance profiles provide ambient credentials
- **Local/manual deployments**: Explicit credentials passed via stack input

#### `module/locals.go` (input transformer)
**Key Function:** `initializeLocals(ctx, stackInput)`

Transforms the protobuf `AwsFsxWindowsFileSystemStackInput` into a strongly-typed `Locals` struct and constructs AWS resource tags.

**Tag Construction**: Every FSx file system receives five mandatory tags:
- `resource=true`: Marks this as an OpenMCF managed resource
- `organization=<org>`: Organization ID from metadata
- `environment=<env>`: Environment (dev/staging/prod) from metadata
- `resource-kind=AwsFsxWindowsFileSystem`: CloudResourceKind enum string
- `resource-id=<id>`: Unique resource identifier from metadata

**Why Tags Matter**: These tags enable:
- Cost allocation reporting by org/env/resource-kind
- Policy enforcement (e.g., "only production resources can use HDD storage")
- Resource discovery and inventory management
- Drift detection (manually created resources lack these tags)

#### `module/file_system.go` (resource implementation)
**Key Function:** `fileSystem(ctx, locals, provider)`

This is the core implementation file containing all FSx provisioning logic. It translates the protobuf `AwsFsxWindowsFileSystemSpec` into Pulumi's `fsx.WindowsFileSystemArgs`.

**Mapping Strategy**: The implementation follows a deliberate pattern of **explicit field-by-field mapping** with conditional application of optional fields. Each optional field is only set when the user provides a non-zero value, preventing AWS API errors from empty/zero values.

#### `module/outputs.go` (constants)
Defines string constants for output keys. This prevents typos in export calls and provides a single source of truth for output names.

**Exported Outputs**:
- `file_system_id`: The FSx file system's resource ID
- `file_system_arn`: Full ARN for IAM policy construction
- `dns_name`: DNS name for mounting via SMB (e.g., `\\fs-0123456789abcdef0.example.com\share`)
- `preferred_file_server_ip`: IP address of the preferred file server (Multi-AZ deployments)
- `remote_administration_endpoint`: Endpoint for remote PowerShell management
- `network_interface_ids`: ENI IDs created for the file system
- `vpc_id`: VPC where the file system resides
- `owner_id`: AWS account ID of the file system owner

## Data Flow Diagram

```
┌──────────────────────────────────────────────────────┐
│ AwsFsxWindowsFileSystemStackInput (protobuf)         │
│  ├─ target: AwsFsxWindowsFileSystem                  │
│  │   ├─ metadata (org, env, name, id)                │
│  │   └─ spec: AwsFsxWindowsFileSystemSpec            │
│  └─ provider_config (optional)                       │
└─────────────────────┬────────────────────────────────┘
                      │
                      ▼
            ┌──────────────────┐
            │ initializeLocals │
            └────────┬─────────┘
                     │ Creates:
                     │ - Locals.AwsFsxWindowsFileSystem
                     │ - Locals.AwsTags (map[string]string)
                     ▼
           ┌──────────────────────┐
           │ AWS Provider Setup   │
           │  ├─ Ambient creds OR │
           │  └─ Explicit creds   │
           └──────────┬───────────┘
                      │
                      ▼
             ┌────────────────┐
             │  fileSystem()  │
             └────────┬───────┘
                      │
                      ├─ Required fields:
                      │    subnet_ids, throughput_capacity, storage_capacity
                      │
                      ├─ Optional scalar fields (conditional):
                      │    deployment_type, storage_type, kms_key_id,
                      │    preferred_subnet_id, active_directory_id,
                      │    automatic_backup_retention_days,
                      │    daily_automatic_backup_start_time,
                      │    weekly_maintenance_start_time,
                      │    copy_tags_to_backups, skip_final_backup
                      │
                      ├─ Optional list fields (conditional):
                      │    security_group_ids, aliases
                      │
                      ├─ Optional nested objects (conditional):
                      │    self_managed_active_directory
                      │    audit_log_configuration
                      │    disk_iops_configuration
                      │
                      └─ Tags from locals.AwsTags
                      │
                      ▼
          ┌────────────────────────────┐
          │ fsx.NewWindowsFileSystem() │
          │  (Pulumi AWS Provider)     │
          └────────────┬───────────────┘
                       │
                       ▼ Creates AWS FSx Resource
             ┌──────────────────────────┐
             │  AWS FSx Service          │
             │   ├─ File System          │
             │   ├─ Network Interfaces   │
             │   ├─ AD Integration       │
             │   └─ Backups              │
             └──────────┬───────────────┘
                        │
                        ▼
              ┌──────────────────────────────────┐
              │ Output Exports                   │
              │  ├─ file_system_id               │
              │  ├─ file_system_arn              │
              │  ├─ dns_name                     │
              │  ├─ preferred_file_server_ip     │
              │  ├─ remote_administration_endpoint│
              │  ├─ network_interface_ids        │
              │  ├─ vpc_id                       │
              │  └─ owner_id                     │
              └──────────────────────────────────┘
```

## Resource Relationships

```
AwsFsxWindowsFileSystemSpec
  │
  ├─ Core Configuration
  │   ├─ subnet_ids ────────────────────┐
  │   ├─ throughput_capacity ───────────┤
  │   ├─ storage_capacity_gib ─────────┤
  │   ├─ deployment_type ──────────────┤   FSx Windows File System
  │   └─ storage_type ─────────────────┤     ├─ ENIs (one per subnet)
  │                                     │     ├─ DNS Name (SMB endpoint)
  ├─ Networking                         │     ├─ Preferred File Server IP
  │   ├─ security_group_ids ───────────┤     ├─ Remote Admin Endpoint
  │   └─ preferred_subnet_id ─────────┤     │
  │                                     │     ├─ Storage (SSD or HDD)
  ├─ Encryption                         │     │   └─ Optional KMS CMK
  │   └─ kms_key_id ──────────────────┤     │
  │                                     │     ├─ Active Directory
  ├─ Active Directory (one of)          │     │   ├─ AWS Managed AD
  │   ├─ active_directory_id ──────────┤     │   └─ Self-Managed AD
  │   └─ self_managed_active_directory ┤     │
  │       ├─ domain_name               │     ├─ Audit Logging
  │       ├─ dns_ips                   │     │   ├─ File Access Level
  │       ├─ username                  │     │   ├─ File Share Access Level
  │       ├─ password                  │     │   └─ Destination (CloudWatch/Firehose)
  │       ├─ file_system_admins_group  │     │
  │       └─ ou_distinguished_name     │     ├─ Disk IOPS
  │                                     │     │   ├─ Mode (AUTOMATIC/USER_PROVISIONED)
  ├─ Audit Logging                      │     │   └─ IOPS value
  │   └─ audit_log_configuration ──────┤     │
  │                                     │     ├─ Backups
  ├─ Disk IOPS                          │     │   ├─ Retention Days
  │   └─ disk_iops_configuration ──────┤     │   ├─ Daily Start Time
  │                                     │     │   ├─ Copy Tags to Backups
  ├─ Backup & Maintenance               │     │   └─ Skip Final Backup
  │   ├─ automatic_backup_retention ───┤     │
  │   ├─ daily_backup_start_time ──────┤     ├─ DNS Aliases
  │   ├─ copy_tags_to_backups ─────────┤     │
  │   ├─ skip_final_backup ────────────┤     └─ Maintenance Window
  │   └─ weekly_maintenance_start_time ┘
  │
  └─ DNS Aliases
      └─ aliases ──────────────────────┘
```

## Key Design Decisions

### 1. Active Directory: Two Mutually Exclusive Paths

**Decision**: Support both AWS Managed Microsoft AD and self-managed AD, but they are mutually exclusive.

**Implementation**:
```go
// Path 1: AWS Managed Microsoft AD
if spec.ActiveDirectoryId != nil && spec.ActiveDirectoryId.GetValue() != "" {
    args.ActiveDirectoryId = pulumi.StringPtr(spec.ActiveDirectoryId.GetValue())
}

// Path 2: Self-managed AD
if spec.SelfManagedActiveDirectory != nil {
    // Build full SMAD configuration
}
```

**Rationale**:
- **AWS Managed AD**: Simpler—just reference an existing Directory Service directory by ID. AWS handles domain join automatically.
- **Self-Managed AD**: More complex—requires domain name, DNS IPs, credentials, and optional OU placement. Needed when the organization runs its own AD on-premises or in EC2.
- **Mutual Exclusivity**: AWS API rejects requests with both fields set. The protobuf validation enforces this at the manifest level.

### 2. Secrets Manager Limitation in Pulumi SDK v7.3.0

**Decision**: The `domain_join_service_account_secret_arn` field from the proto spec is not wired to the Pulumi resource.

**Rationale**:
- The Pulumi AWS SDK v7.3.0 `WindowsFileSystemSelfManagedActiveDirectoryArgs` requires `Username` and `Password` as `StringInput` (not optional).
- The AWS API supports a `DomainJoinServiceAccountSecretArn` field that provides credentials via Secrets Manager, but this is not exposed in the current SDK version.
- The proto spec retains the field for **forward compatibility**—when the SDK adds support, the wiring is a one-line change.
- **Workaround**: Users needing Secrets Manager credentials should use the Terraform module (`iac/tf/`), which supports `domain_join_service_account_secret` natively.

### 3. Conditional Field Application

**Decision**: Every optional field is guarded by a non-zero/non-empty check before being set on the Pulumi args.

**Implementation Pattern**:
```go
if spec.GetDeploymentType() != "" {
    args.DeploymentType = pulumi.StringPtr(spec.GetDeploymentType())
}
```

**Rationale**:
- **AWS API Sensitivity**: Many FSx fields reject empty strings or zero values. For example, setting `DeploymentType` to `""` causes an API error.
- **Protobuf Zero Values**: Protobuf's zero values (empty string, 0, false) are indistinguishable from "not set." The guards ensure that unset fields remain nil in the Pulumi args, letting AWS apply its own defaults.
- **Explicit Over Implicit**: Rather than setting defaults in the module, the defaults are documented in the proto spec and applied by the AWS API.

### 4. StringValue Wrapper Types for IDs

**Decision**: Fields like `subnet_ids`, `security_group_ids`, `kms_key_id`, `active_directory_id`, and `preferred_subnet_id` use protobuf `StringValue` wrapper types.

**Implementation**:
```go
for _, s := range spec.SubnetIds {
    subnetIds = append(subnetIds, pulumi.String(s.GetValue()))
}
```

**Rationale**:
- **Nullable Semantics**: `StringValue` differentiates between "not set" (nil) and "set to empty string." This is critical for fields like `kms_key_id` where nil means "use AWS-managed key" and empty string would cause an error.
- **OpenMCF Convention**: All AWS resource ID references use `StringValue` across the codebase for consistency.

### 5. Tag Injection at Resource Level

**Decision**: Tags are constructed in `locals.go` and attached to the FSx resource directly.

**Rationale**:
- **Consistency**: Every OpenMCF resource has identical tagging structure
- **Immutability**: Tags reflect metadata from the original manifest—they are derived, not configured
- **Cost Allocation**: AWS Cost Explorer can aggregate costs by `organization`, `environment`, and `resource-kind` tags

## FSx Windows-Specific Implementation Details

### Deployment Types

The module supports all FSx for Windows deployment types via the `deployment_type` field:

| Deployment Type | Subnets Required | Use Case |
|----------------|-----------------|----------|
| `SINGLE_AZ_1` | 1 | Development, non-critical workloads |
| `SINGLE_AZ_2` | 1 | Single-AZ with SSD IOPS optimization |
| `MULTI_AZ_1` | 2 | Production, automatic failover |

**Implementation**: The deployment type is passed directly as a string. AWS validates the value and the subnet count alignment.

### Storage Configuration

**Storage Capacity**: Specified in GiB. Minimum depends on deployment type and storage type:
- SSD: 32 GiB minimum
- HDD: 2,000 GiB minimum (only available with `SINGLE_AZ_2` and `MULTI_AZ_1`)

**Throughput Capacity**: Specified in MB/s. Valid values: 8, 16, 32, 64, 128, 256, 512, 1024, 2048.

**Disk IOPS Configuration**: Optional. Two modes:
- `AUTOMATIC` (default): AWS scales IOPS with storage capacity (3 IOPS per GiB for SSD)
- `USER_PROVISIONED`: Explicit IOPS value, up to 160,000

### Audit Log Configuration

The module supports Windows file access auditing with configurable log levels:

**Log Levels** (for both file access and file share access):
- `DISABLED`: No auditing
- `SUCCESS_ONLY`: Audit successful access attempts
- `FAILURE_ONLY`: Audit failed access attempts
- `SUCCESS_AND_FAILURE`: Audit all access attempts

**Destination**: CloudWatch Logs log group ARN or Kinesis Data Firehose delivery stream ARN. If omitted, audit events are emitted but not delivered to a destination.

### Backup Configuration

| Field | Default | Description |
|-------|---------|-------------|
| `automatic_backup_retention_days` | 7 | Days to retain automatic backups (0 disables) |
| `daily_automatic_backup_start_time` | AWS default | Time in `HH:MM` UTC format |
| `copy_tags_to_backups` | false | Propagate resource tags to backup copies |
| `skip_final_backup` | true | Skip creating a final backup on deletion |

**Why `skip_final_backup` defaults to true**: For development and testing, retaining a final backup on every destroy cycle creates unnecessary cost. Production deployments should explicitly set this to `false`.

### DNS Aliases

The `aliases` field allows associating custom DNS names with the file system. These are CNAME records that point to the file system's DNS name. Useful for:
- Migration from existing file servers (keep the same UNC path)
- Shorter, human-friendly mount points

## Error Handling Philosophy

The module follows a **fail-fast** approach:

1. **Validation at Protobuf Level**: The `spec.proto` validation rules catch configuration errors before Pulumi runs
2. **AWS API Errors Propagate**: If AWS rejects a configuration, the error propagates immediately (no silent fallbacks)
3. **No Default Mutations**: The module does not silently change user configuration

## Testing and Debugging

### Unit Testing (Not Present)

The module has no Go unit tests. This is intentional:
- **Integration Testing**: FSx provisioning requires actual AWS API calls and takes 20-30 minutes
- **Validation Testing**: The protobuf validation tests (in the parent directory) verify configuration correctness

### Debugging with Delve

The `debug.sh` script enables step-through debugging:

1. Uncomment the binary option in `Pulumi.yaml`:
   ```yaml
   runtime:
     options:
       binary: ./debug.sh
   ```
2. Run Pulumi CLI commands normally
3. The debug script launches Delve, allowing breakpoints in any module file

**Use Case**: Debugging Active Directory configuration issues or investigating AWS API errors with SMAD credentials.

## Common Pitfalls and Gotchas

### Pitfall 1: Mixing AD Types
**Symptom**: AWS API error when both `active_directory_id` and `self_managed_active_directory` are specified.

**Cause**: These are mutually exclusive. You must choose one AD integration method.

**Solution**: Set only one. The protobuf validation enforces this, but bypassing validation causes an AWS API error.

### Pitfall 2: HDD with Wrong Deployment Type
**Symptom**: AWS API error about unsupported storage type.

**Cause**: HDD storage is only available with `SINGLE_AZ_2` and `MULTI_AZ_1` deployment types.

**Solution**: Use `SSD` storage type or switch to a compatible deployment type.

### Pitfall 3: MULTI_AZ_1 Without Preferred Subnet
**Symptom**: Unpredictable file server placement in Multi-AZ deployments.

**Cause**: Not specifying `preferred_subnet_id` for `MULTI_AZ_1` deployments.

**Solution**: Always set `preferred_subnet_id` for Multi-AZ deployments to control which AZ hosts the active file server.

### Pitfall 4: Self-Managed AD Without Secrets Manager (Pulumi)
**Symptom**: Credentials must be provided as plaintext in the proto spec.

**Cause**: Pulumi SDK v7.3.0 does not expose `DomainJoinServiceAccountSecretArn`.

**Solution**: Use the Terraform module (`iac/tf/`) which supports `domain_join_service_account_secret` natively, or provide `username` and `password` directly.

### Pitfall 5: Insufficient Throughput for Workload
**Symptom**: Slow file system performance despite adequate storage capacity.

**Cause**: Throughput capacity is independent of storage capacity. A large file system with low throughput will bottleneck on I/O.

**Solution**: Size throughput based on workload requirements, not storage size. AWS recommends monitoring `DataReadBytes` and `DataWriteBytes` CloudWatch metrics.

## Performance Considerations

### Pulumi Resource Creation Time

FSx for Windows file system creation is a long-running operation:
- **Single-AZ**: 20-30 minutes typically
- **Multi-AZ**: 30-45 minutes typically
- **With AD join**: Additional 5-10 minutes for domain join operations

### Pulumi Diff Calculation

Changes to most fields trigger file system updates:
- **Immutable Fields**: Changing `deployment_type`, `subnet_ids`, or `active_directory_id` requires file system replacement (destroys and recreates)
- **Mutable Fields**: Changing `storage_capacity` (increase only), `throughput_capacity`, `audit_log_configuration`, and `aliases` triggers in-place updates

**Recommendation**: Use `pulumi preview` before `pulumi up` to understand whether changes will cause file system replacement.

## Future Enhancements

### Secrets Manager Integration (Pulumi)
**What**: Wire `domain_join_service_account_secret_arn` when the Pulumi SDK exposes it.

**Path Forward**: Monitor `pulumi-aws` releases for `WindowsFileSystemSelfManagedActiveDirectoryArgs` additions. The proto field already exists; only the Go wiring needs updating.

### Data Deduplication
**What**: Enable Windows data deduplication on the file system.

**Why Not Implemented**: Data deduplication is configured at the Windows level (inside the file system), not at the AWS API level. It requires post-provisioning PowerShell commands.

### File System Associations
**What**: Associate the file system with additional VPCs or AD domains.

**Path Forward**: Would require additional Pulumi resources (`fsx.DataRepositoryAssociation` or `fsx.FileSystemAssociation`).
