---
title: "ECS Task Execution Role"
description: "This preset creates an IAM role that the ECS agent assumes to pull container images from ECR and write logs to CloudWatch on behalf of your tasks. Every ECS Fargate service needs a task execution..."
type: "preset"
rank: "02"
presetSlug: "02-ecs-task-execution"
componentSlug: "iam-role"
componentTitle: "IAM Role"
provider: "aws"
icon: "package"
order: 2
---

# ECS Task Execution Role

This preset creates an IAM role that the ECS agent assumes to pull container images from ECR and write logs to CloudWatch on behalf of your tasks. Every ECS Fargate service needs a task execution role. This is distinct from the task role (which your application code uses to call AWS APIs).

## When to Use

- Any ECS Fargate service that pulls images from ECR
- Services that write container logs to CloudWatch
- Required for ECS services defined via `AwsEcsService`

## Key Configuration Choices

- **ECS tasks service principal** (`ecs-tasks.amazonaws.com`) -- Only ECS can assume this role
- **Task execution policy** (`AmazonECSTaskExecutionRolePolicy`) -- Grants ECR image pull, CloudWatch log writing, and Secrets Manager read for container secrets
- **Not a task role** -- This role is for the ECS agent infrastructure; your application's AWS permissions go in a separate task role

## Placeholders to Replace

This preset has no placeholders. Deploy as-is and reference the role ARN from `AwsEcsService.iam.taskExecutionRoleArn`.

## Related Presets

- **01-lambda-execution** -- Use instead for Lambda function execution roles
- **03-ec2-ssm** -- Use instead for EC2 instances needing SSM access
