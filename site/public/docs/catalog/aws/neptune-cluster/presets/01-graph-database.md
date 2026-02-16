---
title: "Graph Database (Standard Provisioned)"
description: "This preset creates a standard provisioned Neptune cluster with a single instance (db.r6g.large). Ideal for development, testing, or moderate graph workloads that need Gremlin or SPARQL support..."
type: "preset"
rank: "01"
presetSlug: "01-graph-database"
componentSlug: "neptune-cluster"
componentTitle: "Neptune Cluster"
provider: "aws"
icon: "package"
order: 1
---

# Graph Database (Standard Provisioned)

This preset creates a standard provisioned Neptune cluster with a single instance (db.r6g.large). Ideal for development, testing, or moderate graph workloads that need Gremlin or SPARQL support without high availability requirements.

## When to Use

- Development and testing of graph applications (recommendation engines, fraud detection, knowledge graphs)
- Prototyping with Gremlin or SPARQL before scaling to production
- Workloads with predictable, moderate traffic that don't require multiple read replicas
- Cost-sensitive environments where a single instance is sufficient

## Key Configuration Choices

- **Single instance** (`instanceCount: 1`) — No read replicas; primary handles all reads and writes
- **db.r6g.large** (`instanceClass`) — 2 vCPUs, 16 GiB RAM; memory-optimized for graph traversal workloads
- **Engine 1.3.0.0** (`engineVersion`) — Latest Neptune engine; supports Gremlin and SPARQL
- **Encrypted storage** (`storageEncrypted: true`) — Data at rest encryption with AWS-managed key
- **7-day backup retention** — Automated daily backups for point-in-time recovery
- **Skip final snapshot** (`skipFinalSnapshot: true`) — No snapshot on deletion; simplifies dev environment cleanup
- **No deletion protection** — Cluster can be deleted freely
- **Default port 8182** — Standard Neptune port for Gremlin/SPARQL connections

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing Neptune port (8182) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<vpc-id>` | VPC ID where the cluster will be deployed | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **02-high-availability** — Use instead for production with 2 instances, deletion protection, and IAM auth
- **03-serverless-v2** — Use instead for variable or unpredictable traffic with auto-scaling
