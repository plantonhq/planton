---
title: "Production HA DocumentDB Cluster"
description: "This preset creates a highly available DocumentDB cluster with 3 instances (1 primary + 2 replicas) across Availability Zones. Storage is encrypted, backups are retained for 7 days, and deletion..."
type: "preset"
rank: "01"
presetSlug: "01-production-ha"
componentSlug: "documentdb"
componentTitle: "DocumentDB"
provider: "aws"
icon: "package"
order: 1
---

# Production HA DocumentDB Cluster

This preset creates a highly available DocumentDB cluster with 3 instances (1 primary + 2 replicas) across Availability Zones. Storage is encrypted, backups are retained for 7 days, and deletion protection is enabled. DocumentDB is a MongoDB-compatible document database for applications that need the MongoDB API with managed infrastructure.

## When to Use

- Production MongoDB-compatible workloads requiring high availability
- Applications using the MongoDB driver that need a managed, scalable document database
- Workloads requiring encrypted storage and automated backups

## Key Configuration Choices

- **3 instances** (`instanceCount: 3`) -- Primary writer + 2 read replicas across AZs for automatic failover
- **db.r6g.large** (`instanceClass`) -- 2 vCPUs, 16 GiB RAM; memory-optimized for document workloads
- **Engine 5.0** (`engineVersion: "5.0.0"`) -- MongoDB 5.0 compatible; latest DocumentDB engine
- **Encrypted storage** (`storageEncrypted: true`) -- Data at rest encryption with AWS-managed key
- **7-day backup retention** -- Automated daily backups for point-in-time recovery
- **Deletion protection** -- Prevents accidental cluster deletion
- **Final snapshot on deletion** (`skipFinalSnapshot: false`) -- Creates a recovery snapshot before deletion

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing DocumentDB port (27017) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<vpc-id>` | VPC ID where the cluster will be deployed | AWS VPC console or `AwsVpc` status outputs |
| `<master-password>` | Master database password (store securely) | Your credential management system |

## Related Presets

- **02-development** -- Use instead for single-instance dev/test environments with minimal cost
