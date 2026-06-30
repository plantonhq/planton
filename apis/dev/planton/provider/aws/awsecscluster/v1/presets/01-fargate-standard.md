# Standard Fargate Cluster

This preset creates an ECS cluster using AWS Fargate with CloudWatch Container Insights enabled. Fargate eliminates the need to manage EC2 instances for container workloads -- AWS handles the compute infrastructure. This is the standard starting point for any ECS deployment.

## When to Use

- Running containerized workloads without managing EC2 instances
- Standard production ECS deployments where cost optimization via Spot is not yet needed
- Any ECS cluster that will run Fargate tasks and services

## Key Configuration Choices

- **Fargate only** (`capacityProviders: [FARGATE]`) -- All tasks run on-demand Fargate; predictable pricing with no Spot interruptions
- **Container Insights enabled** (`enableContainerInsights: true`) -- CloudWatch metrics for CPU, memory, network, and storage at the task and service level

## Placeholders to Replace

This preset has no placeholders. Deploy as-is and then create `AwsEcsService` resources targeting this cluster.

## Related Presets

- **02-fargate-cost-optimized** -- Use instead when you want to reduce costs by running a portion of tasks on Fargate Spot
