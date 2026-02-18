# AWS RDS Instance Pulumi Module Overview

This document provides a comprehensive overview of the Pulumi module architecture for deploying AWS RDS DB Instances (single instances, not Aurora clusters).

## Table of Contents

- [Module Architecture](#module-architecture)
- [Resource Flow](#resource-flow)
- [Implementation Details](#implementation-details)
- [RDS Instance vs Aurora](#rds-instance-vs-aurora)
- [Supported Engines](#supported-engines)
- [Security and Networking](#security-and-networking)
- [High Availability](#high-availability)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Module Architecture

The Pulumi module is organized into focused files, each handling a specific aspect of RDS instance deployment:

```
iac/pulumi/
├── main.go              # Entry point for Pulumi program
├── Pulumi.yaml          # Pulumi project configuration
├── Makefile             # Build and deployment automation
├── debug.sh             # Debug helper script
├── README.md            # User-facing documentation
├── examples.md          # Usage examples
└── module/
    ├── main.go          # Orchestration and resource coordination
    ├── locals.go        # Local variables and data transformations
    ├── outputs.go       # Output constant definitions
    ├── instance.go      # Main RDS Instance resource
    └── subnet_group.go  # DB subnet group configuration
```

### File Responsibilities

#### main.go (Entry Point)
- Initializes the Pulumi runtime
- Loads stack input from OpenMCF manifest
- Invokes the module's `Resources()` function
- Handles program-level errors

#### module/main.go (Orchestration)
- Coordinates creation of RDS-related resources
- Manages resource dependencies
- Configures AWS provider (explicit credentials or default)
- Exports stack outputs

#### module/instance.go
- Creates the core RDS Instance resource
- Handles all engine types (PostgreSQL, MySQL, MariaDB, Oracle, SQL Server)
- Manages instance configuration (class, storage, encryption)
- Applies networking settings (VPC, security groups, public access)
- Configures high availability (Multi-AZ)
- Sets up authentication (username, password)

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
   Subnet Group
   (if subnet_ids provided)
       ↓
   RDS Instance
   (depends on subnet group)
       ↓
   Outputs
       ↓
Stack Outputs (Protobuf)
       ↓
User (status.outputs)
```

### Resource Dependencies

1. **Subnet Group** (created first, optional)
   - Depends on: Subnet IDs (from spec)
   - Used by: RDS Instance

2. **RDS Instance** (created last)
   - Depends on: Subnet Group (if created)
   - Exports: Endpoint, ARN, connection details

## Implementation Details

### Provider Configuration

The module supports two provider authentication modes:

**1. Explicit Credentials** (when `ProviderConfig` is provided):
```go
provider, err = aws.NewProvider(ctx, "classic-provider", &aws.ProviderArgs{
    AccessKey: pulumi.String(awsProviderConfig.AccessKeyId),
    SecretKey: pulumi.String(awsProviderConfig.SecretAccessKey),
    Region:    pulumi.String(locals.AwsRdsInstance.Spec.Region),
    Token:     pulumi.StringPtr(awsProviderConfig.SessionToken),
})
```

**2. Default Credentials** (AWS SDK credential chain):
- Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
- AWS credentials file (~/.aws/credentials)
- IAM role for EC2 instances
- ECS task role

### Database Engine Support

The module supports all AWS RDS engine types:

#### PostgreSQL
```yaml
spec:
  engine: "postgres"
  engine_version: "14.10"
  instance_class: "db.t3.medium"
  port: 5432
```

#### MySQL
```yaml
spec:
  engine: "mysql"
  engine_version: "8.0.35"
  instance_class: "db.t3.medium"
  port: 3306
```

#### MariaDB
```yaml
spec:
  engine: "mariadb"
  engine_version: "10.11.5"
  instance_class: "db.t3.medium"
  port: 3306
```

#### Oracle
```yaml
spec:
  engine: "oracle-se2"
  engine_version: "19.0.0.0.ru-2023-10.rur-2023-10.r1"
  instance_class: "db.t3.medium"
  port: 1521
```

#### SQL Server
```yaml
spec:
  engine: "sqlserver-ex"
  engine_version: "15.00.4316.3.v1"
  instance_class: "db.t3.medium"
  port: 1433
```

### Storage Configuration

#### Basic Storage
```go
AllocatedStorage: pulumi.Int(int(spec.AllocatedStorageGb))
```

Minimum storage varies by engine:
- **PostgreSQL**: 20 GiB minimum
- **MySQL**: 20 GiB minimum
- **MariaDB**: 20 GiB minimum
- **Oracle**: 20 GiB minimum
- **SQL Server**: 20 GiB minimum

#### Storage Encryption
```go
StorageEncrypted: pulumi.Bool(spec.StorageEncrypted)
if spec.KmsKeyId != nil && spec.KmsKeyId.GetValue() != "" {
    args.KmsKeyId = pulumi.String(spec.KmsKeyId.GetValue())
}
```

- Uses AWS-managed key if no KMS key specified
- Customer-managed KMS key for compliance requirements
- Cannot be disabled after instance creation

### Credential Management

```go
if spec.Username != "" {
    args.Username = pulumi.String(spec.Username)
}
if spec.Password != "" {
    args.Password = pulumi.String(spec.Password)
}
```

**Security Note**: Passwords are stored in Pulumi state. Consider:
- Using Pulumi secrets encryption
- Rotating passwords after initial creation
- Using IAM database authentication where possible
- Storing passwords in AWS Secrets Manager separately

### Network Configuration

#### Subnet Group
```go
// Created when subnet_ids provided
if len(spec.SubnetIds) >= 2 {
    subnetGroup = createSubnetGroup(...)
}

// Or use existing group
if spec.DbSubnetGroupName != nil {
    args.DbSubnetGroupName = pulumi.String(spec.DbSubnetGroupName.GetValue())
}
```

#### Security Groups
```go
var sgIds pulumi.StringArray
for _, sg := range spec.SecurityGroupIds {
    if sg.GetValue() != "" {
        sgIds = append(sgIds, pulumi.String(sg.GetValue()))
    }
}
args.VpcSecurityGroupIds = sgIds
```

#### Public Access
```go
PubliclyAccessible: pulumi.Bool(spec.PubliclyAccessible)
```

**Security Warning**: Set `publicly_accessible: false` for production databases!

## RDS Instance vs Aurora

### Key Differences

| Feature | RDS Instance | Aurora |
|---------|--------------|--------|
| Storage | EBS volumes | Distributed storage layer |
| Replication | Asynchronous | Storage-level (synchronous) |
| Failover Time | 1-2 minutes | 10-30 seconds |
| Read Replicas | Up to 5 | Up to 15 |
| Scaling | Manual resize | Auto-scaling (Serverless v2) |
| Storage Auto-Scale | Yes (to 64 TiB) | Yes (to 128 TiB) |
| Cost | Lower for small workloads | Higher base cost |

### When to Use RDS Instance

✅ **RDS Instance is ideal for:**
- Small to medium workloads (< 100 connections)
- Cost-sensitive applications
- Engines Aurora doesn't support (Oracle, SQL Server, MariaDB)
- Simple deployment without advanced HA requirements
- Development and testing environments

❌ **Use Aurora instead for:**
- High read throughput requirements
- Sub-30 second failover requirements
- Need for 10+ read replicas
- Variable workload patterns (Serverless v2)
- Multi-region disaster recovery

## Security and Networking

### Multi-AZ Deployment

```yaml
spec:
  multi_az: true
  subnet_ids:
    - value: "subnet-1a"  # Availability Zone 1
    - value: "subnet-1b"  # Availability Zone 2
```

**How it works:**
- AWS automatically provisions a standby replica in a different AZ
- Synchronous replication to standby
- Automatic failover in case of AZ failure
- Same endpoint (DNS automatically redirects)
- Increased cost (~2x) but significantly improved availability

### Private Subnet Deployment

```yaml
spec:
  publicly_accessible: false
  subnet_ids:
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[0].id"
    - valueFrom:
        kind: AwsVpc
        name: "my-vpc"
        fieldPath: "status.outputs.privateSubnets.[1].id"
```

**Best Practice**: Always use private subnets for production databases.

### Security Group Configuration

```yaml
spec:
  security_group_ids:
    - valueFrom:
        kind: AwsSecurityGroup
        name: "app-sg"
        fieldPath: "status.outputs.securityGroupId"
```

Security group should allow:
- Ingress on database port from application security groups
- No egress restrictions (unless required)

### Encryption at Rest

```yaml
spec:
  storage_encrypted: true
  kms_key_id:
    valueFrom:
      kind: AwsKmsKey
      name: "rds-key"
      fieldPath: "status.outputs.keyArn"
```

### Encryption in Transit

All RDS engines support SSL/TLS connections:
- **PostgreSQL**: `sslmode=require` in connection string
- **MySQL**: `--ssl-mode=REQUIRED` connection option
- **SQL Server**: `Encrypt=True` in connection string
- **Oracle**: Requires Oracle Native Network Encryption

## High Availability

### Multi-AZ Architecture

```
┌─────────────────────────────────────────────────┐
│                   AWS Region                     │
│                                                  │
│  ┌──────────────────┐      ┌──────────────────┐ │
│  │  Availability     │      │  Availability     │ │
│  │  Zone A          │      │  Zone B          │ │
│  │                  │      │                  │ │
│  │  ┌────────────┐  │      │  ┌────────────┐  │ │
│  │  │  Primary   │  │      │  │  Standby   │  │ │
│  │  │  RDS       │◄─┼──────┼─►│  RDS       │  │ │
│  │  │  Instance  │  │ Sync │  │  Instance  │  │ │
│  │  └────────────┘  │ Repl │  └────────────┘  │ │
│  │                  │      │                  │ │
│  └──────────────────┘      └──────────────────┘ │
│                                                  │
└─────────────────────────────────────────────────┘
           ▲
           │
    Single DNS Endpoint
   (Automatic Failover)
```

### Backup Configuration

RDS automatically creates backups:
- **Automated Backups**: Daily snapshots retained 1-35 days
- **Backup Window**: Configurable time window (default: random)
- **Transaction Logs**: Continuous backup for point-in-time recovery
- **Manual Snapshots**: User-triggered, retained indefinitely

### Disaster Recovery

**Recovery Time Objective (RTO)**:
- Single-AZ: ~5-15 minutes (restore from backup)
- Multi-AZ: ~1-2 minutes (automatic failover)

**Recovery Point Objective (RPO)**:
- Automated backups: Last snapshot + transaction logs
- Multi-AZ: Zero data loss (synchronous replication)

## Best Practices

### Instance Sizing

**Development:**
```yaml
instance_class: "db.t3.micro"   # 1 vCPU, 1 GiB RAM
allocated_storage_gb: 20
multi_az: false
```

**Production (Small):**
```yaml
instance_class: "db.t3.medium"  # 2 vCPU, 4 GiB RAM
allocated_storage_gb: 100
multi_az: true
storage_encrypted: true
```

**Production (Large):**
```yaml
instance_class: "db.m6i.2xlarge"  # 8 vCPU, 32 GiB RAM
allocated_storage_gb: 500
multi_az: true
storage_encrypted: true
```

### Storage Planning

**Storage Auto-Scaling:**
RDS can automatically increase storage when free space is low:
- Set initial `allocated_storage_gb` to expected baseline
- RDS scales up to 64 TiB (standard RDS) or 128 TiB (Aurora)
- Scaling is one-way (cannot decrease storage)

**IOPS Considerations:**
- **General Purpose (gp3)**: Default, 3000 IOPS baseline
- **Provisioned IOPS (io1)**: For high-performance workloads
- **Magnetic**: Legacy, not recommended

### Parameter Groups

```yaml
spec:
  parameter_group_name: "custom-postgres-14"
```

Use custom parameter groups for:
- Connection limits (`max_connections`)
- Memory settings (`shared_buffers`)
- Query optimization settings
- Logging configuration

### Option Groups

```yaml
spec:
  option_group_name: "custom-oracle-options"
```

Required for:
- Oracle Advanced Security (TDE, encryption)
- SQL Server features (SSRS, SSIS)
- Enhanced monitoring

### Monitoring

**CloudWatch Metrics:**
- CPU Utilization
- Database Connections
- Free Storage Space
- Read/Write IOPS
- Network Throughput

**Enhanced Monitoring:**
- OS-level metrics
- Process-level metrics
- More granular data (down to 1 second)

**Performance Insights:**
- SQL-level performance analysis
- Wait event analysis
- Top queries by load

## Troubleshooting

### Common Issues

#### 1. Instance Creation Timeout

**Symptom**: Pulumi times out waiting for instance to become available

**Causes**:
- Large initial storage allocation
- Complex parameter group
- Network connectivity issues

**Solutions**:
- Increase Pulumi timeout
- Start with smaller storage, let auto-scaling handle growth
- Verify VPC and subnet configurations

#### 2. Connection Refused

**Symptom**: Cannot connect to database endpoint

**Causes**:
- Security group doesn't allow inbound traffic
- Database in private subnet without proper routing
- Wrong port number
- Instance not yet available

**Solutions**:
```bash
# Check security group rules
aws ec2 describe-security-groups --group-ids <sg-id>

# Test connectivity (from within VPC)
nc -zv <endpoint> <port>

# Check instance status
aws rds describe-db-instances --db-instance-identifier <id>
```

#### 3. Storage Full

**Symptom**: Database writes fail, logs show storage errors

**Solutions**:
- Enable storage auto-scaling
- Manually increase storage
- Clean up unnecessary data
- Archive old logs

**Prevention**:
Set up CloudWatch alarms:
```yaml
# Create alarm for low free storage
Threshold: < 10% free space
Action: Notify operations team
```

#### 4. High CPU Usage

**Symptom**: Queries slow, CPU metrics at 100%

**Solutions**:
- Identify slow queries using Performance Insights
- Optimize queries (indexes, query structure)
- Increase instance class
- Implement read replicas for read-heavy workloads

#### 5. Credential Issues

**Symptom**: Authentication failures

**Causes**:
- Incorrect username/password
- Password changed outside Pulumi
- IAM authentication misconfigured

**Solutions**:
```bash
# Reset password via AWS CLI
aws rds modify-db-instance \
  --db-instance-identifier <id> \
  --master-user-password <new-password> \
  --apply-immediately

# Then update Pulumi state
pulumi refresh
```

### Debugging

**Enable Debug Output:**
```bash
export PULUMI_DEBUG_COMMANDS=true
pulumi up --logtostderr -v=9 2>&1 | tee pulumi-debug.log
```

**Check RDS Events:**
```bash
aws rds describe-events \
  --source-type db-instance \
  --source-identifier <instance-id> \
  --duration 60
```

**Verify Network Configuration:**
```bash
# Check route tables
aws ec2 describe-route-tables --filters "Name=vpc-id,Values=<vpc-id>"

# Check subnet configuration
aws ec2 describe-subnets --subnet-ids <subnet-id>
```

## State Management

### Pulumi State

The module uses Pulumi's state management to track:
- Instance configuration
- Subnet group associations
- Resource dependencies
- Sensitive data (passwords)

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

### Secrets Encryption

**Encrypt Passwords:**
```bash
# Set password as secret
pulumi config set --secret dbPassword "my-secure-password"
```

In manifest:
```yaml
spec:
  password: "${PULUMI_CONFIG_PASSWORD}"
```

## Migration and Updates

### Engine Version Upgrades

**Minor Version Upgrades** (e.g., 14.9 → 14.10):
```yaml
# Update version
engine_version: "14.10"
```

Apply during maintenance window:
```bash
pulumi up --yes
```

**Major Version Upgrades** (e.g., 13 → 14):
1. Create snapshot
2. Test upgrade on snapshot clone
3. Plan downtime
4. Apply upgrade:
```yaml
engine_version: "14.10"
```

### Instance Class Changes

```yaml
# Scale up
instance_class: "db.m6i.2xlarge"  # From db.t3.medium
```

**Downtime**: Brief interruption during instance restart

### Storage Increase

```yaml
# Increase storage
allocated_storage_gb: 200  # From 100
```

**Downtime**: None (online operation)

**Note**: Cannot decrease storage size!

## Performance Optimization

### Query Performance

1. **Enable Performance Insights**
2. **Analyze slow queries**
3. **Add indexes** for frequently filtered columns
4. **Optimize queries** based on EXPLAIN output
5. **Cache frequently accessed data** in application layer

### Read Replicas

For read-heavy workloads, create read replicas:
```yaml
# Primary instance
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: primary-db

---
# Read replica (separate resource)
apiVersion: aws.openmcf.org/v1
kind: AwsRdsInstance
metadata:
  name: read-replica
spec:
  replicate_source_db: "primary-db"
```

### Connection Pooling

Implement connection pooling:
- **PostgreSQL**: PgBouncer
- **MySQL**: ProxySQL
- **Application-level**: HikariCP, c3p0

## Cost Optimization

### Strategies

1. **Right-size instances**: Start small, scale as needed
2. **Use Reserved Instances**: 30-70% savings for predictable workloads
3. **Delete unused snapshots**: Only keep necessary backups
4. **Turn off Multi-AZ for dev/test**: Cut costs in half
5. **Use t3 instances**: Burstable performance for variable workloads

### Cost Breakdown

**Example Production Instance:**
- **Instance**: db.t3.medium (~$50/month)
- **Storage**: 100 GiB GP3 (~$11.50/month)
- **Multi-AZ**: 2x instance cost (~$100/month total for instance)
- **Backups**: 100 GiB (~$9.50/month)

**Total**: ~$121/month for HA setup

## References

- [AWS RDS Documentation](https://docs.aws.amazon.com/rds/)
- [Pulumi AWS RDS Instance](https://www.pulumi.com/registry/packages/aws/api-docs/rds/instance/)
- [RDS Best Practices](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_BestPractices.html)
- [OpenMCF Architecture](https://github.com/plantonhq/openmcf/blob/main/architecture/deployment-component.md)

## Conclusion

This Pulumi module provides a production-ready abstraction for deploying AWS RDS instances with best practices built in. By handling subnet groups and instance configuration, it reduces the complexity of RDS deployment while preserving the full power of RDS features.

The module supports the full spectrum of RDS deployments:
- All major database engines (PostgreSQL, MySQL, MariaDB, Oracle, SQL Server)
- Multi-AZ high availability
- Storage encryption
- Custom parameter and option groups
- Network isolation in private subnets
- Flexible authentication options

For most use cases, the module's defaults provide a secure, reliable starting point that can be customized as needed.

