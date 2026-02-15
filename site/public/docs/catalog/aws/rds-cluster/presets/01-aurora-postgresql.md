---
title: "Aurora PostgreSQL Cluster"
description: "This preset creates a production-ready Aurora PostgreSQL cluster with RDS-managed master password (stored in Secrets Manager), encrypted storage, deletion protection, 7-day backup retention, and..."
type: "preset"
rank: "01"
presetSlug: "01-aurora-postgresql"
componentSlug: "rds-cluster"
componentTitle: "RDS Cluster"
provider: "aws"
icon: "package"
order: 1
---

# Aurora PostgreSQL Cluster

This preset creates a production-ready Aurora PostgreSQL cluster with RDS-managed master password (stored in Secrets Manager), encrypted storage, deletion protection, 7-day backup retention, and PostgreSQL logs exported to CloudWatch. Aurora PostgreSQL is the most popular engine choice for new relational database deployments on AWS.

## When to Use

- Production relational databases using PostgreSQL-compatible SQL
- Applications requiring high availability (Aurora automatically replicates across 3 AZs)
- Workloads benefiting from Aurora's performance improvements over standard PostgreSQL (up to 3x throughput)

## Key Configuration Choices

- **Aurora PostgreSQL 15.4** (`engine: aurora-postgresql`, `engineVersion: "15.4"`) -- Update to the latest minor version for your environment
- **Managed password** (`manageMasterUserPassword: true`) -- RDS creates and rotates the master password in AWS Secrets Manager automatically
- **Encrypted storage** (`storageEncrypted: true`) -- Data at rest encrypted with AWS-managed key; specify `kmsKeyId` for customer-managed key
- **Deletion protection** (`deletionProtection: true`) -- Prevents accidental cluster deletion
- **7-day backup retention** -- Automated daily backups retained for 1 week; increase up to 35 days for compliance
- **Final snapshot required** (`skipFinalSnapshot: false`) -- Creates a snapshot before deletion for recovery
- **CloudWatch logs** (`enabledCloudwatchLogsExports: [postgresql]`) -- PostgreSQL logs exported to CloudWatch for debugging and monitoring

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing database port (5432) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<database-name>` | Name of the initial database to create | Your application configuration |
| `<final-snapshot-name>` | Identifier for the final snapshot (e.g., `myapp-final-2026-02-14`) | Your naming convention |

## Related Presets

- **02-aurora-mysql** -- Use instead for MySQL-compatible workloads
- **03-aurora-serverless-v2** -- Use instead for workloads with variable or unpredictable traffic patterns
