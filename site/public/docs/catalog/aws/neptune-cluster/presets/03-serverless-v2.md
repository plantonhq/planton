---
title: "Neptune Serverless v2"
description: "This preset creates a Neptune cluster with Serverless v2 auto-scaling, where compute capacity automatically adjusts between 1.0 and 16.0 Neptune Capacity Units (NCUs) based on workload demand. Ideal..."
type: "preset"
rank: "03"
presetSlug: "03-serverless-v2"
componentSlug: "neptune-cluster"
componentTitle: "Neptune Cluster"
provider: "aws"
icon: "package"
order: 3
---

# Neptune Serverless v2

This preset creates a Neptune cluster with Serverless v2 auto-scaling, where compute capacity automatically adjusts between 1.0 and 16.0 Neptune Capacity Units (NCUs) based on workload demand. Ideal for applications with variable, unpredictable, or spiky traffic. Neptune supports Gremlin and SPARQL for graph workloads such as recommendation engines, fraud detection, and knowledge graphs.

## When to Use

- Applications with variable or unpredictable graph query traffic
- Development and staging environments where cost should track actual usage
- Workloads that need Neptune's graph capabilities but don't want to manage instance sizes
- Serverless application architectures (Lambda, Step Functions) with intermittent graph access
- Batch or ETL jobs that load graph data periodically

## Key Configuration Choices

- **Neptune Serverless** (`instanceClass: db.serverless`) — Auto-scaling compute; no fixed instance size
- **Serverless v2 scaling** (`serverlessV2Scaling`) — Auto-scales between 1.0 and 16.0 NCUs; 1 NCU ≈ 2 GiB RAM
- **Minimum 1.0 NCU** (`minCapacity: 1.0`) — Baseline capacity; Neptune Serverless does not scale to zero
- **Maximum 16.0 NCU** (`maxCapacity: 16.0`) — Handles significant traffic spikes; adjust based on your peak load
- **IAM database authentication** (`iamDatabaseAuthenticationEnabled: true`) — Secure access without master password
- **Encrypted storage** (`storageEncrypted: true`) — Data at rest encryption
- **7-day backup retention** — Automated daily backups
- **Skip final snapshot** (`skipFinalSnapshot: true`) — No snapshot on deletion; suitable for non-prod

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing Neptune port (8182) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<vpc-id>` | VPC ID where the cluster will be deployed | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **01-graph-database** — Use instead for provisioned Neptune with predictable capacity
- **02-high-availability** — Use instead for production with 2 instances, deletion protection, and IAM auth
