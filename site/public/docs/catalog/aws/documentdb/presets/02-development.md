---
title: "Development DocumentDB Cluster"
description: "This preset creates a single-instance DocumentDB cluster for development and testing. It uses a smaller instance class (db.t3.medium) and skips the final snapshot on deletion to simplify teardown...."
type: "preset"
rank: "02"
presetSlug: "02-development"
componentSlug: "documentdb"
componentTitle: "DocumentDB"
provider: "aws"
icon: "package"
order: 2
---

# Development DocumentDB Cluster

This preset creates a single-instance DocumentDB cluster for development and testing. It uses a smaller instance class (db.t3.medium) and skips the final snapshot on deletion to simplify teardown. All other fields use defaults (engine 5.0, port 27017, encrypted storage).

## When to Use

- Development and testing environments where high availability is not needed
- Prototyping with a MongoDB-compatible database
- Cost-sensitive environments (single instance is ~3x cheaper than a 3-instance production cluster)

## Key Configuration Choices

- **Single instance** (`instanceCount: 1`) -- No replicas; no automatic failover
- **db.t3.medium** (`instanceClass`) -- 2 vCPUs, 4 GiB RAM; burstable and cost-effective
- **Skip final snapshot** (`skipFinalSnapshot: true`) -- No snapshot on deletion; simplifies dev environment cleanup
- **No deletion protection** -- Dev clusters can be deleted freely
- **Defaults preserved** -- Engine 5.0, port 27017, encrypted storage, 7-day backups still apply

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing DocumentDB port (27017) | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<vpc-id>` | VPC ID | AWS VPC console or `AwsVpc` status outputs |
| `<master-password>` | Master database password | Your credential management system |

## Related Presets

- **01-production-ha** -- Use instead for production with 3 instances and high availability
