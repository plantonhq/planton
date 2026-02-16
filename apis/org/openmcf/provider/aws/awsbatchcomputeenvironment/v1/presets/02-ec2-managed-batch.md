# EC2 Managed Batch

This preset creates an EC2-based AWS Batch compute environment with auto-scaling from 0 to 512 vCPUs. Uses `optimal` instance type selection and two priority-separated job queues. Ideal for production batch workloads requiring GPU access, custom AMIs, or sustained throughput.

## When to Use

- Batch jobs that need GPU instances (p3, g4dn) or specific EC2 instance types
- Workloads requiring custom AMIs with pre-installed software
- Production environments with predictable, high-throughput batch processing
- Scenarios where you need fine-grained control over instance types and scaling
- Workloads that benefit from multi-queue priority separation (e.g., critical vs. background jobs)

## Key Configuration Choices

- **EC2** (`type`) — On-demand EC2 instances managed by AWS Batch
- **optimal** (`instanceTypes`) — AWS Batch selects the best-fit instance type for each job's resource requirements
- **BEST_FIT_PROGRESSIVE** (`allocationStrategy`) — Selects the cheapest instance type that fits, progressing to larger types as needed
- **0 min / 512 max vCPUs** — Scales to zero when idle; up to 512 vCPUs under load
- **Update policy** — Waits up to 60 minutes for running jobs before replacing instances during updates
- **Two queues** — `high-priority` (10) for time-sensitive jobs; `low-priority` (1) for background work

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing outbound access for instances | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<ecs-instance-profile-arn>` | IAM instance profile ARN for ECS agent on EC2 instances | AWS IAM console or `AwsIamRole` status outputs |

## Related Presets

- **01-fargate-batch** — Use instead for serverless, zero-management batch processing
- **03-spot-cost-optimized-batch** — Use instead to reduce costs with Spot instances
