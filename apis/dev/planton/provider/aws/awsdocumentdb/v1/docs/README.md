# Deploying AWS DocumentDB: From Manual Operations to Production Infrastructure as Code

## Introduction

Amazon DocumentDB represents AWS's answer to the growing demand for MongoDB-compatible document databases in the cloud—a fully managed service that provides fast, scalable, and highly available document storage with an API compatible with MongoDB 4.0 and 5.0 workloads. While MongoDB Atlas offers the "official" MongoDB experience, DocumentDB provides an AWS-native alternative deeply integrated with the AWS ecosystem.

The fundamental challenge isn't whether you can create a DocumentDB cluster—the AWS console makes that straightforward. The challenge is deploying DocumentDB **consistently**, **securely**, and **reproducibly** across multiple environments while managing the complexity of networking, encryption, backup policies, and credential management that production workloads demand.

This document explores the landscape of DocumentDB deployment methods, from anti-patterns that plague many organizations to production-ready infrastructure-as-code solutions. More importantly, it explains why Planton standardizes on specific approaches and what that means for teams building reliable document database infrastructure.

## Evolution and Historical Context

### The Rise of Document Databases

The early 2010s saw an explosion in document database adoption, led by MongoDB. Unlike traditional relational databases with rigid schemas, document databases store data in flexible JSON-like documents, making them ideal for:

- **Rapidly evolving schemas**: Applications where data structures change frequently
- **Hierarchical data**: Nested objects that map naturally to document structures
- **Developer velocity**: JavaScript developers found MongoDB's JSON documents intuitive
- **Horizontal scaling**: Built-in sharding for distributed workloads

MongoDB became the de facto standard for NoSQL document databases, but running it reliably in production presented challenges:

- **Operational complexity**: Managing replica sets, shards, elections, and failover
- **Security configuration**: Authentication, encryption, network isolation
- **Backup and recovery**: Point-in-time recovery, disaster recovery planning
- **Performance tuning**: Index management, query optimization, resource provisioning

### AWS DocumentDB: MongoDB Compatibility, AWS Operations

In January 2019, AWS launched DocumentDB with MongoDB compatibility—a fully managed service that implements the MongoDB 3.6 API (later expanded to 4.0 and 5.0) on top of AWS's proven distributed storage architecture, similar to Aurora.

**Key architectural differences from MongoDB:**

- **Storage layer**: DocumentDB uses a distributed, fault-tolerant storage layer (similar to Aurora) that automatically replicates data 6 ways across 3 Availability Zones
- **Compute-storage separation**: Storage scales automatically; you only provision compute instances
- **AWS integration**: Native VPC support, IAM authentication, CloudWatch metrics, AWS Backup
- **Managed operations**: Automated patching, backups, failover—no MongoDB operational expertise required

**Trade-offs:**

- **Not open source**: Unlike MongoDB Community Edition, DocumentDB is proprietary
- **MongoDB compatibility gaps**: Some MongoDB features aren't supported (see compatibility matrix)
- **AWS lock-in**: Deep AWS integration means limited portability
- **Pricing model**: Different from MongoDB Atlas; requires AWS cost management expertise

### The Deployment Challenge

Whether you choose MongoDB Atlas or AWS DocumentDB, the deployment challenge is similar: how do you provision, configure, and maintain document database infrastructure reliably across environments?

Manual operations (console clicking, ad-hoc CLI commands) work for exploration but fail at scale. Infrastructure as Code—defining your database infrastructure in version-controlled, reviewable, testable code—is the production answer.

## The Deployment Maturity Spectrum

### Level 0: Manual Console Operations (The Anti-Pattern)

The AWS Management Console provides the quickest path to a running DocumentDB cluster. Navigate to DocumentDB, click "Create cluster," fill in the configuration form, and within minutes you have a functioning document database.

For learning AWS or running a proof-of-concept, this approach works. For anything beyond exploration, it's a trap.

**Why console operations fail at scale:**

