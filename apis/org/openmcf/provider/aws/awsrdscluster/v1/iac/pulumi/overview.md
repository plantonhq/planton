# AWS RDS Cluster Pulumi Module Overview

This document provides a comprehensive overview of the Pulumi module architecture for deploying AWS RDS Clusters (primarily Aurora MySQL and Aurora PostgreSQL).

## Table of Contents

- [Module Architecture](#module-architecture)
- [Resource Flow](#resource-flow)
- [Implementation Details](#implementation-details)
- [Aurora vs Standard RDS](#aurora-vs-standard-rds)
- [Serverless v2 Support](#serverless-v2-support)
- [Security and Networking](#security-and-networking)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Module Architecture

The Pulumi module is structured into specialized files, each handling a specific aspect of RDS Cluster deployment:

```
iac/pulumi/
├── main.go                     # Entry point for Pulumi program
├── Pulumi.yaml                 # Pulumi project configuration
├── Makefile                    # Deployment automation targets
├── debug.sh                    # Debug helper script
├── README.md                   # User-facing documentation
├── examples.md                 # Usage examples
└── module/
    ├── main.go                 # Orchestration and resource coordination
    ├── locals.go               # Local variables and data transformations
    ├── outputs.go              # Output constant definitions
    ├── rds_cluster.go          # Main RDS Cluster resource
    ├── cluster_param_group.go  # Cluster parameter group management
    ├── security_group.go       # Security group creation and rules
    └── subnet_group.go         # DB subnet group configuration
```

### File Responsibilities

#### main.go (Entry Point)
- Initializes the Pulumi runtime
- Loads stack input from OpenMCF manifest
- Invokes the module's `Resources()` function
- Handles program-level errors

#### module/main.go (Orchestration)
- Coordinates creation of all RDS-related resources
- Manages resource dependencies
- Configures AWS provider (explicit credentials or default)
- Exports stack outputs

#### module/rds_cluster.go
- Creates the core RDS Cluster resource
- Handles Aurora-specific configurations
- Manages master user credentials (managed or manual)
- Configures Serverless v2 scaling
- Applies storage encryption and backup settings

#### module/cluster_param_group.go
- Creates DB Cluster Parameter Group when custom parameters are needed
- Applies cluster-level configuration parameters
- Supports immediate or pending-reboot apply methods

#### module/security_group.go
- Creates a security group for database access
- Configures ingress rules from CIDR blocks
- Associates additional security groups

#### module/subnet_group.go
- Creates DB Subnet Group when subnet IDs are provided
- Ensures multi-AZ deployment capability
- Uses existing subnet group if name is specified

#### module/locals.go
- Initializes local variables from stack input
- Transforms and prepares data for resource creation
- Manages labels and tags

#### module/outputs.go
- Defines output constant names
- Ensures consistent output naming across the module

## Resource Flow

```
User Manifest (YAML)
       ↓
OpenMCF CLI
       ↓
Stack Input (Protobuf)
       ↓
main.go (Pulumi entry)
       ↓
module.Resources()
       ↓
    ┌──┴──┐
    │     │
    ↓     ↓
Provider  Locals
    │     │
    └──┬──┘
       ↓
┌──────┴──────┐
│             │
↓             ↓
Security    Subnet
Group       Group
│             │
└──────┬──────┘
       ↓
   Parameter
     Group
       ↓
   RDS Cluster ← (depends on all above)
       ↓
   Outputs
       ↓
Stack Outputs (Protobuf)
       ↓
User (status.outputs)
```

### Resource Dependencies

1. **Security Group** (created first, optional)
   - Depends on: VPC ID (from spec)
   - Used by: RDS Cluster

2. **Subnet Group** (created first, optional)
   - Depends on: Subnet IDs (from spec)
   - Used by: RDS Cluster

3. **Parameter Group** (created first, optional)
   - Depends on: Engine family
   - Used by: RDS Cluster

4. **RDS Cluster** (created last)
   - Depends on: Security Group, Subnet Group, Parameter Group
   - Exports: Endpoints, ARN, connection details

## Implementation Details

### Provider Configuration

The module supports two provider authentication modes:

**1. Explicit Credentials** (when `ProviderConfig` is provided):
```go
provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
    AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
    SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
    Region:    pulumi.String(locals.AwsRdsCluster.Spec.Region),
    Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
})
```

**2. Default Credentials** (AWS SDK credential chain):
- Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- AWS credentials file (~/.aws/credentials)
- IAM role for EC2 instances
- ECS task role

### Master User Password Management

The module supports two password management strategies:

#### Managed Passwords (Recommended)
When `manage_master_user_password: true`:
- AWS RDS stores the password in AWS Secrets Manager
- Automatic rotation supported
- No password in Pulumi state or manifests
- Optional KMS encryption of the secret

```yaml
spec:
  manage_master_user_password: true
  master_user_secret_kms_key_id:
    value: "arn:aws:kms:us-east-1:123456789012:key/..."
```

#### Manual Passwords
When `manage_master_user_password: false`:
- Password specified in manifest
- Stored in Pulumi state (use secrets encryption)
- Manual rotation required

```yaml
spec:
  manage_master_user_password: false
  username: "admin"
  password: "my-secure-password"
```

**Security Note**: Always use managed passwords in production unless you have a specific reason to manage passwords manually.

### Storage Encryption

#### At-Rest Encryption
```go
if spec.StorageEncrypted {
    args.StorageEncrypted = pulumi.Bool(true)
    if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
        args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
    }
}
```

- Uses AWS-managed key if no KMS key specified
- Customer-managed KMS key for compliance requirements
- Cannot be disabled after cluster creation

#### In-Transit Encryption
- Enforced by Aurora for all connections
- Use SSL/TLS in connection strings
- Certificate verification recommended

## Aurora vs Standard RDS

### Why Aurora?

Aurora isn't simply "managed MySQL/PostgreSQL"—it's a complete architectural reimagining:

**Storage Architecture:**
- **Standard RDS**: EBS volumes attached to each instance
- **Aurora**: Shared, distributed storage layer replicated 6 ways across 3 AZs

**Replication:**
- **Standard RDS**: Asynchronous replication, seconds of lag
- **Aurora**: Storage-level replication, millisecond lag

**Read Replicas:**
- **Standard RDS**: Up to 5 replicas, full data copy for each
- **Aurora**: Up to 15 replicas, share the same storage

**Failover:**
- **Standard RDS**: 1-2 minutes (promote replica, wait for catch-up)
- **Aurora**: 10-30 seconds (just redirect DNS)

**Performance:**
- **Aurora MySQL**: 5x throughput of standard MySQL
- **Aurora PostgreSQL**: 3x throughput of standard PostgreSQL

### When to Use Aurora

✅ **Aurora is ideal for:**
- High read throughput requirements (multiple read replicas)
- Fast failover requirements (sub-30 second RTO)
- Unpredictable growth (storage auto-scales to 128 TiB)
- Multi-region disaster recovery (Global Database)
- Variable workload patterns (Serverless v2)

❌ **Standard RDS might suffice for:**
- Small, predictable workloads where cost is paramount
- Engines Aurora doesn't support (Oracle, SQL Server, MariaDB)
- Legacy applications with specific engine version requirements

## Serverless v2 Support

Aurora Serverless v2 provides automatic, near-instantaneous scaling based on workload demand.

### Configuration

```yaml
spec:
  engine: "aurora-postgresql"
  engine_version: "14.6"
  engine_mode: "provisioned"  # Required for Serverless v2
  serverless_v2_scaling:
    min_capacity: 0.5  # ACUs (1 ACU = 2GB RAM)
    max_capacity: 16.0
```

### Implementation

```go
if spec.ServerlessV2Scaling != nil {
    args.ServerlessV2ScalingConfiguration = &rds.ClusterServerlessV2ScalingConfigurationArgs{
        MinCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MinCapacity),
        MaxCapacity: pulumi.Float64(spec.ServerlessV2Scaling.MaxCapacity),
    }
}
```

### Capacity Units (ACUs)

- **1 ACU** = 2 GiB RAM + corresponding CPU
- **Minimum**: 0.5 ACUs (1 GiB RAM)
- **Maximum**: 128 ACUs (256 GiB RAM)
- **Scaling**: Increments of 0.5 ACUs
- **Speed**: Scales in fractions of a second

### Cost Optimization

**Development/Test Pattern:**
```yaml
serverless_v2_scaling:
  min_capacity: 0.5   # Minimal cost during off-hours
  max_capacity: 4.0   # Cap to prevent runaway costs
```

**Production Pattern:**
```yaml
serverless_v2_scaling:
  min_capacity: 2.0   # Ensure baseline performance
  max_capacity: 32.0  # Handle traffic spikes
```

**Variable Workload Pattern:**
```yaml
serverless_v2_scaling:
  min_capacity: 1.0   # Low baseline
  max_capacity: 64.0  # High burst capacity
```

### Serverless v2 vs Serverless v1

| Feature | Serverless v1 | Serverless v2 |
|---------|--------------|---------------|
| Scaling Speed | Minutes | Seconds |
| Multi-AZ | No | Yes |
| Read Replicas | No | Yes |
| Pause to Zero | Yes | No |
| Production Ready | Limited | Full |
| Engine Support | Limited versions | Current versions |

**Recommendation**: Use Serverless v2 for production workloads. Serverless v1 is mainly for development.

## Security and Networking

### Security Group Management

The module can create and manage a security group:

```yaml
spec:
  vpc_id:
    value: "vpc-12345"
  allowed_cidr_blocks:
    - "10.0.0.0/8"
    - "172.16.0.0/12"
```

**Implementation:**
- Creates a security group in the specified VPC
- Adds ingress rules for each CIDR block
- Allows traffic on the configured database port
- Associates the created security group with the cluster

### Subnet Group Configuration

**Option 1: Provide Subnet IDs**
```yaml
spec:
  subnet_ids:
    - value: "subnet-abc123"
    - value: "subnet-def456"
    - value: "subnet-ghi789"
```
Module creates a DB Subnet Group automatically.

**Option 2: Use Existing Subnet Group**
```yaml
spec:
  db_subnet_group_name:
    value: "my-existing-subnet-group"
```
Module uses the existing group.

**Best Practice**: Always provide at least 3 subnets across different Availability Zones for maximum availability.

### IAM Database Authentication

```yaml
spec:
  iam_database_authentication_enabled: true
```

Benefits:
- No password management required
- Uses IAM policies for access control
- Short-lived authentication tokens (15 minutes)
- Integrates with audit logging

## Best Practices

### High Availability

1. **Multi-AZ Deployment**
   ```yaml
   subnet_ids:
     - value: "subnet-1a"  # AZ 1
     - value: "subnet-1b"  # AZ 2
     - value: "subnet-1c"  # AZ 3
   ```

2. **Automated Backups**
   ```yaml
   backup_retention_period: 7  # Days
   preferred_backup_window: "03:00-04:00"  # UTC
   copy_tags_to_snapshot: true
   ```

3. **Deletion Protection**
   ```yaml
   deletion_protection: true
   skip_final_snapshot: false
   final_snapshot_identifier: "my-cluster-final-snapshot"
   ```

### Performance Optimization

1. **CloudWatch Logs Export**
   ```yaml
   # Aurora MySQL
   enabled_cloudwatch_logs_exports:
     - "error"
     - "general"
     - "slowquery"
   
   # Aurora PostgreSQL
   enabled_cloudwatch_logs_exports:
     - "postgresql"
   ```

2. **Storage Type**
   ```yaml
   storage_type: "aurora-iopt1"  # Higher IOPS for intensive workloads
   ```

3. **Parameter Tuning**
   ```yaml
   parameters:
     - name: "max_connections"
       value: "1000"
       apply_method: "immediate"
     - name: "shared_buffers"
       value: "{DBInstanceClassMemory/32768}"
       apply_method: "pending-reboot"
   ```

### Security Hardening

1. **Encryption**
   ```yaml
   storage_encrypted: true
   kms_key_id:
     value: "arn:aws:kms:us-east-1:123456789012:key/..."
   manage_master_user_password: true
   master_user_secret_kms_key_id:
     value: "arn:aws:kms:us-east-1:123456789012:key/..."
   ```

2. **Network Isolation**
   ```yaml
   # Use private subnets only
   subnet_ids:
     - value: "subnet-private-1a"
     - value: "subnet-private-1b"
   
   # Restrict access
   allowed_cidr_blocks:
     - "10.0.0.0/8"  # Internal VPC only
   ```

3. **Access Control**
   ```yaml
   iam_database_authentication_enabled: true
   ```

### Maintenance Windows

```yaml
preferred_maintenance_window: "sun:03:00-sun:04:00"
preferred_backup_window: "02:00-03:00"
```

**Important**: Backup window must not overlap maintenance window.

### Cost Optimization

1. **Serverless v2 for Variable Workloads**
   ```yaml
   serverless_v2_scaling:
     min_capacity: 0.5  # Low baseline
     max_capacity: 16.0  # Reasonable cap
   ```

2. **Backup Retention**
   ```yaml
   backup_retention_period: 7  # Balance cost vs recovery needs
   ```

3. **Parameter Group Reuse**
   ```yaml
   db_cluster_parameter_group_name: "shared-cluster-params"
   ```

## Troubleshooting

### Common Issues

#### 1. Cluster Creation Timeout

**Symptom**: Pulumi times out waiting for cluster to become available

**Causes**:
- Large restore from snapshot
- Engine version upgrade
- Network connectivity issues

**Solutions**:
- Increase Pulumi timeout settings
- Check VPC and subnet configurations
- Verify security group allows management traffic

#### 2. Password Management Error

**Symptom**: Error about password when using managed passwords

**Cause**: Password specified when `manage_master_user_password: true`

**Solution**:
```yaml
# Remove password field
manage_master_user_password: true
# Do NOT set password field
```

#### 3. Subnet Group Validation Failure

**Symptom**: "Subnets must be in at least 2 different Availability Zones"

**Cause**: All subnets in the same AZ

**Solution**:
```yaml
subnet_ids:
  - value: "subnet-1a"  # us-east-1a
  - value: "subnet-1b"  # us-east-1b
```

#### 4. Serverless v2 Scaling Not Working

**Symptom**: Cluster doesn't scale with Serverless v2 configuration

**Cause**: Engine mode not set to "provisioned" or instances not created

**Solution**:
```yaml
engine_mode: "provisioned"  # Required for Serverless v2
serverless_v2_scaling:
  min_capacity: 0.5
  max_capacity: 16.0
```

Then create DB instances separately with appropriate instance class.

#### 5. Snapshot Restore Failures

**Symptom**: Error restoring from snapshot

**Causes**:
- Snapshot from different engine/version
- Encryption key access issues
- VPC/subnet mismatch

**Solutions**:
- Verify snapshot engine matches cluster engine
- Check KMS key permissions
- Ensure subnet group is compatible

### Debugging

**Enable Debug Output:**
```bash
export PULUMI_DEBUG_COMMANDS=true
pulumi up --logtostderr -v=9 2>&1 | tee pulumi-debug.log
```

**Check AWS RDS Events:**
```bash
aws rds describe-events \
  --source-type db-cluster \
  --source-identifier my-cluster \
  --duration 60
```

**Verify Connectivity:**
```bash
# From within VPC
nc -zv <cluster-endpoint> <port>

# Check security group rules
aws ec2 describe-security-groups \
  --group-ids <sg-id>
```

## State Management

### Pulumi State

The module uses Pulumi's state management to track:
- Cluster configuration
- Security group rules
- Subnet group associations
- Parameter group settings
- Resource dependencies

### State Backend Options

**S3 Backend (Recommended for teams):**
```bash
pulumi login s3://my-pulumi-state-bucket
```

**Pulumi Cloud:**
```bash
pulumi login
```

**Local:**
```bash
pulumi login --local
```

## Migration and Updates

### Engine Version Upgrades

Aurora supports in-place minor version upgrades:

```yaml
# Before
engine_version: "14.6"

# After
engine_version: "14.7"
```

Run `pulumi up` to apply. Aurora will upgrade during the maintenance window.

**Major Version Upgrades**: Require testing and may involve downtime. Create a snapshot and test the upgrade in a non-production cluster first.

### Scaling Configuration Changes

Serverless v2 capacity can be adjusted on the fly:

```yaml
# Increase maximum capacity
serverless_v2_scaling:
  min_capacity: 0.5
  max_capacity: 32.0  # Increased from 16.0
```

Changes take effect immediately.

## References

- [AWS Aurora Documentation](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/)
- [Pulumi AWS RDS Cluster](https://www.pulumi.com/registry/packages/aws/api-docs/rds/cluster/)
- [Aurora Serverless v2](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless-v2.html)
- [OpenMCF Architecture](https://github.com/plantonhq/openmcf/blob/main/architecture/deployment-component.md)

## Conclusion

This Pulumi module provides a production-ready abstraction for deploying AWS Aurora clusters with best practices built in. By handling security groups, subnet groups, parameter groups, and cluster configuration, it reduces the complexity of Aurora deployment while preserving the full power of Aurora's features.

The module supports the full spectrum of Aurora deployments:
- Traditional provisioned instances
- Serverless v2 for automatic scaling
- Global databases for disaster recovery
- Read replicas for scaling reads
- Advanced security with IAM authentication and encryption

For most use cases, the module's defaults provide a secure, highly-available starting point that can be customized as needed.

