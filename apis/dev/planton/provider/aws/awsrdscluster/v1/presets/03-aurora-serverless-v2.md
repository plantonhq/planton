# Aurora Serverless v2

This preset creates an Aurora PostgreSQL cluster with Serverless v2 auto-scaling, where compute capacity automatically adjusts between 0.5 and 16 ACUs based on workload demand. The Data API (HTTP endpoint) is enabled for serverless access patterns. Ideal for applications with variable, unpredictable, or spiky traffic.

## When to Use

- Applications with variable or unpredictable traffic (scales to near-zero during idle periods)
- Development and staging environments where cost should track actual usage
- Serverless application architectures using Lambda or Step Functions with the Data API
- Workloads that need Aurora's reliability but don't want to manage instance sizes

## Key Configuration Choices

- **Serverless v2 scaling** (`serverlessV2Scaling`) -- Auto-scales between 0.5 and 16 ACUs; 1 ACU = ~2 GiB RAM
- **Minimum 0.5 ACU** (`minCapacity: 0.5`) -- Near-zero baseline for cost optimization during idle periods
- **Maximum 16 ACU** (`maxCapacity: 16`) -- Handles significant traffic spikes; adjust based on your peak load
- **Data API enabled** (`enableHttpEndpoint: true`) -- HTTP-based SQL access without persistent connections; ideal for Lambda
- **Same production defaults** as other Aurora presets: encrypted storage, deletion protection, managed password

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing database port from application tier | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<database-name>` | Name of the initial database to create | Your application configuration |
| `<final-snapshot-name>` | Identifier for the final snapshot | Your naming convention |

## Related Presets

- **01-aurora-postgresql** -- Use instead for provisioned Aurora with predictable capacity
- **02-aurora-mysql** -- Use instead for MySQL-compatible provisioned workloads