1. **No reproducibility**: Creating a cluster through the GUI involves dozens of decisions—subnet groups, security groups, parameter groups, instance classes, encryption settings—none of which is captured in a reviewable, versionable format.

2. **Configuration drift**: When someone manually tweaks a setting (maybe adjusting backup windows or modifying parameter groups), those changes exist only in AWS's state. Six months later, nobody remembers why that change was made.

3. **Environment inconsistency**: Your production cluster was created manually. Staging was created manually. Development was created manually. They're probably all configured differently in subtle, undocumented ways.

4. **Security blind spots**: One checkbox—"Publicly accessible"—can expose your database to the internet. Without code review, these mistakes slip through.

**Verdict:** Acceptable for learning and one-off experiments. Unacceptable for any environment that matters.

### Level 1: AWS CLI Scripts (Scriptable but Stateless)

The AWS CLI represents the first step toward automation:

```bash
# Create a DocumentDB cluster
aws docdb create-db-cluster \
  --db-cluster-identifier my-docdb-cluster \
  --engine docdb \
  --engine-version 5.0.0 \
  --master-username docdbadmin \
  --master-user-password "SecurePassword123!" \
  --vpc-security-group-ids sg-abc123 \
  --db-subnet-group-name my-docdb-subnet-group \
  --storage-encrypted \
  --backup-retention-period 7

# Add instances to the cluster
aws docdb create-db-instance \
  --db-instance-identifier my-docdb-instance-1 \
  --db-instance-class db.r6g.large \
  --db-cluster-identifier my-docdb-cluster \
  --engine docdb

aws docdb create-db-instance \
  --db-instance-identifier my-docdb-instance-2 \
  --db-instance-class db.r6g.large \
  --db-cluster-identifier my-docdb-cluster \
  --engine docdb
```

This is better—commands can be saved, version controlled, and repeated. But CLI scripting is **imperative** and **stateless**:

