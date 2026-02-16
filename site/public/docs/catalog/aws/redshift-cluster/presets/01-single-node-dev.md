---
title: "Single-Node Development Cluster"
description: "This preset creates a single-node Redshift cluster on the dc2.large instance type for development and testing. The single-node topology combines the leader and compute roles on one node, keeping..."
type: "preset"
rank: "01"
presetSlug: "01-single-node-dev"
componentSlug: "redshift-cluster"
componentTitle: "Redshift Cluster"
provider: "aws"
icon: "package"
order: 1
---

# Single-Node Development Cluster

This preset creates a single-node Redshift cluster on the dc2.large instance type for development and testing. The single-node topology combines the leader and compute roles on one node, keeping costs low while providing a functional SQL analytics environment. No final snapshot is taken on deletion, and automated snapshots are retained for only 1 day.

## When to Use

- Local development and testing of analytical queries against Redshift
- Validating ETL pipelines and data loading (COPY) workflows before promoting to production
- Prototyping dashboards and BI tool integrations with a live Redshift endpoint

## Key Configuration Choices

- **dc2.large node type** (`nodeType: dc2.large`) -- Dense compute SSD node; low cost, suitable for small datasets up to ~160 GB compressed
- **Single node** (`numberOfNodes: 1`) -- Leader and compute on one node; no inter-node communication overhead
- **Managed password** (`manageMasterPassword: true`) -- AWS Secrets Manager creates and rotates the master password automatically
- **Encryption enabled** (`encrypted: true`) -- Data at rest encrypted with the AWS-managed Redshift service key
- **Skip final snapshot** (`skipFinalSnapshot: true`) -- No snapshot on deletion; appropriate for ephemeral dev clusters
- **1-day snapshot retention** (`automatedSnapshotRetentionPeriod: 1`) -- Minimal retention for dev workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |

## Related Presets

- **02-multi-node-production** -- Use for production workloads requiring multi-node compute, encryption with a customer-managed KMS key, and audit logging
- **03-analytics-workload** -- Use for large-scale analytics with Multi-AZ, concurrency scaling, and Redshift Spectrum IAM roles
