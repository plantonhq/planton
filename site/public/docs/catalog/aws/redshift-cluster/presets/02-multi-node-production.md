---
title: "Multi-Node Production Data Warehouse"
description: "This preset creates a 2-node RA3 Redshift cluster configured for production workloads. RA3 nodes decouple compute and storage by automatically tiering data between local SSD and Amazon S3, so you can..."
type: "preset"
rank: "02"
presetSlug: "02-multi-node-production"
componentSlug: "redshift-cluster"
componentTitle: "Redshift Cluster"
provider: "aws"
icon: "package"
order: 2
---

# Multi-Node Production Data Warehouse

This preset creates a 2-node RA3 Redshift cluster configured for production workloads. RA3 nodes decouple compute and storage by automatically tiering data between local SSD and Amazon S3, so you can scale each dimension independently. The cluster enforces encryption with a customer-managed KMS key, requires SSL for all connections, logs all user activity to CloudWatch, and takes a final snapshot before deletion.

## When to Use

- Production data warehouses serving BI dashboards and reporting
- Workloads requiring compliance controls: encryption at rest with CMK, SSL enforcement, and audit logging
- Environments where data volume may grow beyond local SSD capacity (RA3 managed storage scales to the petabyte range)

## Key Configuration Choices

- **ra3.xlplus node type** (`nodeType: ra3.xlplus`) -- Managed storage nodes; compute and storage scale independently; recommended for most production workloads
- **2-node cluster** (`numberOfNodes: 2`) -- Multi-node topology with a dedicated leader node and 2 compute nodes for parallel query execution
- **Managed password with CMK** (`manageMasterPassword: true`, `masterPasswordSecretKmsKeyId`) -- AWS Secrets Manager manages the password, encrypted with a customer-managed KMS key
- **Storage encryption with CMK** (`encrypted: true`, `kmsKeyId`) -- Data at rest encrypted with the specified customer-managed KMS key
- **Enhanced VPC routing** (`enhancedVpcRouting: true`) -- All COPY/UNLOAD traffic routed through the VPC; enables VPC flow log monitoring and endpoint policies
- **SSL required** (`parameters: require_ssl=true`) -- Rejects unencrypted client connections
- **User activity logging** (`parameters: enable_user_activity_logging=true`) -- Logs all SQL statements executed by users
- **CloudWatch audit logs** (`logging.logDestinationType: cloudwatch`) -- Connection, user activity, and user DDL logs streamed to CloudWatch Logs
- **7-day snapshot retention** (`automatedSnapshotRetentionPeriod: 7`) -- Automated snapshots kept for one week
- **Final snapshot on deletion** (`skipFinalSnapshot: false`) -- Creates `my-prod-warehouse-final` before cluster deletion
- **Saturday maintenance window** (`preferredMaintenanceWindow: "sat:03:00-sat:04:00"`) -- Maintenance scheduled during low-traffic hours
- **S3 access IAM role** (`iamRoles`) -- Attached role allows COPY/UNLOAD from S3

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<kms-key-arn>` | ARN of the customer-managed KMS key for cluster encryption and Secrets Manager | AWS KMS console or `AwsKmsKey` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<redshift-s3-access-role-arn>` | ARN of the IAM role granting Redshift read/write access to S3 | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-single-node-dev** -- Use for development/testing with minimal cost and no production safeguards
- **03-analytics-workload** -- Use for large-scale analytics with Multi-AZ, concurrency scaling, and Spectrum integration
