# Spot Cost-Optimized Batch

This preset creates a high-capacity Spot-based AWS Batch compute environment optimized for cost. Uses capacity-optimized allocation across multiple instance families with a fair-share scheduling policy to divide compute across teams. Ideal for large-scale, cost-sensitive batch processing that tolerates occasional interruptions.

## When to Use

- Large-scale data processing, ETL pipelines, or ML training jobs where cost is the primary concern
- Workloads that are interruption-tolerant (can checkpoint and restart)
- Multi-team environments where compute capacity must be fairly shared
- Scenarios where you want to maximize throughput per dollar spent
- Jobs that run for minutes to hours (not seconds) and benefit from Spot pricing

## Key Configuration Choices

- **SPOT** (`type`) — EC2 Spot instances at up to 90% discount over On-Demand
- **SPOT_CAPACITY_OPTIMIZED** (`allocationStrategy`) — Selects from the deepest Spot capacity pools to minimize interruptions
- **60% bid** (`bidPercentage`) — Pay up to 60% of On-Demand price; jobs are reclaimed if Spot price exceeds this
- **1024 max vCPUs** — High concurrency ceiling for parallel processing
- **Multi-family instances** — Five instance types across m5/c5/r5 families for capacity diversity
- **Three AZ subnets** — Maximizes Spot pool diversity across Availability Zones
- **Fair-share scheduling** — Team-data gets 2x the share of team-ml (1.0 vs 0.5 weight)
- **RUNNABLE timeout** — Jobs stuck waiting for resources for 2 hours are automatically cancelled
- **1-hour share decay** — Recent usage decays quickly, allowing faster recovery after bursts

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az3>` | Private subnet in the third Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing outbound access for instances | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<ecs-instance-profile-arn>` | IAM instance profile ARN for ECS agent on EC2 instances | AWS IAM console or `AwsIamRole` status outputs |
| `<spot-fleet-role-arn>` | IAM role ARN for EC2 Spot Fleet requests | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-fargate-batch** — Use instead for serverless, zero-management batch processing
- **02-ec2-managed-batch** — Use instead for On-Demand instances with guaranteed availability
