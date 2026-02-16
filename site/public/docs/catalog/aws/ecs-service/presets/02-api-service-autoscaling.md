---
title: "API Service with Autoscaling"
description: "This preset deploys a Fargate-based ECS service with hostname-based ALB routing and CPU-based autoscaling. It starts with 2 replicas and scales to 10 based on a 75% CPU target. This is the standard..."
type: "preset"
rank: "02"
presetSlug: "02-api-service-autoscaling"
componentSlug: "ecs-service"
componentTitle: "ECS Service"
provider: "aws"
icon: "package"
order: 2
---

# API Service with Autoscaling

This preset deploys a Fargate-based ECS service with hostname-based ALB routing and CPU-based autoscaling. It starts with 2 replicas and scales to 10 based on a 75% CPU target. This is the standard pattern for production API services that need to handle variable traffic loads.

## When to Use

- API services with their own hostname (e.g., `api.example.com`) that need to scale with traffic
- Production services where traffic is variable or seasonal
- Microservice architectures where each service has a dedicated hostname on a shared ALB

## Key Configuration Choices

- **Hostname-based routing** (`routingType: hostname`) -- Routes traffic based on the `Host` header; multiple services can share a single ALB
- **CPU-based autoscaling** (`targetCpuPercent: 75`) -- Scales out when average CPU exceeds 75%, scales in when it drops below
- **2-10 tasks** (`minTasks: 2`, `maxTasks: 10`) -- Always at least 2 for HA; up to 10 for peak traffic
- **1024 CPU / 2048 MiB memory** -- 1 full vCPU with 2 GiB RAM; suitable for API workloads
- **Health check on /health** -- ALB verifies each task is responding before routing traffic

## Placeholders to Replace

| Placeholder | Description | Where to Find |
| --- | --- | --- |
| `<ecs-cluster-arn>` | ARN of the ECS cluster | AWS ECS console or `AwsEcsCluster` status outputs |
| `<ecr-repo-uri>` | ECR repository URI | AWS ECR console or `AwsEcrRepo` status outputs |
| `<private-subnet-id-az1>` | Private subnet in the first Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<private-subnet-id-az2>` | Private subnet in the second Availability Zone | AWS VPC console or `AwsVpc` status outputs |
| `<security-group-id>` | Security group allowing traffic from the ALB | AWS EC2 console or `AwsSecurityGroup` status outputs |
| `<alb-arn>` | ARN of the Application Load Balancer | AWS EC2 console or `AwsAlb` status outputs |
| `<api-hostname>` | Hostname for this service (e.g., `api.example.com`) | Your DNS configuration |

## Related Presets

- **01-web-service-alb** -- Use instead for path-based routing without autoscaling
- **03-background-worker** -- Use instead for services that process messages without ALB
