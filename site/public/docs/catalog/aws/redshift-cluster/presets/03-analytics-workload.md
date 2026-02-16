---
title: "High-Performance Analytics Cluster"
description: "This preset creates a 4-node RA3 Redshift cluster sized for large-scale analytical workloads. The ra3.4xlarge nodes each provide 12 vCPUs, 96 GiB RAM, and managed storage that automatically tiers..."
type: "preset"
rank: "03"
presetSlug: "03-analytics-workload"
componentSlug: "redshift-cluster"
componentTitle: "Redshift Cluster"
provider: "aws"
icon: "package"
order: 3
---

# High-Performance Analytics Cluster

This preset creates a 4-node RA3 Redshift cluster sized for large-scale analytical workloads. The ra3.4xlarge nodes each provide 12 vCPUs, 96 GiB RAM, and managed storage that automatically tiers data between local SSD and S3. Multi-AZ is enabled for automatic failover, concurrency scaling is set to 5 additional clusters for burst capacity, and two IAM roles are attached for S3 data loading and Redshift Spectrum queries against external tables.

## When to Use

- Data lake analytics querying petabytes of data across Redshift local tables and S3 external tables (Spectrum)
- Workloads with high concurrency requirements where multiple BI tools and users run queries simultaneously
- Environments requiring high availability with automatic failover across Availability Zones
- Teams that need 14-day snapshot retention for compliance or disaster recovery

## Key Configuration Choices

- **ra3.4xlarge node type** (`nodeType: ra3.4xlarge`) -- 12 vCPUs, 96 GiB RAM per node; managed storage scales independently up to petabytes
- **4-node cluster** (`numberOfNodes: 4`) -- Dedicated leader node with 4 compute nodes for parallel query execution across large datasets
- **Multi-AZ** (`multiAz: true`) -- Automatic failover to a standby cluster in a different AZ; requires RA3 node types
- **Concurrency scaling** (`parameters: max_concurrency_scaling_clusters=5`) -- Up to 5 additional transient clusters can be added automatically to handle query bursts
- **Managed password with CMK** (`manageMasterPassword: true`, `masterPasswordSecretKmsKeyId`) -- Password lifecycle managed by AWS Secrets Manager, encrypted with a customer-managed KMS key
- **Storage encryption with CMK** (`encrypted: true`, `kmsKeyId`) -- Data at rest encrypted with the specified customer-managed KMS key
- **Enhanced VPC routing** (`enhancedVpcRouting: true`) -- All COPY/UNLOAD traffic routed through the VPC for network-level controls
- **SSL required** (`parameters: require_ssl=true`) -- Rejects unencrypted client connections
- **User activity logging** (`parameters: enable_user_activity_logging=true`) -- Logs all SQL statements
- **CloudWatch audit logs** (`logging.logDestinationType: cloudwatch`) -- Connection, user activity, and user DDL logs streamed to CloudWatch Logs
- **14-day snapshot retention** (`automatedSnapshotRetentionPeriod: 14`) -- Two weeks of automated snapshots for point-in-time recovery
- **Deferred modifications** (`applyImmediately: false`) -- Configuration changes apply during the next maintenance window to avoid disrupting running queries
- **Sunday maintenance window** (`preferredMaintenanceWindow: "sun:02:00-sun:04:00"`) -- Maintenance scheduled during lowest-traffic hours
- **Two IAM roles** (`iamRoles`) -- One for S3 COPY/UNLOAD operations, one for Spectrum external table queries
- **Default IAM role** (`defaultIamRoleArn`) -- S3 access role used by default when SQL does not specify a role

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<kms-key-arn>` | ARN of the customer-managed KMS key for cluster encryption and Secrets Manager | AWS KMS console or `AwsKmsKey` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<redshift-s3-access-role-arn>` | ARN of the IAM role granting Redshift read/write access to S3 buckets | AWS IAM console or `AwsIamRole` status outputs |
| `<redshift-spectrum-role-arn>` | ARN of the IAM role granting Redshift Spectrum access to the Glue Data Catalog and S3 | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-single-node-dev** -- Use for development/testing with minimal cost and no production safeguards
- **02-multi-node-production** -- Use for standard production workloads that do not require Multi-AZ or Spectrum integration
