# Fargate Batch (Serverless)

This preset creates a serverless AWS Batch compute environment using Fargate. AWS manages all infrastructure — no EC2 instances to configure, patch, or scale. Ideal for teams that want to run batch jobs without managing compute infrastructure.

## When to Use

- Batch jobs that run containers without needing GPU or custom AMIs
- Teams that want zero infrastructure management overhead
- Variable workloads where you pay only for the vCPU and memory used per job
- Environments where patching and AMI management is undesirable
- Quick prototyping of batch processing pipelines

## Key Configuration Choices

- **FARGATE** (`type`) — Serverless containers; AWS provisions and manages compute resources
- **256 max vCPUs** (`maxVcpus`) — Upper limit on concurrent Fargate vCPUs; adjust based on workload parallelism
- **Single queue** (`jobQueues`) — One queue with priority 1; add more queues for workload isolation
- **Multi-AZ subnets** — Two subnets for high availability across Availability Zones

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing outbound access for containers | AWS EC2 console or `AwsSecurityGroup` status outputs |

## Related Presets

- **02-ec2-managed-batch** — Use instead when you need GPU instances, custom AMIs, or sustained compute capacity
- **03-spot-cost-optimized-batch** — Use instead for cost-sensitive workloads tolerant of interruptions
