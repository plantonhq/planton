# Single-Node Development Cluster

This preset creates a single-node Redshift cluster on the ra3.large node type for development and testing. The single-node topology combines the leader and compute roles on one node, keeping costs low while providing a functional SQL analytics environment. No final snapshot is taken on deletion, and automated snapshots are retained for only 1 day.

## When to Use

- Local development and testing of analytical queries against Redshift
- Validating ETL pipelines and data loading (COPY) workflows before promoting to production
- Prototyping dashboards and BI tool integrations with a live Redshift endpoint

## Key Configuration Choices

- **ra3.large node type** (`nodeType: ra3.large`) -- The smallest orderable class; managed storage tiers between local SSD and S3, so small dev datasets stay cheap (the legacy dc2 dense-compute family is no longer creatable for new clusters)
- **Single node** (`numberOfNodes: 1`) -- Leader and compute on one node; no inter-node communication overhead
- **Managed password** (`manageMasterPassword: true`) -- AWS Secrets Manager creates and rotates the admin password automatically; no secret in the manifest or IaC state
- **Encryption by default** -- At-rest encryption is on unless explicitly disabled (the AWS default this kind preserves), using the AWS-managed Redshift service key
- **Default database** -- No `databaseName` pin; AWS creates the default `dev` database
- **Skip final snapshot** (`skipFinalSnapshot: true`) -- No snapshot on deletion; appropriate for ephemeral dev clusters
- **1-day snapshot retention** (`automatedSnapshotRetentionPeriod: 1`) -- Minimal retention for dev workloads

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsSubnet` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsSubnet` status outputs |

## Related Presets

- **02-multi-node-production** -- Use for production workloads requiring multi-node compute, encryption with a customer-managed KMS key, audit logging, and cross-region snapshot copy
- **03-analytics-workload** -- Use for large-scale analytics with Multi-AZ, concurrency scaling, and Redshift Spectrum IAM roles
