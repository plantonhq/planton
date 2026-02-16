---
title: "Background Worker"
description: "This preset deploys a Fargate-based ECS service for background processing without an ALB. The container has no exposed port -- it pulls work from a queue (SQS, Redis, etc.) or runs scheduled tasks...."
type: "preset"
rank: "03"
presetSlug: "03-background-worker"
componentSlug: "ecs-service"
componentTitle: "ECS Service"
provider: "aws"
icon: "package"
order: 3
---

# Background Worker

This preset deploys a Fargate-based ECS service for background processing without an ALB. The container has no exposed port -- it pulls work from a queue (SQS, Redis, etc.) or runs scheduled tasks. This is the standard pattern for message consumers, event processors, and async job workers.

## When to Use

- Queue consumers that poll SQS, Redis, or Kafka for messages
- Event-driven processors triggered by SNS, EventBridge, or other AWS services
- Batch processing or cron-like workloads running on ECS

## Key Configuration Choices

- **No ALB** -- No `alb` configuration; the service is not exposed to HTTP traffic
- **No container port** -- Worker does not accept inbound connections
- **Single replica** (`replicas: 1`) -- Start with 1; increase or add autoscaling for higher throughput
- **256 CPU / 512 MiB memory** -- Minimal resource allocation for lightweight workers
- **Environment variable for queue URL** -- Pass the work source via `QUEUE_URL`; add more variables as needed

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<ecs-cluster-arn>` | ARN of the ECS cluster | AWS ECS console or `AwsEcsCluster` status outputs |
| `<ecr-repo-uri>` | ECR repository URI | AWS ECR console or `AwsEcrRepo` status outputs |
| `<sqs-queue-url>` | URL of the SQS queue (or other work source) | AWS SQS console |
| `<private-subnet-id>` | Private subnet for the worker task | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing outbound access to queue and dependencies | AWS EC2 console or `AwsSecurityGroup` status outputs |

## Related Presets

- **01-web-service-alb** -- Use instead for HTTP services that need ALB integration
- **02-api-service-autoscaling** -- Use instead for API services with traffic-based scaling