- The CLI doesn't know what already exists
- Running the script twice fails (resources already exist)
- Updating configurations requires different commands (`modify-db-cluster`)
- Deletion requires yet another command, and you must remember all resources
- No drift detection (manual changes aren't tracked)

**Operational challenges:**

- Complex dependency ordering (subnet groups before clusters, clusters before instances)
- Credential management is your problem (passwords in scripts or environment variables)
- State tracking is manual (maintain lists of what exists where)
- No preview of changes before applying

**Verdict:** Useful for automation scripts and quick operations. Insufficient for production infrastructure requiring lifecycle management.

### Level 2: Configuration Management Tools (Ansible)

Ansible adds **declarative intent** to imperative scripting:

```yaml
- name: Ensure DocumentDB cluster exists
  amazon.aws.docdb_cluster:
    db_cluster_identifier: my-docdb-cluster
    engine: docdb
    engine_version: "5.0.0"
    master_username: docdbadmin
    master_user_password: "{{ vault_docdb_password }}"
    vpc_security_group_ids:
      - sg-abc123
    db_subnet_group_name: my-docdb-subnet-group
    storage_encrypted: true
    backup_retention_period: 7
    state: present
```

Ansible ensures the cluster exists with specified configuration, creating only if needed (idempotent). However, configuration management tools have limitations for cloud infrastructure:

- **Limited state management**: No native understanding of cloud resource dependencies
- **Basic drift detection**: Less sophisticated than dedicated IaC tools
- **No plan/preview**: Changes apply directly without showing what will happen first

**Verdict:** Valuable for orchestration, but typically paired with dedicated IaC tools for infrastructure management.

### Level 3: Production Infrastructure as Code (Terraform, Pulumi)

Modern IaC tools provide the rigor required for production operations:

- **Declarative configuration**: Define desired state, not steps
- **State tracking**: Know what exists, detect drift
- **Dependency graphs**: Automatic ordering of resource operations
- **Change previewing**: See what will change before applying
- **Concurrent safety**: Locking prevents conflicting changes

## Comparing Production-Ready IaC Tools

### Terraform/OpenTofu: The Industry Standard

Terraform has become the de facto standard for infrastructure-as-code:

```hcl
resource "aws_docdb_cluster" "main" {
  cluster_identifier      = "my-docdb-cluster"
  engine                  = "docdb"
  engine_version          = "5.0.0"
  master_username         = "docdbadmin"
  master_password         = var.master_password
  
  db_subnet_group_name    = aws_docdb_subnet_group.main.name
  vpc_security_group_ids  = [aws_security_group.docdb.id]
  
  storage_encrypted       = true
  kms_key_id              = aws_kms_key.docdb.arn
  
  backup_retention_period = 7
  preferred_backup_window = "03:00-04:00"
  
  deletion_protection     = true
  skip_final_snapshot     = false
  final_snapshot_identifier = "my-docdb-final-snapshot"
  
  enabled_cloudwatch_logs_exports = ["audit", "profiler"]
  
  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

resource "aws_docdb_cluster_instance" "instances" {
  count              = 3
  identifier         = "my-docdb-instance-${count.index + 1}"
  cluster_identifier = aws_docdb_cluster.main.id
  instance_class     = "db.r6g.large"
  
  auto_minor_version_upgrade = true
  
  tags = {
    Environment = "production"
    ManagedBy   = "terraform"
  }
}

resource "aws_docdb_subnet_group" "main" {
  name       = "my-docdb-subnet-group"
  subnet_ids = var.private_subnet_ids
  
  tags = {
    Name = "DocumentDB Subnet Group"
  }
}
```

**Strengths:**
- **Massive ecosystem**: Thousands of providers and community modules
- **Multi-cloud support**: Same tooling for AWS, GCP, Azure
- **Mature state management**: Remote backends with locking
- **Battle-tested at scale**: Enterprises managing thousands of resources

**OpenTofu** is the community fork maintaining Terraform compatibility with open-source governance.

**When to choose Terraform/OpenTofu:**
- Building multi-cloud infrastructure
- Team values mature, well-documented ecosystem
- Preference for declarative configuration over general-purpose programming

### Pulumi: Infrastructure as Real Code

Pulumi uses familiar programming languages for infrastructure:

```go
package main

import (
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/docdb"
    "github.com/pulumi/pulumi-aws/sdk/v6/go/aws/ec2"
    "github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
    pulumi.Run(func(ctx *pulumi.Context) error {
        // Create subnet group
        subnetGroup, err := docdb.NewSubnetGroup(ctx, "docdb-subnet-group", &docdb.SubnetGroupArgs{
            SubnetIds: pulumi.StringArray{
                pulumi.String("subnet-12345678"),
                pulumi.String("subnet-87654321"),
            },
        })
        if err != nil {
            return err
        }

        // Create security group
        securityGroup, err := ec2.NewSecurityGroup(ctx, "docdb-sg", &ec2.SecurityGroupArgs{
            VpcId: pulumi.String("vpc-12345678"),
            Ingress: ec2.SecurityGroupIngressArray{
                &ec2.SecurityGroupIngressArgs{
                    Protocol:   pulumi.String("tcp"),
                    FromPort:   pulumi.Int(27017),
                    ToPort:     pulumi.Int(27017),
                    CidrBlocks: pulumi.StringArray{pulumi.String("10.0.0.0/16")},
                },
            },
        })
        if err != nil {
            return err
        }

        // Create DocumentDB cluster
        cluster, err := docdb.NewCluster(ctx, "docdb-cluster", &docdb.ClusterArgs{
            ClusterIdentifier:       pulumi.String("my-docdb-cluster"),
            Engine:                  pulumi.String("docdb"),
            EngineVersion:           pulumi.String("5.0.0"),
            MasterUsername:          pulumi.String("docdbadmin"),
            MasterPassword:          cfg.RequireSecret("docdb_password"),
            DbSubnetGroupName:       subnetGroup.Name,
            VpcSecurityGroupIds:     pulumi.StringArray{securityGroup.ID()},
            StorageEncrypted:        pulumi.Bool(true),
            BackupRetentionPeriod:   pulumi.Int(7),
            PreferredBackupWindow:   pulumi.String("03:00-04:00"),
            DeletionProtection:      pulumi.Bool(true),
            SkipFinalSnapshot:       pulumi.Bool(false),
            FinalSnapshotIdentifier: pulumi.String("my-docdb-final-snapshot"),
            EnabledCloudwatchLogsExports: pulumi.StringArray{
                pulumi.String("audit"),
                pulumi.String("profiler"),
            },
        })
        if err != nil {
            return err
        }

        // Create cluster instances
        for i := 0; i < 3; i++ {
            _, err := docdb.NewClusterInstance(ctx, fmt.Sprintf("docdb-instance-%d", i+1), &docdb.ClusterInstanceArgs{
                Identifier:             pulumi.Sprintf("my-docdb-instance-%d", i+1),
                ClusterIdentifier:      cluster.ID(),
                InstanceClass:          pulumi.String("db.r6g.large"),
                AutoMinorVersionUpgrade: pulumi.Bool(true),
            })
            if err != nil {
                return err
            }
        }

        // Export outputs
        ctx.Export("clusterEndpoint", cluster.Endpoint)
        ctx.Export("clusterReaderEndpoint", cluster.ReaderEndpoint)
        
        return nil
    })
}
```

**Strengths:**
- **Use familiar languages**: Go, TypeScript, Python, C#, Java
- **Full IDE support**: Type checking, autocompletion, refactoring
- **Programming language power**: Loops, conditionals, functions, testing
- **Excellent secrets management**: Encrypted in state by default

**When to choose Pulumi:**
- Team consists of software developers preferring code over DSLs
- Complex logic needed in infrastructure definitions
- Want infrastructure tightly coupled with application code

### CloudFormation: AWS-Native IaC

CloudFormation is AWS's original IaC service:

```yaml
Resources:
  DocumentDBCluster:
    Type: AWS::DocDB::DBCluster
    Properties:
      DBClusterIdentifier: my-docdb-cluster
      Engine: docdb
      EngineVersion: "5.0.0"
      MasterUsername: docdbadmin
      MasterUserPassword: !Ref MasterPassword
      DBSubnetGroupName: !Ref DocDBSubnetGroup
      VpcSecurityGroupIds:
        - !Ref DocDBSecurityGroup
      StorageEncrypted: true
      BackupRetentionPeriod: 7
      DeletionProtection: true
      EnableCloudwatchLogsExports:
        - audit
        - profiler
```

**Strengths:**
- No external state management
- Zero additional cost
- Immediate AWS feature support
- Deep AWS integration
- Robust rollback on failures

**Considerations:**
- AWS-only (no multi-cloud)
- Verbose templates
- Limited modularity

**When to choose CloudFormation:**
- All-in on AWS with no multi-cloud requirements
- Want minimal external dependencies
- Prefer AWS-native solutions

## The Planton Approach

Planton provides a minimal, validated API that abstracts DocumentDB deployment complexity while supporting both Terraform and Pulumi as first-class deployment targets.

### The 80/20 Configuration Philosophy

AWS DocumentDB exposes dozens of configuration parameters. Most teams need about 20% of them for 80% of use cases. Planton's API reflects this philosophy.

### Essential Fields (What Planton Exposes)

#### Region
- **region**: The AWS region where the resource will be created (required)

#### Networking
- **subnetIds**: Private subnets for the DB subnet group (>=2 for HA)
- **dbSubnetGroupName**: Alternative to subnet_ids if group already exists
- **securityGroupIds**: Security groups controlling database access
- **allowedCidrBlocks**: IPv4 CIDRs for ingress rules
- **vpcId**: VPC for cluster networking context

#### Engine Configuration
- **engineVersion**: DocumentDB version ("4.0.0", "5.0.0")
- **port**: Connection port (default: 27017)

#### Compute and Scaling
- **instanceCount**: Number of instances in the cluster
- **instanceClass**: Instance type (db.r5.large, db.r6g.xlarge, etc.)

#### Authentication
- **masterUsername**: Master user name
- **masterPassword**: Master password (should use secrets management)

#### Encryption
- **storageEncrypted**: Enable encryption at rest (default: true)
- **kmsKeyId**: Customer-managed KMS key for encryption

#### Backup and Recovery
- **backupRetentionPeriod**: Days to retain backups (1-35)
- **preferredBackupWindow**: Daily backup window
- **skipFinalSnapshot**: Whether to skip final snapshot on deletion
- **finalSnapshotIdentifier**: Identifier for final snapshot

#### Maintenance
- **preferredMaintenanceWindow**: Weekly maintenance window
- **autoMinorVersionUpgrade**: Enable automatic minor version upgrades

#### Protection
- **deletionProtection**: Prevent accidental cluster deletion

#### Monitoring
- **enabledCloudwatchLogsExports**: Log types to export (audit, profiler)

#### Custom Parameters
- **clusterParameterGroupName**: Custom parameter group
- **clusterParameters**: Custom parameters for the group

### What We Default or Omit

Many settings have sensible defaults or are managed at the infrastructure platform level:

- **Engine**: Always "docdb" (this is DocumentDB-specific)
- **Storage type**: DocumentDB uses distributed storage (no configuration needed)
- **Instance availability zones**: Automatically distributed by AWS
- **Performance Insights**: Not available for DocumentDB
- **Read replicas**: Managed through instanceCount (all instances can serve reads)

### Why Dual IaC Support

Different teams have different needs. Rather than forcing a choice:

1. **Define once**: Specify configuration in Planton's validated API schema
2. **Deploy with your preferred tool**: Planton generates Terraform HCL or Pulumi code
3. **Maintain flexibility**: Teams choose the IaC tool that fits their culture

## Production Best Practices

### High Availability: Multiple Instances Are Non-Negotiable

DocumentDB clusters support up to 15 instances (1 primary, up to 15 replicas). For production:

- **Minimum 3 instances**: 1 primary + 2 replicas across different AZs
- **Automatic failover**: If primary fails, replica promotes automatically (typically <30 seconds)
- **Read scaling**: All replicas can serve read traffic

```yaml
spec:
  region: us-east-1
  instanceCount: 3
  instanceClass: db.r6g.large
```

### Network Isolation: Private Subnets Only

Production DocumentDB clusters should **never** have public access:

- Place in private subnets across multiple AZs
- Security groups allow access only from application tier
- Use bastion hosts or VPN for administrative access
- Consider AWS PrivateLink for cross-VPC access

### Encryption: Always, Everywhere

**At-rest encryption** (via KMS):
- Enable for all clusters containing sensitive data
- Use customer-managed keys for compliance requirements
- Cannot be enabled after cluster creation—start encrypted

**In-transit encryption** (TLS):
- DocumentDB enforces TLS by default
- Applications must use TLS-enabled connection strings
- Download the AWS DocumentDB CA certificate for connections

```yaml
spec:
  region: us-east-1
  storageEncrypted: true
  kmsKeyId:
    valueFrom:
      kind: AwsKmsKey
      name: docdb-key
      fieldPath: status.outputs.key_arn
```

### Backup Strategy

**Automated backups**:
- Enabled by default with configurable retention (1-35 days)
- Point-in-time recovery within retention window
- Backups stored redundantly

**Manual snapshots**:
- Create before major changes (engine upgrades, migrations)
- Can be copied cross-region for disaster recovery
- Retained until explicitly deleted

**Final snapshots**:
- Create when deleting clusters (unless explicitly skipped)
- Essential for data preservation compliance

```yaml
spec:
  region: us-east-1
  backupRetentionPeriod: 14
  preferredBackupWindow: "03:00-04:00"
  skipFinalSnapshot: false
  finalSnapshotIdentifier: docdb-prod-final-snapshot
```

### Credential Management: No Hardcoded Passwords

**Best practices:**
- Use secrets management (AWS Secrets Manager, Vault)
- Reference secrets in manifests via secret syntax
- Rotate credentials regularly
- Consider IAM database authentication where supported

```yaml
spec:
  region: us-east-1
  masterPassword: ${secrets-group/docdb/MASTER_PASSWORD}
```

### Monitoring and Observability

**CloudWatch metrics** to monitor:
- `CPUUtilization`: Instance CPU usage
- `DatabaseConnections`: Active connections
- `ReadLatency` / `WriteLatency`: Query latency
- `FreeableMemory`: Available memory
- `VolumeBytesUsed`: Storage consumption

**CloudWatch Logs** to export:
- **audit**: Track database activity for compliance
- **profiler**: Analyze slow queries and performance

```yaml
spec:
  region: us-east-1
  enabledCloudwatchLogsExports:
    - audit
    - profiler
```

### Cost Optimization

**Instance right-sizing:**
- Start with smaller instances, scale based on actual usage
- Use Graviton instances (db.r6g) for ~20% cost savings
- Monitor and adjust based on CPU/memory utilization

**Reserved instances:**
- 1-year commitment: ~35% savings
- 3-year commitment: ~55% savings
- Ideal for stable production workloads

**Non-production environments:**
- Use smaller instance classes
- Consider single-instance clusters (no HA)
- Stop clusters outside business hours (up to 7 days)

## MongoDB Compatibility Considerations

DocumentDB implements the MongoDB API but is not MongoDB. Key differences:

### Supported Features
- Basic CRUD operations
- Aggregation pipeline (most operators)
- Indexes (single field, compound, geospatial, text)
- Transactions (single and multi-document)
- Change streams

### Notable Limitations
- **No sharding**: DocumentDB handles scaling differently
- **No full-text search**: Use Amazon OpenSearch instead
- **Limited aggregation operators**: Some pipeline stages unsupported
- **No MongoDB Realm**: Cloud sync and serverless functions not available
- **Driver version requirements**: Use specific driver versions for compatibility

### Migration Considerations
- Test application against DocumentDB before migration
- Use AWS Database Migration Service for data migration
- Review MongoDB compatibility matrix for unsupported features
- Plan for application changes if using unsupported features

## Conclusion

AWS DocumentDB provides a compelling option for teams needing MongoDB-compatible document databases with AWS-native operations. The managed service eliminates operational complexity—no replica set management, no manual backups, no security patch scheduling.

The deployment challenge, however, remains: manual console operations don't scale, and ad-hoc scripting creates technical debt. Infrastructure as Code—whether Terraform, Pulumi, or CloudFormation—provides the foundation for reliable, reproducible database infrastructure.

Planton's approach simplifies DocumentDB deployment by:
1. **Focusing on essential configuration**: The 80/20 fields that matter for production
2. **Providing validated APIs**: Catch configuration errors before deployment
3. **Supporting multiple IaC tools**: Choose Terraform or Pulumi based on team preference
4. **Enforcing best practices**: Encryption, network isolation, and backup policies by default

Whether migrating from self-managed MongoDB, evaluating DocumentDB against MongoDB Atlas, or standardizing on a document database platform, the key is treating database infrastructure as code—reviewable, testable, version-controlled, and reproducible.

## References

- [AWS DocumentDB Developer Guide](https://docs.aws.amazon.com/documentdb/latest/developerguide/what-is.html)
- [DocumentDB Instance Classes](https://docs.aws.amazon.com/documentdb/latest/developerguide/db-instance-classes.html)
- [DocumentDB Engine Versions](https://docs.aws.amazon.com/documentdb/latest/developerguide/release-notes.html)
- [MongoDB Compatibility](https://docs.aws.amazon.com/documentdb/latest/developerguide/mongo-apis.html)
- [DocumentDB Best Practices](https://docs.aws.amazon.com/documentdb/latest/developerguide/best-practices.html)
- [Terraform AWS DocumentDB Provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/docdb_cluster)
- [Pulumi AWS DocumentDB](https://www.pulumi.com/registry/packages/aws/api-docs/docdb/)
