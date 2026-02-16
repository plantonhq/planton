# High Availability Neptune Cluster

This preset creates a highly available Neptune cluster with 2 instances (1 primary writer + 1 read replica) across Availability Zones. IAM database authentication is enabled, deletion protection is on, and audit/slowquery logs are exported to CloudWatch. Neptune is a fully managed graph database supporting Gremlin and SPARQL for recommendation engines, fraud detection, and knowledge graphs.

## When to Use

- Production graph workloads requiring high availability and automatic failover
- Applications using Gremlin or SPARQL that need read scaling across replicas
- Workloads requiring IAM database authentication (no master password)
- Compliance-sensitive environments needing audit logs and deletion protection

## Key Configuration Choices

- **2 instances** (`instanceCount: 2`) — Primary writer + 1 read replica across AZs for automatic failover
- **db.r6g.large** (`instanceClass`) — 2 vCPUs, 16 GiB RAM; memory-optimized for graph workloads
- **IAM database authentication** (`iamDatabaseAuthenticationEnabled: true`) — Secure access via IAM; no master password
- **Encrypted storage** (`storageEncrypted: true`) — Data at rest encryption with AWS-managed key
- **14-day backup retention** — Extended retention for point-in-time recovery
- **Deletion protection** (`deletionProtection: true`) — Prevents accidental cluster deletion
- **Final snapshot on deletion** (`skipFinalSnapshot: false`) — Creates recovery snapshot before deletion
- **CloudWatch logs** (`enabledCloudwatchLogsExports: [audit, slowquery]`) — Audit for compliance; slowquery for performance tuning

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing Neptune port (8182) from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<vpc-id>` | VPC ID where the cluster will be deployed | AWS VPC console or `AwsVpc` status outputs |
| `<final-snapshot-name>` | Identifier for the final snapshot (e.g., `neptune-ha-final-2026-02-16`) | Your naming convention |

## Related Presets

- **01-graph-database** — Use instead for single-instance dev/test environments with minimal cost
- **03-serverless-v2** — Use instead for variable traffic patterns with auto-scaling compute
